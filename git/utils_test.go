package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func chdir(t *testing.T, dir string) func() {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restoring working directory: %v", err)
		}
	}
}

func checkout(t *testing.T, repo string, commit string, pr string, depth int) (checkoutPath string) {
	repoAbsolutePath, err := filepath.Abs(repo)
	if err != nil {
		t.Fatalf("can't get repo absolute path: %s", err)
	}

	src := filepath.Join(repoAbsolutePath, ".notgit")
	dst := filepath.Join(repoAbsolutePath, ".git")

	err = os.Rename(src, dst)
	if err != nil {
		t.Fatalf("can't move .gitfolder '%s' to '%s': %v", repo, dst, err)
	}

	defer os.Rename(dst, src)

	tmpDir, err := os.MkdirTemp("", "scan-action-")
	if err != nil {
		t.Fatalf("can't create tmp directory: %s", err)
	}

	defer chdir(t, tmpDir)()

	prRef := fmt.Sprintf("refs/remotes/pull/%s/merge", pr)
	commitRef := fmt.Sprintf("+%s:%s", commit, prRef)

	var cmds []*exec.Cmd
	cmds = append(cmds, exec.Command("/usr/bin/git", "version"))
	cmds = append(cmds, exec.Command("/usr/bin/git", "config", "--global", "--add", "safe.directory", tmpDir))
	cmds = append(cmds, exec.Command("/usr/bin/git", "init", tmpDir))
	cmds = append(cmds, exec.Command("/usr/bin/git", "remote", "add", "origin", repoAbsolutePath))
	cmds = append(cmds, exec.Command("/usr/bin/git", "config", "--local", "gc.auto", "0"))

	cmds = append(cmds, exec.Command("/usr/bin/git", "-c", "protocol.version=2", "fetch", "--no-tags", "--prune", "--no-recurse-submodules", fmt.Sprintf("--depth=%d", depth), "origin", commitRef))
	cmds = append(cmds, exec.Command("/usr/bin/git", "sparse-checkout", "disable"))
	cmds = append(cmds, exec.Command("/usr/bin/git", "config", "--local", "--unset-all", "extensions.worktreeConfig"))

	cmds = append(cmds, exec.Command("/usr/bin/git", "checkout", "--progress", "--force", prRef))

	for _, cmd := range cmds {
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("can't run %s: %s\nOutput:\n%s", cmd.String(), err, out)
		}
	}

	return tmpDir
}
