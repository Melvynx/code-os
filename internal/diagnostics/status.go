package diagnostics

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/screenshots"
	"github.com/melvynx/code-os/internal/skills"
)

type Severity string

const (
	Pass Severity = "pass"
	Warn Severity = "warn"
	Fail Severity = "fail"
)

type Check struct {
	ID     string   `json:"id"`
	Label  string   `json:"label"`
	Detail string   `json:"detail"`
	Status Severity `json:"status"`
	Group  string   `json:"group"`
}

type ImageSample struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	URL       string    `json:"url,omitempty"`
	Kind      string    `json:"kind"`
	Size      int64     `json:"size"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	Decodable bool      `json:"decodable"`
	CreatedAt time.Time `json:"createdAt"`
	Issue     string    `json:"issue,omitempty"`
}

type ImageReport struct {
	ScreenshotsRoot       string        `json:"screenshotsRoot"`
	FilesRoot             string        `json:"filesRoot"`
	SharedRoots           bool          `json:"sharedRoots"`
	ScreenshotsRootExists bool          `json:"screenshotsRootExists"`
	FilesRootExists       bool          `json:"filesRootExists"`
	IndexedCount          int           `json:"indexedCount"`
	FilesImageCount       int           `json:"filesImageCount"`
	DecodableCount        int           `json:"decodableCount"`
	EmptyCount            int           `json:"emptyCount"`
	UndecodableCount      int           `json:"undecodableCount"`
	BypassConfigured      bool          `json:"bypassConfigured"`
	BypassPrivate         bool          `json:"bypassPrivate"`
	Recent                []ImageSample `json:"recent"`
}

type Report struct {
	GeneratedAt time.Time     `json:"generatedAt"`
	Healthy     bool          `json:"healthy"`
	Failed      int           `json:"failed"`
	Warnings    int           `json:"warnings"`
	Checks      []Check       `json:"checks"`
	Images      ImageReport   `json:"images"`
	Skills      skills.Status `json:"skills"`
}

type Options struct {
	ConfigPath string
	Skills     skills.Inspector
}

func Inspect(cfg config.Config, options Options) Report {
	images := inspectImages(cfg)
	report := Report{
		GeneratedAt: time.Now().UTC(),
		Images:      images,
		Skills:      options.Skills.Inspect(cfg.Skills),
	}
	if options.ConfigPath != "" {
		report.Checks = append(report.Checks, check("configuration", "environment", "Configuration", options.ConfigPath, true, true))
	}
	report.Checks = append(report.Checks,
		check("loopback", "environment", "Loopback dashboard", cfg.Address, isLoopbackAddress(cfg.Address), true),
		check("dashboard-host", "environment", "Dashboard hostname", empty(cfg.Cloudflare.DashboardHost, "missing"), cfg.Cloudflare.DashboardHost != "", true),
		check("access-policy", "environment", "Cloudflare Access policy", "must be enabled before publishing", cfg.Cloudflare.RequireAccess, true),
	)
	if cfg.Auth.PasswordFile != "" {
		report.Checks = append(report.Checks, check("dashboard-auth", "auth", "Dashboard authentication", cfg.Auth.Username, isPrivateRegularFile(cfg.Auth.PasswordFile), true))
	}
	if cfg.Auth.BypassKeyFile != "" {
		report.Checks = append(report.Checks, check("media-bypass", "images", "Media bypass key", cfg.Auth.BypassKeyFile, isPrivateRegularFile(cfg.Auth.BypassKeyFile), true))
	} else if cfg.Cloudflare.DashboardHost != "" {
		report.Checks = append(report.Checks, check("media-bypass", "images", "Media bypass key", "missing", false, true))
	}
	if cfg.Auth.SessionKeyFile != "" {
		report.Checks = append(report.Checks, check("session-key", "auth", "Session signing key", cfg.Auth.SessionKeyFile, isPrivateRegularFile(cfg.Auth.SessionKeyFile), true))
	}
	if cfg.Auth.TrustedIPsFile != "" {
		report.Checks = append(report.Checks, check("trusted-ips", "auth", "Trusted IP storage", cfg.Auth.TrustedIPsFile, isPrivateRegularFile(cfg.Auth.TrustedIPsFile), true))
	}
	for _, root := range cfg.ProjectsRoots {
		report.Checks = append(report.Checks, check("projects-root", "environment", "Projects root", root, isDirectory(root), true))
	}
	report.Checks = append(report.Checks,
		check("screenshots-root", "images", "Screenshot root", cfg.ScreenshotsRoot, images.ScreenshotsRootExists, false),
		check("files-root", "images", "Private files root", cfg.FilesRoot, images.FilesRootExists, true),
		check("indexed-images", "images", "Indexed screenshots", fmt.Sprintf("%d indexed", images.IndexedCount), images.IndexedCount > 0, false),
		check("decodable-images", "images", "Decodable images", imageDecodeDetail(images), images.EmptyCount == 0 && images.UndecodableCount == 0, imageCount(images) > 0 && images.DecodableCount == 0),
		check("portly", "environment", "Portly", cfg.PortlyBinary, commandExists(cfg.PortlyBinary), true),
		check("git", "environment", "Git", "git", commandExists("git"), true),
	)
	if runtime.GOOS == "linux" {
		report.Checks = append(report.Checks,
			check("user-service", "environment", "Code OS user service", "enabled", commandSucceeds("systemctl", "--user", "is-enabled", "code-os.service"), true),
			check("runtime", "environment", "Code OS runtime", "active", commandSucceeds("systemctl", "--user", "is-active", "code-os.service"), true),
			check("linger", "environment", "Boot without login", "systemd linger enabled", lingerEnabled(), true),
		)
	}
	if cfg.Cloudflare.TokenFile != "" {
		report.Checks = append(report.Checks, check("cloudflare-token", "environment", "Cloudflare token file", cfg.Cloudflare.TokenFile, isRegularFile(cfg.Cloudflare.TokenFile), false))
	}
	if report.Skills.Configured {
		report.Checks = append(report.Checks, Check{
			ID: "skills-sync", Group: "skills", Label: "Skills synchronization",
			Detail: report.Skills.Message, Status: skillsCheckStatus(report.Skills),
		})
	}
	for _, item := range report.Checks {
		switch item.Status {
		case Fail:
			report.Failed++
		case Warn:
			report.Warnings++
		}
	}
	report.Healthy = report.Failed == 0
	return report
}

func inspectImages(cfg config.Config) ImageReport {
	report := ImageReport{
		ScreenshotsRoot:       cfg.ScreenshotsRoot,
		FilesRoot:             cfg.FilesRoot,
		SharedRoots:           samePath(cfg.ScreenshotsRoot, cfg.FilesRoot),
		ScreenshotsRootExists: isDirectory(cfg.ScreenshotsRoot),
		FilesRootExists:       isDirectory(cfg.FilesRoot),
		BypassConfigured:      cfg.Auth.BypassKeyFile != "",
		BypassPrivate:         isPrivateRegularFile(cfg.Auth.BypassKeyFile),
	}
	if indexed, err := (screenshots.Indexer{Root: cfg.ScreenshotsRoot}).Scan(); err == nil {
		report.IndexedCount = len(indexed)
		for _, image := range indexed {
			sample := inspectFile(image.Path, "screenshot", image.URL)
			report.add(sample)
			if len(report.Recent) < 8 {
				report.Recent = append(report.Recent, sample)
			}
		}
	}
	if report.SharedRoots || !report.FilesRootExists {
		return report
	}
	_ = filepath.WalkDir(cfg.FilesRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !isImageName(entry.Name()) {
			return walkErr
		}
		sample := inspectFile(path, "file", "/files/"+filepath.ToSlash(mustRel(cfg.FilesRoot, path)))
		report.FilesImageCount++
		report.add(sample)
		if len(report.Recent) < 8 {
			report.Recent = append(report.Recent, sample)
		}
		return nil
	})
	return report
}

func (report *ImageReport) add(sample ImageSample) {
	if sample.Size == 0 {
		report.EmptyCount++
		return
	}
	if sample.Decodable {
		report.DecodableCount++
		return
	}
	report.UndecodableCount++
}

func inspectFile(path, kind, url string) ImageSample {
	info, err := os.Stat(path)
	sample := ImageSample{Name: filepath.Base(path), Path: path, URL: url, Kind: kind}
	if err != nil {
		sample.Issue = err.Error()
		return sample
	}
	sample.Size = info.Size()
	sample.CreatedAt = info.ModTime().UTC()
	if sample.Size == 0 {
		sample.Issue = "empty file"
		return sample
	}
	width, height, issue := decodeImage(path)
	sample.Width = width
	sample.Height = height
	sample.Issue = issue
	sample.Decodable = issue == "" && width > 0 && height > 0
	if strings.EqualFold(filepath.Ext(path), ".svg") {
		sample.Decodable = issue == ""
	}
	return sample
}

func decodeImage(path string) (int, int, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err.Error()
	}
	if strings.EqualFold(filepath.Ext(path), ".svg") {
		if !bytes.Contains(bytes.ToLower(data), []byte("<svg")) {
			return 0, 0, "not an SVG document"
		}
		return 0, 0, ""
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, "could not decode image dimensions"
	}
	if config.Width == 0 || config.Height == 0 {
		return config.Width, config.Height, "decoded image has zero dimensions"
	}
	return config.Width, config.Height, ""
}

func imageCount(images ImageReport) int {
	if images.SharedRoots {
		return images.IndexedCount
	}
	return images.IndexedCount + images.FilesImageCount
}

func imageDecodeDetail(images ImageReport) string {
	total := imageCount(images)
	if total == 0 {
		return "no images indexed yet"
	}
	if images.EmptyCount == 0 && images.UndecodableCount == 0 {
		return fmt.Sprintf("%d of %d decode with non-zero dimensions", images.DecodableCount, total)
	}
	return fmt.Sprintf("%d empty, %d undecodable", images.EmptyCount, images.UndecodableCount)
}

func skillsCheckStatus(status skills.Status) Severity {
	switch status.State {
	case skills.StateSynced:
		return Pass
	case skills.StateDirty, skills.StateAhead, skills.StateLocked, skills.StateUnconfigured:
		return Warn
	default:
		return Fail
	}
}

func check(id, group, label, detail string, ok, required bool) Check {
	status := Pass
	if !ok && required {
		status = Fail
	} else if !ok {
		status = Warn
	}
	return Check{ID: id, Label: label, Detail: detail, Status: status, Group: group}
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

func commandSucceeds(name string, arguments ...string) bool {
	return exec.Command(name, arguments...).Run() == nil
}

func lingerEnabled() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	output, err := exec.Command("loginctl", "show-user", currentUser.Username, "-p", "Linger", "--value").Output()
	return err == nil && strings.TrimSpace(string(output)) == "yes"
}

func samePath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && leftAbs == rightAbs
}

func isImageName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".avif", ".gif", ".jpeg", ".jpg", ".png", ".svg", ".webp":
		return true
	default:
		return false
	}
}

func mustRel(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Base(path)
	}
	return relative
}

func empty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
