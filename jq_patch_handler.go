package gitops

import (
	"encoding/json"
	"io"

	"github.com/itchyny/gojq"
)

type JqPatchHander struct {
	Filter string
}

func (h JqPatchHander) Patch(in io.Reader, out io.Writer) error {
	query, err := gojq.Parse(h.Filter)
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	var data map[string]any
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "    ")
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		err = encoder.Encode(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func JqPatch(filter string) JqPatchHander {
	return JqPatchHander{
		Filter: filter,
	}
}
