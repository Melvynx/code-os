//go:build linux

package processes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/melvynx/code-os/internal/model"
)

const clockTicksPerSecond = 100

type Scanner struct {
	ProcRoot string
}

type process struct {
	pid       int
	parentPID int
	startTime uint64
	cpuTicks  uint64
	memory    int64
	command   string
	arguments string
}

func (scanner Scanner) Scan() ([]model.AgentProcess, error) {
	processes, uptime, err := scanner.readProcesses()
	if err != nil {
		return nil, err
	}
	return agentProcesses(processes, uptime), nil
}

func (scanner Scanner) Terminate(id string) error {
	pid, startTime, err := parseID(id)
	if err != nil {
		return err
	}
	processes, _, err := scanner.readProcesses()
	if err != nil {
		return err
	}
	target, exists := processes[pid]
	if !exists || target.startTime != startTime || !isAgentRoot(target, processes) {
		return errors.New("agent process is no longer available")
	}
	if isAncestor(pid, os.Getpid(), processes) {
		return errors.New("refusing to terminate the process hosting Code OS")
	}
	running, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find agent process: %w", err)
	}
	if err := running.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("terminate agent process: %w", err)
	}
	return nil
}

func (scanner Scanner) readProcesses() (map[int]process, float64, error) {
	root := scanner.ProcRoot
	if root == "" {
		root = "/proc"
	}
	uptimeBytes, err := os.ReadFile(filepath.Join(root, "uptime"))
	if err != nil {
		return nil, 0, fmt.Errorf("read process uptime: %w", err)
	}
	uptimeFields := strings.Fields(string(uptimeBytes))
	if len(uptimeFields) == 0 {
		return nil, 0, errors.New("process uptime is empty")
	}
	uptime, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		return nil, 0, fmt.Errorf("parse process uptime: %w", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, 0, fmt.Errorf("read process directory: %w", err)
	}
	result := make(map[int]process)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || !entry.IsDir() {
			continue
		}
		item, err := readProcess(filepath.Join(root, entry.Name()), pid)
		if err == nil {
			result[pid] = item
		}
	}
	return result, uptime, nil
}

func readProcess(directory string, pid int) (process, error) {
	statBytes, err := os.ReadFile(filepath.Join(directory, "stat"))
	if err != nil {
		return process{}, err
	}
	stat := string(statBytes)
	closing := strings.LastIndex(stat, ")")
	opening := strings.Index(stat, "(")
	if opening < 0 || closing <= opening || closing+2 >= len(stat) {
		return process{}, errors.New("invalid process stat")
	}
	fields := strings.Fields(stat[closing+2:])
	if len(fields) < 22 {
		return process{}, errors.New("incomplete process stat")
	}
	parentPID, err := strconv.Atoi(fields[1])
	if err != nil {
		return process{}, err
	}
	userTicks, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return process{}, err
	}
	systemTicks, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return process{}, err
	}
	startTime, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return process{}, err
	}
	statusBytes, _ := os.ReadFile(filepath.Join(directory, "status"))
	cmdlineBytes, _ := os.ReadFile(filepath.Join(directory, "cmdline"))
	return process{
		pid: pid, parentPID: parentPID, startTime: startTime, cpuTicks: userTicks + systemTicks,
		memory: residentMemory(statusBytes), command: stat[opening+1 : closing],
		arguments: strings.TrimSpace(strings.ReplaceAll(string(cmdlineBytes), "\x00", " ")),
	}, nil
}

func residentMemory(status []byte) int64 {
	for _, line := range strings.Split(string(status), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "VmRSS:" {
			kilobytes, _ := strconv.ParseInt(fields[1], 10, 64)
			return kilobytes * 1024
		}
	}
	return 0
}

