package gitops

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

type SedPatchHandler struct {
	Commands []string
}

func (h SedPatchHandler) Patch(in io.Reader, out io.Writer) error {
	if len(h.Commands) < 1 {
		return errors.New("sed: no commands provided")
	}

	cmds := []string{"-i", "sed", "-E"}
	for i := 0; i < len(h.Commands); i++ {
		cmds = append(cmds, "-e", h.Commands[i])
	}
	cmd := exec.Command("env", cmds...)

	var b bytes.Buffer
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = &b

	err := cmd.Run()
	if err != nil {
		return CommandErr{
			Command: cmd.String(),
			Err:     err,
			Stderr:  b.String(),
		}
	}
	return nil
}

func SedPatch(cmds []string) SedPatchHandler {
	return SedPatchHandler{
		Commands: cmds,
	}
}
