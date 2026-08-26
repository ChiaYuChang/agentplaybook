package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	"github.com/spf13/cobra"
)

// NewArtifactCmd creates the `workflow artifact [name]` command.
func NewArtifactCmd(k *knowledge.Knowledge) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact [name]",
		Short: "Inspect document contracts, required sections, and message schemas",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			// Bare discovery
			if len(args) == 0 {
				return printArtifactsCatalog(out, k.Artifacts())
			}

			artifactName := args[0]
			artifact, ok := k.Artifact(artifactName)
			if !ok {
				available := make([]string, 0, len(k.Artifacts()))
				for _, a := range k.Artifacts() {
					available = append(available, a.Name)
				}
				return fmt.Errorf("unknown artifact %q (available: %s)", artifactName, strings.Join(available, ", "))
			}

			// Return full artifact JSON
			return writeJSON(out, artifact)
		},
	}

	return cmd
}

func printArtifactsCatalog(w io.Writer, artifacts []knowledge.Artifact) error {
	var b strings.Builder
	b.WriteString("Available artifacts:\n\n")
	for _, a := range artifacts {
		fmt.Fprintf(&b, "  %-16s %s\n", a.Name, a.Description)
	}
	b.WriteString("\nUsage:\n  workflow artifact <name>")

	return writeText(w, b.String())
}
