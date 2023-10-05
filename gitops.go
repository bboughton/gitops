package gitops

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

type CommandErr struct {
	Command string
	Err     error
	Stderr  string
}

func (err CommandErr) Error() string {
	return fmt.Sprintf("%v\nexec: %v\nstderr: %v", err.Err.Error(), err.Command, err.Stderr)
}

type PatchHandler interface {
	Patch(in io.Reader, out io.Writer) error
}

func GenPatch(url, file string, patch PatchHandler, out io.Writer, message string) error {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpdir)

	err = exec.Command("git", "clone", "--depth", "1", url, tmpdir).Run()
	if err != nil {
		return fmt.Errorf("unable to clone repo %q: %w", url, err)
	}

	orig, err := os.Open(filepath.Join(tmpdir, file))
	if err != nil {
		return err
	}

	var b bytes.Buffer
	err = patch.Patch(orig, &b)
	if err != nil {
		return fmt.Errorf("patching: %w", err)
	}

	err = orig.Close()
	if err != nil {
		return err
	}

	update, err := os.Create(filepath.Join(tmpdir, file))
	if err != nil {
		return err
	}

	io.Copy(update, &b)

	err = exec.Command("git", "-C", tmpdir, "diff", "--quiet").Run()
	if err == nil {
		return fmt.Errorf("no changes detected")
	}

	err = exec.Command("git", "-C", tmpdir, "add", file).Run()
	if err != nil {
		return fmt.Errorf("command failed: git add: %w", err)
	}

	err = exec.Command("git", "-C", tmpdir, "commit", "-m", message).Run()
	if err != nil {
		return fmt.Errorf("command failed: git commit: %w", err)
	}

	cmd := exec.Command("git", "-C", tmpdir, "format-patch", "HEAD^", "--stdout")
	cmd.Stdout = out
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: git format-patch: %w", err)
	}

	return nil
}

func PushPatch(url, patch string) error {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpdir)

	err = exec.Command("git", "clone", "--depth", "1", url, tmpdir).Run()
	if err != nil {
		return fmt.Errorf("clone repo %q: %w", url, err)
	}

	patch, err = filepath.Abs(patch)
	if err != nil {
		return fmt.Errorf("patch file: %w", err)
	}

	_, err = os.Stat(patch)
	if err != nil {
		return fmt.Errorf("patch file: %w", err)
	}

	err = exec.Command("git", "-C", tmpdir, "am", patch).Run()
	if err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	err = exec.Command("git", "-C", tmpdir, "push").Run()
	if err != nil {
		return fmt.Errorf("push patch: %w", err)
	}

	return nil
}

func ParseCompositeURL(u string) (string, string) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", ""
	}
	fragment := parsed.Fragment
	parsed.Fragment = ""
	return parsed.String(), fragment
}
