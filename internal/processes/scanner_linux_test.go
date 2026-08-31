//go:build linux

package processes

import (
	"testing"
)

func TestAgentProcessesGroupsDescendantResources(t *testing.T) {
	t.Parallel()
	items := map[int]process{
		100: {pid: 100, parentPID: 1, startTime: 50_000, cpuTicks: 100, memory: 1000, command: "codex", arguments: "codex app-server"},
		101: {pid: 101, parentPID: 100, startTime: 50_010, cpuTicks: 50, memory: 2000, command: "node", arguments: "node mcp"},
		102: {pid: 102, parentPID: 100, startTime: 50_020, cpuTicks: 25, memory: 3000, command: "codex-code-mode", arguments: "codex-code-mode-host"},
		200: {pid: 200, parentPID: 2, startTime: 75_000, cpuTicks: 40, memory: 4000, command: "codex", arguments: "codex app-server proxy"},
		300: {pid: 300, parentPID: 1, startTime: 80_000, cpuTicks: 1000, memory: 5000, command: "node", arguments: "node server.js"},
	}

	agents := agentProcesses(items, 1000)

	if len(agents) != 2 {
		t.Fatalf("agents = %d, want 2", len(agents))
	}
	if agents[0].PID != 100 || agents[0].MemoryBytes != 6000 || agents[0].ProcessCount != 3 {
		t.Fatalf("grouped Codex agent = %#v", agents[0])
	}
	if agents[1].Name != "Codex proxy" || agents[1].ProcessCount != 1 {
		t.Fatalf("proxy agent = %#v", agents[1])
	}
}

func TestProcessIDRejectsPIDReuseAndMalformedValues(t *testing.T) {
	t.Parallel()
	pid, startTime, err := parseID("42:9001")
	if err != nil || pid != 42 || startTime != 9001 {
		t.Fatalf("parseID() = %d, %d, %v", pid, startTime, err)
	}
	for _, value := range []string{"", "42", "0:1", "42:0", "pid:start"} {
		if _, _, err := parseID(value); err == nil {
			t.Fatalf("parseID(%q) accepted invalid value", value)
		}
	}
}

func TestIsAncestorOnlyMatchesProcessLineage(t *testing.T) {
	t.Parallel()
	items := map[int]process{
		10: {pid: 10, parentPID: 1},
		11: {pid: 11, parentPID: 10},
		12: {pid: 12, parentPID: 11},
		20: {pid: 20, parentPID: 1},
	}
	if !isAncestor(10, 12, items) {
		t.Fatal("expected PID 10 to be an ancestor of PID 12")
	}
	if isAncestor(20, 12, items) {
		t.Fatal("unrelated PID was treated as an ancestor")
	}
}

func TestAgentSignaturesDetectWrappedCLIsWithoutMatchingShellCommands(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		item      process
		agentName string
	}{
		{
			name:      "Cursor worker with generic process name",
			item:      process{command: "MainThread", arguments: "/root/.cursor-server/agent-cli/.local/bin/cursor-agent worker start"},
			agentName: "Cursor agent",
		},
		{
			name:      "Claude distributed as a Node CLI",
			item:      process{command: "node", arguments: "node /opt/node_modules/@anthropic-ai/claude-code/cli.js"},
			agentName: "Claude agent",
		},
		{
			name:      "Gemini distributed as a Node CLI",
			item:      process{command: "node", arguments: "node /opt/node_modules/@google/gemini-cli/dist/index.js"},
			agentName: "Gemini agent",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !isAgent(test.item) {
				t.Fatal("expected wrapped CLI to be detected as an agent")
			}
			if got := agentName(test.item); got != test.agentName {
				t.Fatalf("agentName() = %q, want %q", got, test.agentName)
			}
		})
	}

	shell := process{command: "bash", arguments: "bash -lc rg cursor-agent /tmp/log"}
	if isAgent(shell) {
		t.Fatal("shell command mentioning an agent was treated as an agent process")
	}
}
