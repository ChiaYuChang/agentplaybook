package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	"github.com/spf13/cobra"
)

// NewFlowCmd creates the `workflow flow [name]` command.
func NewFlowCmd(k *knowledge.Knowledge) *cobra.Command {
	var flagStep int

	cmd := &cobra.Command{
		Use:   "flow [name]",
		Short: "Inspect end-to-end multi-agent orchestration procedures",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			// Bare discovery
			if len(args) == 0 {
				if flagStep != 0 {
					return errors.New("cannot use --step flag without specifying a flow name")
				}
				return printFlowsCatalog(out, k.Flows())
			}

			flowName := args[0]
			flow, ok := k.Flow(flowName)
			if !ok {
				available := make([]string, 0, len(k.Flows()))
				for _, f := range k.Flows() {
					available = append(available, f.Name)
				}
				return fmt.Errorf("unknown flow %q (available: %s)", flowName, strings.Join(available, ", "))
			}

			if flagStep < 0 {
				return errors.New("step index must be positive")
			}

			if flagStep > 0 {
				step, ok := k.FlowStep(flowName, flagStep)
				if !ok {
					return fmt.Errorf("flow %q step %d not found (available: steps 1 to %d)", flowName, flagStep, len(flow.Steps))
				}
				return writeJSON(out, step)
			}

			// Default: return full flow JSON
			return writeJSON(out, flow)
		},
	}

	cmd.Flags().IntVar(&flagStep, "step", 0, "Show only the specified step index within the flow")

	return cmd
}

func printFlowsCatalog(w io.Writer, flows []knowledge.Flow) error {
	var b strings.Builder
	b.WriteString("Available flows:\n\n")
	for _, f := range flows {
		fmt.Fprintf(&b, "  %-8s %s\n", f.Name, f.Description)
	}
	b.WriteString("\nUsage:\n  agentplaybook flow <name> [flags]\n\nFlags:\n")
	b.WriteString("  --step int   Show only the specified step index within the flow")

	return writeText(w, b.String())
}
