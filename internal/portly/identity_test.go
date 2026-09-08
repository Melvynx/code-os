package portly

import "testing"

func TestInferDomainFromProjectName(t *testing.T) {
	t.Parallel()
	if got := InferDomain("lumail.io", "/root/projects/lumail.io"); got != "lumail.io" {
		t.Fatalf("lumail.io = %q", got)
	}
	if got := InferDomain("spyland.ing", "/root/projects/spyland.ing"); got != "spyland.ing" {
		t.Fatalf("spyland.ing = %q", got)
	}
}

func TestInferDomainFromKnownProductAlias(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"kliq":            "kliq.sh",
		"codeline":        "codeline.app",
		"ai-builder-club": "ai-builder.club",
	}
	for name, want := range cases {
		if got := InferDomain(name, "/root/projects/"+name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestInferDomainFromWorktreePath(t *testing.T) {
	t.Parallel()
	path := "/root/.cursor/worktrees/lumail.io__SSH__code-vps_/k9ko"
	if got := InferDomain("lumail-k9ko", path); got != "lumail.io" {
		t.Fatalf("worktree domain = %q", got)
	}
}

func TestInferDomainIgnoresSupportingProjects(t *testing.T) {
	t.Parallel()
	if got := InferDomain("dokploy-github-status", "/root/.agents/dokploy-github-status"); got != "" {
		t.Fatalf("supporting project domain = %q", got)
	}
	if got := InferDomain("notes", "/tmp/readme.md"); got != "" {
		t.Fatalf("file-like domain = %q", got)
	}
}
