package gitops

import (
	"io"
	"os"
)

type ReplacePatchHandler struct {
	LocalFile string
}

func ReplacePatch(path string) ReplacePatchHandler {
	return ReplacePatchHandler{
		LocalFile: path,
	}
}

func (h ReplacePatchHandler) Patch(_ io.Reader, out io.Writer) error {
	file, err := os.Open(h.LocalFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, file)
	return err
}