func agentProcesses(processes map[int]process, uptime float64) []model.AgentProcess {
	children := make(map[int][]int)
	for pid, item := range processes {
		children[item.parentPID] = append(children[item.parentPID], pid)
	}
	result := make([]model.AgentProcess, 0)
	for _, item := range processes {
		if !isAgentRoot(item, processes) {
			continue
		}
		pids := descendants(item.pid, children)
		var cpuTicks uint64
		var memory int64
		for _, pid := range pids {
			cpuTicks += processes[pid].cpuTicks
			memory += processes[pid].memory
		}
		lifetime := uptime - float64(item.startTime)/clockTicksPerSecond
		cpuPercent := 0.0
		if lifetime > 0 {
			cpuPercent = float64(cpuTicks) / clockTicksPerSecond / lifetime * 100
		}
		result = append(result, model.AgentProcess{
			ID: processID(item), Name: agentName(item), Command: item.command, PID: item.pid,
			CPUPercent: cpuPercent, MemoryBytes: memory, ProcessCount: len(pids),
		})
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].MemoryBytes == result[right].MemoryBytes {
			return result[left].PID < result[right].PID
		}
		return result[left].MemoryBytes > result[right].MemoryBytes
	})
	return result
}

func isAgentRoot(item process, processes map[int]process) bool {
	if !isAgent(item) {
		return false
	}
	for parentPID := item.parentPID; parentPID > 1; {
		parent, exists := processes[parentPID]
		if !exists {
			break
		}
		if isAgent(parent) {
			return false
		}
		parentPID = parent.parentPID
	}
	return true
}

func isAgent(item process) bool {
	command := strings.ToLower(item.command)
	arguments := strings.ToLower(item.arguments)
	knownCommands := []string{"codex", "codex-code-mode", "claude", "cursor", "cursor-agent", "opencode", "aider", "gemini"}
	for _, known := range knownCommands {
		if command == known || strings.HasPrefix(command, known+"-") {
			return true
		}
	}
	if isCommandWrapper(command) {
		return false
	}
	signatures := []string{
		"/.codex/", "@openai/codex",
		"/cursor-agent", "anysphere.cursor-agent-worker",
		"@anthropic-ai/claude-code", "/claude-code/",
		"/opencode/", "/opencode-",
		"/aider/", "aider.main",
		"@google/gemini-cli", "/gemini-cli/",
	}
	for _, signature := range signatures {
		if strings.Contains(arguments, signature) {
			return true
		}
	}
	return false
}

func isCommandWrapper(command string) bool {
	switch command {
	case "bash", "sh", "zsh", "fish", "rg", "grep", "sed", "awk", "ps":
		return true
	default:
		return false
	}
}

func agentName(item process) string {
	command := strings.ToLower(item.command)
	arguments := strings.ToLower(item.arguments)
	switch {
	case strings.Contains(command, "codex") || strings.Contains(arguments, "/.codex/") || strings.Contains(arguments, "@openai/codex"):
		if strings.Contains(arguments, "proxy") {
			return "Codex proxy"
		}
		return "Codex agent"
	case strings.Contains(command, "claude") || strings.Contains(arguments, "claude-code"):
		return "Claude agent"
	case strings.Contains(command, "cursor") || strings.Contains(arguments, "cursor-agent"):
		return "Cursor agent"
	case strings.Contains(command, "opencode") || strings.Contains(arguments, "/opencode"):
		return "OpenCode agent"
	case strings.Contains(command, "aider") || strings.Contains(arguments, "aider.main"):
		return "Aider agent"
	case strings.Contains(command, "gemini") || strings.Contains(arguments, "gemini-cli"):
		return "Gemini agent"
	default:
		return "Development agent"
	}
}

func descendants(root int, children map[int][]int) []int {
	result := []int{root}
	for index := 0; index < len(result); index++ {
		result = append(result, children[result[index]]...)
	}
	return result
}

func isAncestor(target, child int, processes map[int]process) bool {
	for child > 1 {
		if child == target {
			return true
		}
		item, exists := processes[child]
		if !exists {
			return false
		}
		child = item.parentPID
	}
	return false
}

func processID(item process) string {
	return fmt.Sprintf("%d:%d", item.pid, item.startTime)
}

func parseID(id string) (int, uint64, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid agent process identifier")
	}
	pid, err := strconv.Atoi(parts[0])
	if err != nil || pid <= 1 {
		return 0, 0, errors.New("invalid agent process identifier")
	}
	startTime, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || startTime == 0 {
		return 0, 0, errors.New("invalid agent process identifier")
	}
	return pid, startTime, nil
}
