package portly

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationsMapsIdentityAndPublicURL(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "portly")
	script := `#!/bin/sh
cat <<'EOF'
{"projects":[{"id":"prj_1","name":"kliq","root":"/root/projects/kliq","icon":"link","color":"#111111","servers":[{"id":"srv_1","name":"start-all","projectID":"prj_1","projectName":"kliq","command":"pnpm start-all","directory":"/root/projects/kliq","state":"running","port":3050,"pid":9,"healthy":true,"url":"http://localhost:3050","cpuPercent":1.5,"memoryBytes":20,"residentMemoryBytes":18,"restartCount":0}]}]}
EOF
`
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	applications, err := (Client{Binary: binary, PublicPortHost: "port{port}.mlvcdn.com"}).Applications(context.Background())
	if err != nil {
		t.Fatalf("Applications() error = %v", err)
	}
	if len(applications) != 1 {
		t.Fatalf("len = %d", len(applications))
	}
	app := applications[0]
	if app.Domain != "kliq.sh" || app.ProjectColor != "#111111" || app.ProjectIcon != "link" || app.PublicURL != "https://port3050.mlvcdn.com" {
		t.Fatalf("application = %+v", app)
	}
}

func TestStopPassesExactServerIDToPortly(t *testing.T) {
	directory := t.TempDir()
	logPath := filepath.Join(directory, "arguments")
	binary := filepath.Join(directory, "portly")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$CODE_OS_TEST_LOG\"\n"
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODE_OS_TEST_LOG", logPath)

	if err := (Client{Binary: binary}).Stop(context.Background(), "srv_safe-id"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	arguments, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(arguments)) != "stop srv_safe-id --json" {
		t.Fatalf("arguments = %q", arguments)
	}
}

func TestStopReturnsPortlyFailureWithoutShellExecution(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	binary := filepath.Join(directory, "portly")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho 'not found' >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	err := (Client{Binary: binary}).Stop(context.Background(), "$(touch /tmp/never)")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Stop() error = %v", err)
	}
}
