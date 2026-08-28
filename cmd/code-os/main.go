package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/melvynx/code-os/internal/cloudflare"
	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/model"
	"github.com/melvynx/code-os/internal/server"
	"github.com/melvynx/code-os/internal/store"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "code-os:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) == 0 {
		printUsage()
		return nil
	}
	switch arguments[0] {
	case "setup":
		return setup(arguments[1:])
	case "dashboard", "serve":
		return dashboard(arguments[1:])
	case "scan":
		return scan(arguments[1:])
	case "status":
		return status(arguments[1:])
	case "doctor":
		return doctor(arguments[1:])
	case "cloudflare":
		return cloudflareCommand(arguments[1:])
	case "service":
		return serviceCommand(arguments[1:])
	case "skills-sync":
		return skillsSync(arguments[1:])
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q; run code-os help", arguments[0])
	}
}

func printUsage() {
	fmt.Print(`Code OS — your dev stack, everywhere

Usage:
  code-os setup       Configure this machine
  code-os dashboard   Run the loopback command center
  code-os scan        Print a fresh JSON snapshot
  code-os status      Print a compact summary
  code-os doctor      Check the environment
  code-os cloudflare  Print the tunnel ingress configuration
  code-os service     Install or inspect the user service
  code-os skills-sync Synchronize the configured private skills repository
  code-os version     Print the version
`)
}

