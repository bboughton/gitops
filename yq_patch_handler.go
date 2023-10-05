package gitops

import (
	"bufio"
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
)

type YqPatchHander struct {
	Filter string
}

func (h YqPatchHander) Patch(in io.Reader, out io.Writer) error {
	printer := yqlib.NewPrinter(yqlib.NewYamlEncoder(2, false, yqlib.NewDefaultYamlPreferences()), yqlib.NewSinglePrinterWriter(out))

	// Discard log output, required as yqlib uses a global log lib
	logging.SetBackend(logging.NewLogBackend(io.Discard, "", 0))

	yqlib.InitExpressionParser()
	node, err := yqlib.ExpressionParser.ParseExpression(h.Filter)
	if err != nil {
		return err
	}

	input, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(bytes.NewReader(input))

	var currentIndex uint
	decoder := yqlib.NewYamlDecoder(yqlib.NewDefaultYamlPreferences())
	err = decoder.Init(reader)
	if err != nil {
		return err
	}
	var fileIndex int
	treeNavigator := yqlib.NewDataTreeNavigator()
	for {
		candidateNode, errorReading := decoder.Decode()

		if errors.Is(errorReading, io.EOF) {
			fileIndex = fileIndex + 1
			return nil
		} else if errorReading != nil {
			return fmt.Errorf("bad input '%v': %w", input, errorReading)
		}
		candidateNode.Document = currentIndex
		candidateNode.FileIndex = fileIndex

		inputList := list.New()
		inputList.PushBack(candidateNode)

		result, errorParsing := treeNavigator.GetMatchingNodes(yqlib.Context{MatchingNodes: inputList}, node)
		if errorParsing != nil {
			return errorParsing
		}
		err = printer.PrintResults(result.MatchingNodes)
		if err != nil {
			return err
		}
		currentIndex = currentIndex + 1
	}
}

func YqPatch(filter string) YqPatchHander {
	return YqPatchHander{
		Filter: filter,
	}
}
