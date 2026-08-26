package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	"github.com/spf13/cobra"
)

var defaultVersion = "dev"

// Execute runs the CLI with the provided arguments and streams.
func Execute(args []string, stdout, stderr io.Writer, version string) error {
	k, err := knowledge.Load()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return err
	}

	if version == "" {
		version = defaultVersion
	}

	rootCmd := NewRootCmd(k, version)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	return rootCmd.Execute()
}

// NewRootCmd constructs the Cobra root command hierarchy.
func NewRootCmd(k *knowledge.Knowledge, version string) *cobra.Command {
	var (
		flagLanguage  bool
		flagTransport bool
		flagVersion   bool
	)

	cmd := &cobra.Command{
		Use:           "workflow",
		Short:         "workflow - Multi-Agent Collaboration Manual CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			activeFlags := 0
			if flagLanguage {
				activeFlags++
			}
			if flagTransport {
				activeFlags++
			}
			if flagVersion {
				activeFlags++
			}

			if activeFlags > 1 {
				return errors.New("cannot specify multiple query flags simultaneously")
			}

			out := cmd.OutOrStdout()

			if flagLanguage {
				return writeJSON(out, k.Config.Languages)
			}
			if flagTransport {
				return writeText(out, k.Config.Transport)
			}
			if flagVersion {
				return writeText(out, version)
			}

			// Bare invocation: print discovery manual
			return printRootDiscovery(out)
		},
	}

	cmd.Flags().BoolVar(&flagLanguage, "language", false, "Show supported collaboration languages")
	cmd.Flags().BoolVar(&flagTransport, "transport", false, "Show underlying communication transport")
	cmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "Show CLI version")

	cmd.AddCommand(NewRoleCmd(k))
	cmd.AddCommand(NewFlowCmd(k))
	cmd.AddCommand(NewArtifactCmd(k))
	cmd.AddCommand(NewRuleCmd(k))

	return cmd
}

func printRootDiscovery(w io.Writer) error {
	catalog := `workflow - Multi-Agent Collaboration Manual CLI

A read-only informational manual for multi-agent workflows.
Query roles, flows, artifact contracts, and behavioral rules on demand.

Usage:
  workflow [command]
  workflow [flags]

Knowledge Domains:
  role        Inspect participant identities, boundaries, and responsibilities
  flow        Inspect end-to-end multi-agent orchestration procedures
  artifact    Inspect document contracts, required sections, and message schemas
  rule        Inspect operational policies, invariants, and protocols

Global Flags:
  --language   Show supported collaboration languages
  --transport  Show underlying communication transport
  --version    Show CLI version`
	return writeText(w, catalog)
}