func setup(arguments []string) error {
	defaults, err := config.Defaults()
	if err != nil {
		return err
	}
	defaultConfigPath, err := config.DefaultPath()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	configPath := flags.String("config", defaultConfigPath, "configuration path")
	name := flags.String("name", defaults.EnvironmentName, "environment name")
	typeName := flags.String("type", defaults.EnvironmentType, "local or remote")
	address := flags.String("address", defaults.Address, "dashboard listen address")
	projectsRoot := flags.String("projects-root", defaults.ProjectsRoots[0], "projects root")
	screenshotsRoot := flags.String("screenshots-root", defaults.ScreenshotsRoot, "private screenshot root")
	dashboardHost := flags.String("dashboard-host", "", "Cloudflare dashboard hostname")
	portHost := flags.String("public-port-host", "", "public port hostname template, for example port{port}.example.com")
	filesRoot := flags.String("files-root", defaults.FilesRoot, "private files root")
	skillsRepository := flags.String("skills-repository", "", "Git repository used to synchronize agent skills")
	skillsDirectory := flags.String("skills-directory", "", "local agent skills checkout")
	skillsBranch := flags.String("skills-branch", "main", "agent skills Git branch")
	portlyBinary := flags.String("portly", defaults.PortlyBinary, "Portly binary")
	tunnelID := flags.String("cloudflare-tunnel-id", "", "Cloudflare tunnel ID")
	tunnelMode := flags.String("cloudflare-tunnel-mode", "shared", "shared or dedicated tunnel")
	accountID := flags.String("cloudflare-account-id", "", "Cloudflare account ID")
	zoneID := flags.String("cloudflare-zone-id", "", "Cloudflare zone ID")
	tokenFile := flags.String("cloudflare-token-file", "", "path to a scoped Cloudflare token")
	authUsername := flags.String("dashboard-username", "", "dashboard sign-in username")
	authPasswordFile := flags.String("dashboard-password-file", "", "path to dashboard password")
	authBypassKeyFile := flags.String("dashboard-bypass-key-file", "", "path to the media-only bypass key")
	authSessionKeyFile := flags.String("dashboard-session-key-file", "", "path to the session signing key")
	nonInteractive := flags.Bool("non-interactive", false, "do not prompt for missing values")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if !*nonInteractive && isTerminal(os.Stdin) {
		reader := bufio.NewReader(os.Stdin)
		*name = prompt(reader, "Environment name", *name)
		*typeName = prompt(reader, "Environment type (local/remote)", *typeName)
		*projectsRoot = prompt(reader, "Projects directory", *projectsRoot)
		*screenshotsRoot = prompt(reader, "Private screenshots directory", *screenshotsRoot)
		*dashboardHost = prompt(reader, "Dashboard hostname", *dashboardHost)
	}
	if strings.TrimSpace(*dashboardHost) == "" {
		return errors.New("dashboard hostname is required; pass --dashboard-host")
	}
	cfg := defaults
	cfg.EnvironmentName = strings.TrimSpace(*name)
	cfg.EnvironmentType = strings.TrimSpace(*typeName)
	cfg.Address = strings.TrimSpace(*address)
	cfg.ProjectsRoots = []string{expandPath(*projectsRoot)}
	cfg.ScreenshotsRoot = expandPath(*screenshotsRoot)
	cfg.FilesRoot = expandPath(*filesRoot)
	cfg.PortlyBinary = strings.TrimSpace(*portlyBinary)
	cfg.PublicPortHost = strings.TrimSpace(*portHost)
	cfg.Cloudflare = config.Cloudflare{
		DashboardHost: strings.TrimSpace(*dashboardHost), TunnelMode: strings.TrimSpace(*tunnelMode), TunnelID: strings.TrimSpace(*tunnelID),
		AccountID: strings.TrimSpace(*accountID), ZoneID: strings.TrimSpace(*zoneID),
		TokenFile: expandOptionalPath(*tokenFile), RequireAccess: true,
	}
	cfg.Auth = config.Auth{
		Username: strings.TrimSpace(*authUsername), PasswordFile: expandOptionalPath(*authPasswordFile),
		BypassKeyFile: expandOptionalPath(*authBypassKeyFile), SessionKeyFile: expandOptionalPath(*authSessionKeyFile),
	}
	if cfg.Auth.PasswordFile != "" && cfg.Auth.SessionKeyFile == "" {
		cfg.Auth.SessionKeyFile = filepath.Join(filepath.Dir(cfg.Auth.PasswordFile), "session-key")
	}
	cfg.Skills = config.Skills{
		Repository: strings.TrimSpace(*skillsRepository), Directory: expandOptionalPath(*skillsDirectory), Branch: strings.TrimSpace(*skillsBranch),
	}
	if cfg.Auth.PasswordFile != "" {
		if err := ensurePasswordFile(cfg.Auth.PasswordFile); err != nil {
			return err
		}
	}
	if cfg.Auth.BypassKeyFile != "" {
		if err := ensurePasswordFile(cfg.Auth.BypassKeyFile); err != nil {
			return err
		}
	}
	if cfg.Auth.SessionKeyFile != "" {
		if err := ensurePasswordFile(cfg.Auth.SessionKeyFile); err != nil {
			return err
		}
	}
	if err := config.Save(*configPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Code OS configured at %s\n", *configPath)
	fmt.Printf("Dashboard origin: http://%s\n", cfg.Address)
	if cfg.Auth.PasswordFile != "" {
		fmt.Printf("Public hostname: https://%s (origin authentication enabled)\n", cfg.Cloudflare.DashboardHost)
		fmt.Printf("Dashboard credentials: user %s, password in %s\n", cfg.Auth.Username, cfg.Auth.PasswordFile)
		if cfg.Auth.BypassKeyFile != "" {
			fmt.Printf("Media bypass key: %s\n", cfg.Auth.BypassKeyFile)
		}
	} else {
		fmt.Printf("Public hostname: https://%s (Cloudflare Access required)\n", cfg.Cloudflare.DashboardHost)
	}
	fmt.Println("Next: code-os doctor")
	return nil
}

func dashboard(arguments []string) error {
	flags := flag.NewFlagSet("dashboard", flag.ContinueOnError)
	configPath := addConfigFlag(flags)
	refreshInterval := flags.Duration("refresh", 20*time.Second, "snapshot refresh interval")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if !isLoopbackAddress(cfg.Address) {
		return fmt.Errorf("refusing non-loopback dashboard address %q", cfg.Address)
	}
	database, err := store.Open(server.DatabasePath(cfg))
	if err != nil {
		return err
	}
	defer database.Close()
	service := server.NewService(cfg, database)
	rootContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	refreshContext, cancel := context.WithTimeout(rootContext, 30*time.Second)
	service.Refresh(refreshContext)
	cancel()
	go service.RunRefreshLoop(rootContext, *refreshInterval)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	httpServer := server.HTTPServer{Config: cfg, ConfigPath: *configPath, Service: service, Logger: logger}
	handler, err := httpServer.Handler()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.Address, err)
	}
	runtimeServer := &http.Server{
		Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}
	result := make(chan error, 1)
	go func() {
		logger.Info("Code OS dashboard listening", "address", cfg.Address, "hostname", cfg.Cloudflare.DashboardHost)
		result <- runtimeServer.Serve(listener)
	}()
	select {
	case <-rootContext.Done():
		shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		return runtimeServer.Shutdown(shutdownContext)
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func scan(arguments []string) error {
	cfg, service, closeStore, err := commandService(arguments, "scan")
	if err != nil {
		return err
	}
	defer closeStore()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	snapshot := service.Refresh(ctx)
	_ = cfg
	return printJSON(snapshot)
}

func status(arguments []string) error {
	cfg, service, closeStore, err := commandService(arguments, "status")
	if err != nil {
		return err
	}
	defer closeStore()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	snapshot := service.Refresh(ctx)
	overview := modelSummarize(snapshot)
	fmt.Printf("Environment   %s (%s)\n", cfg.EnvironmentName, cfg.EnvironmentType)
	fmt.Printf("Dashboard     https://%s\n", cfg.Cloudflare.DashboardHost)
	fmt.Printf("Projects      %d (%d modified)\n", overview.ProjectCount, overview.ModifiedProjects)
	fmt.Printf("Worktrees     %d (%d modified)\n", overview.WorktreeCount, overview.ModifiedWorktrees)
	fmt.Printf("Applications  %d running (%d unhealthy)\n", overview.RunningApps, overview.UnhealthyApps)
	fmt.Printf("Screenshots   %d\n", overview.ScreenshotCount)
	fmt.Printf("Warnings      %d\n", overview.WarningCount)
	return nil
}

func doctor(arguments []string) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	configPath := addConfigFlag(flags)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	type check struct {
		label, detail string
		ok, required  bool
	}
	checks := []check{
		{"Configuration", *configPath, true, true},
		{"Loopback dashboard", cfg.Address, isLoopbackAddress(cfg.Address), true},
		{"Dashboard hostname", cfg.Cloudflare.DashboardHost, cfg.Cloudflare.DashboardHost != "", true},
		{"Cloudflare Access policy", "must be enabled before publishing", cfg.Cloudflare.RequireAccess, true},
	}
	if cfg.Auth.PasswordFile != "" {
		checks = append(checks, check{"Dashboard authentication", cfg.Auth.Username, isPrivateRegularFile(cfg.Auth.PasswordFile), true})
	}
	if cfg.Auth.BypassKeyFile != "" {
		checks = append(checks, check{"Media bypass key", cfg.Auth.BypassKeyFile, isPrivateRegularFile(cfg.Auth.BypassKeyFile), true})
	}
	if cfg.Auth.SessionKeyFile != "" {
		checks = append(checks, check{"Session signing key", cfg.Auth.SessionKeyFile, isPrivateRegularFile(cfg.Auth.SessionKeyFile), true})
	}
	for _, root := range cfg.ProjectsRoots {
		checks = append(checks, check{"Projects root", root, isDirectory(root), true})
	}
	checks = append(checks,
		check{"Screenshot root", cfg.ScreenshotsRoot, isDirectory(cfg.ScreenshotsRoot), false},
		check{"Portly", cfg.PortlyBinary, commandExists(cfg.PortlyBinary), true},
		check{"Git", "git", commandExists("git"), true},
	)
	if cfg.Cloudflare.TokenFile != "" {
		checks = append(checks, check{"Cloudflare token file", cfg.Cloudflare.TokenFile, isRegularFile(cfg.Cloudflare.TokenFile), false})
	}
	failed := false
	for _, item := range checks {
		mark := "✓"
		if !item.ok {
			mark = "!"
			if item.required {
				mark = "✗"
				failed = true
			}
		}
		fmt.Printf("%s %-24s %s\n", mark, item.label, item.detail)
	}
	if failed {
		return errors.New("required checks failed")
	}
	return nil
}

