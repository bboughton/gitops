package main

import (
	"errors"
	"io"
	"os"

	"github.com/bboughton/gitops"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:          "gitops",
		Short:        "utility cli for interacting with gitops repositories",
		SilenceUsage: true,
	}

	root.AddCommand(func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "format-patch",
			Short: "prepare patch for submission",
			Long:  "prepare each non-merge commit with its \"patch\" in one \"message\" per commit, formatted to resemble a UNIX mailbox. The output of this command is compatible with the output from git-format-patch and will work with gitops push-patch.",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return errors.New("composite url required")
				}

				repo, file := gitops.ParseCompositeURL(args[0])

				sedCmds, err := cmd.Flags().GetStringSlice("sed")
				if err != nil {
					return err
				}

				yqFilter, err := cmd.Flags().GetString("yq")
				if err != nil {
					return err
				}

				jqFilter, err := cmd.Flags().GetString("jq")
				if err != nil {
					return err
				}

				localFile, err := cmd.Flags().GetString("file")
				if err != nil {
					return err
				}

				out, err := parseOut(cmd.Flags().GetString("out"))
				if err != nil {
					return err
				}

				message, err := cmd.Flags().GetString("message")
				if err != nil {
					return err
				}

				var patcher gitops.PatchHandler
				if localFile != "" {
					patcher = gitops.ReplacePatch(localFile)
				} else if yqFilter != "" {
					patcher = gitops.YqPatch(yqFilter)
				} else if jqFilter != "" {
					patcher = gitops.JqPatch(jqFilter)
				} else if len(sedCmds) > 0 {
					patcher = gitops.SedPatch(sedCmds)
				} else {
					return errors.New("no patch strategry selected")
				}

				return gitops.GenPatch(repo, file, patcher, out, message)
			},
		}
		cmd.Flags().StringSlice("sed", nil, "sed command used to update the provided file")
		cmd.Flags().String("yq", "", "yq filter used to update the provided file")
		cmd.Flags().String("file", "", "local file to replace the remote file")
		cmd.Flags().String("jq", "", "jq filter used to update the provided file")
		cmd.Flags().String("out", "-", "path to write patch file, when set to '-' write to stdout")
		cmd.Flags().String("message", "", "message to use for the commit, when left blank a generic message will be used")
		return cmd
	}())

	root.AddCommand(func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "push-patch",
			Short: "push patch to repository",
			Long:  "push the given patch file to the given gitops repository",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return errors.New("repo url required")
				}

				repo := args[0]

				patch, err := cmd.Flags().GetString("patch")
				if err != nil {
					return err
				}

				return gitops.PushPatch(repo, patch)
			},
		}
		cmd.Flags().String("patch", "", "path to patch file")
		return cmd
	}())

	root.Execute()
}

func parseOut(out string, err error) (io.WriteCloser, error) {
	if err != nil {
		return nil, err
	}

	if out == "" || out == "-" {
		return writeNoopCloser{os.Stdout}, nil
	}

	f, err := os.Create(out)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type writeNoopCloser struct {
	io.Writer
}

func (wc writeNoopCloser) Close() error { return nil }