func cloudflareCommand(arguments []string) error {
	flags := flag.NewFlagSet("cloudflare", flag.ContinueOnError)
	configPath := addConfigFlag(flags)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	payload, err := cloudflare.Render(cfg)
	if err != nil {
		return err
	}
	fmt.Print(payload)
	fmt.Fprintf(os.Stderr, "Protect https://%s with Cloudflare Access before publishing.\n", cfg.Cloudflare.DashboardHost)
	return nil
}

func serviceCommand(arguments []string) error {
	if runtime.GOOS != "linux" {
		return errors.New("service installation currently supports Linux systemd user services")
	}
	if len(arguments) == 0 {
		arguments = []string{"print"}
	}
	action := arguments[0]
	flags := flag.NewFlagSet("service "+action, flag.ContinueOnError)
	configPath := addConfigFlag(flags)
	if err := flags.Parse(arguments[1:]); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	unit := renderUnit(executable, *configPath, cfg.DataDir, cfg.Cloudflare.TokenFile)
	if action == "print" {
		fmt.Print(unit)
		return nil
	}
	if action != "install" {
		return fmt.Errorf("unknown service action %q; use print or install", action)
	}
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	unitPath := filepath.Join(userConfig, "systemd", "user", "code-os.service")
	if err := os.MkdirAll(filepath.Dir(unitPath), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(unitPath, []byte(unit), 0o600); err != nil {
		return fmt.Errorf("write user service: %w", err)
	}
	if cfg.Skills.Repository != "" {
		syncServicePath := filepath.Join(userConfig, "systemd", "user", "code-os-skills-sync.service")
		syncTimerPath := filepath.Join(userConfig, "systemd", "user", "code-os-skills-sync.timer")
		if err := os.WriteFile(syncServicePath, []byte(renderSkillsSyncUnit(executable, *configPath)), 0o600); err != nil {
			return fmt.Errorf("write skills sync service: %w", err)
		}
		if err := os.WriteFile(syncTimerPath, []byte(renderSkillsSyncTimer()), 0o600); err != nil {
			return fmt.Errorf("write skills sync timer: %w", err)
		}
		fmt.Printf("Installed %s and %s\n", syncServicePath, syncTimerPath)
	}
	fmt.Printf("Installed %s\n", unitPath)
	fmt.Println("Enable with: systemctl --user daemon-reload && systemctl --user enable --now code-os.service")
	return nil
}

func skillsSync(arguments []string) error {
	flags := flag.NewFlagSet("skills-sync", flag.ContinueOnError)
	configPath := addConfigFlag(flags)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if cfg.Skills.Repository == "" || cfg.Skills.Directory == "" || cfg.Skills.Branch == "" {
		return errors.New("skills repository, directory, and branch must be configured")
	}
	if _, err := os.Stat(cfg.Skills.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(cfg.Skills.Directory), 0o700); err != nil {
			return fmt.Errorf("create skills parent directory: %w", err)
		}
		return runCommand("git", "clone", "--branch", cfg.Skills.Branch, "--single-branch", cfg.Skills.Repository, cfg.Skills.Directory)
	}
	if _, err := gitOutput(cfg.Skills.Directory, "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("skills directory is not a Git repository: %w", err)
	}
	remote, err := gitOutput(cfg.Skills.Directory, "remote", "get-url", "origin")
	if err != nil {
		return errors.New("skills repository has no origin remote")
	}
	if strings.TrimSpace(remote) != cfg.Skills.Repository {
		return fmt.Errorf("skills origin %q does not match configured repository", strings.TrimSpace(remote))
	}
	branch, err := gitOutput(cfg.Skills.Directory, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || strings.TrimSpace(branch) != cfg.Skills.Branch {
		return fmt.Errorf("skills checkout must be on branch %q", cfg.Skills.Branch)
	}
	gitDirectory, err := gitOutput(cfg.Skills.Directory, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	lockDirectory := filepath.Join(strings.TrimSpace(gitDirectory), "code-os-sync.lock")
	if err := os.Mkdir(lockDirectory, 0o700); err != nil {
		if os.IsExist(err) {
			fmt.Println("Code OS skills sync: another sync is already running")
			return nil
		}
		return fmt.Errorf("create skills sync lock: %w", err)
	}
	defer os.Remove(lockDirectory)
	for _, marker := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD"} {
		if _, err := os.Stat(filepath.Join(strings.TrimSpace(gitDirectory), marker)); err == nil {
			return errors.New("resolve the existing Git operation before synchronizing skills")
		}
	}
	if err := runGit(cfg.Skills.Directory, "diff", "--quiet", "--diff-filter=U"); err != nil {
		return errors.New("resolve skills repository conflicts before synchronizing")
	}
	if err := runGit(cfg.Skills.Directory, "add", "-A"); err != nil {
		return err
	}
	if err := runGit(cfg.Skills.Directory, "diff", "--cached", "--quiet"); err != nil {
		hostname, _ := os.Hostname()
		message := fmt.Sprintf("sync(%s): %s", hostname, time.Now().UTC().Format(time.RFC3339))
		if err := runGit(cfg.Skills.Directory, "commit", "-m", message); err != nil {
			return err
		}
	}
	if err := runGit(cfg.Skills.Directory, "pull", "--rebase", "--autostash", "origin", cfg.Skills.Branch); err != nil {
		return err
	}
	if err := runGit(cfg.Skills.Directory, "push", "origin", cfg.Skills.Branch); err != nil {
		return err
	}
	fmt.Println("Code OS skills sync: repository is up to date")
	return nil
}

func gitOutput(directory string, arguments ...string) (string, error) {
	arguments = append([]string{"-C", directory}, arguments...)
	command := exec.Command("git", arguments...)
	output, err := command.Output()
	return string(output), err
}

func runGit(directory string, arguments ...string) error {
	return runCommand("git", append([]string{"-C", directory}, arguments...)...)
}

func runCommand(name string, arguments ...string) error {
	command := exec.Command(name, arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	return nil
}

func commandService(arguments []string, name string) (config.Config, *server.Service, func(), error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	configPath := addConfigFlag(flags)
	if err := flags.Parse(arguments); err != nil {
		return config.Config{}, nil, func() {}, err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return config.Config{}, nil, func() {}, err
	}
	database, err := store.Open(server.DatabasePath(cfg))
	if err != nil {
		return config.Config{}, nil, func() {}, err
	}
	return cfg, server.NewService(cfg, database), func() { database.Close() }, nil
}

func addConfigFlag(flags *flag.FlagSet) *string {
	path, _ := config.DefaultPath()
	return flags.String("config", path, "configuration path")
}

func prompt(reader *bufio.Reader, label, fallback string) string {
	fmt.Printf("%s [%s]: ", label, fallback)
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func expandPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if path == "~" {
			return home
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	absolute, err := filepath.Abs(path)
	if err == nil {
		return absolute
	}
	return path
}

func expandOptionalPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return expandPath(path)
}

func ensurePasswordFile(path string) error {
	if isRegularFile(path) {
		if err := os.Chmod(path, 0o600); err != nil {
			return fmt.Errorf("secure secret file: %w", err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create password directory: %w", err)
	}
	secret := make([]byte, 24)
	if _, err := rand.Read(secret); err != nil {
		return fmt.Errorf("generate dashboard password: %w", err)
	}
	value := base64.RawURLEncoding.EncodeToString(secret) + "\n"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create dashboard password file: %w", err)
	}
	if _, err := file.WriteString(value); err != nil {
		file.Close()
		return fmt.Errorf("write dashboard password: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close dashboard password file: %w", err)
	}
	return nil
}

func isLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func isPrivateRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm() == 0o600
}

func commandExists(command string) bool {
	if strings.ContainsRune(command, filepath.Separator) {
		return isRegularFile(command)
	}
	_, err := exec.LookPath(command)
	return err == nil
}

func renderUnit(executable, configPath, dataDir, tokenFile string) string {
	writablePaths := []string{dataDir, filepath.Dir(configPath)}
	if tokenFile != "" {
		writablePaths = append(writablePaths, filepath.Dir(tokenFile))
	}
	quotedWritablePaths := make([]string, 0, len(writablePaths))
	for _, path := range writablePaths {
		quotedWritablePaths = append(quotedWritablePaths, strconv.Quote(path))
	}
	return fmt.Sprintf(`[Unit]
Description=Code OS Command Center
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s dashboard --config %s
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=%s

[Install]
WantedBy=default.target
`, strconv.Quote(executable), strconv.Quote(configPath), strings.Join(quotedWritablePaths, " "))
}

func renderSkillsSyncUnit(executable, configPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Synchronize Code OS agent skills
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=%s skills-sync --config %s
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=default.target
`, strconv.Quote(executable), strconv.Quote(configPath))
}

func renderSkillsSyncTimer() string {
	return `[Unit]
Description=Run Code OS skills synchronization

[Timer]
OnBootSec=45s
OnUnitActiveSec=2min
Persistent=true

[Install]
WantedBy=timers.target
`
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

// Kept local to avoid coupling command output to a transport package.
func modelSummarize(snapshot model.Snapshot) model.Overview {
	return model.Summarize(snapshot)
}
