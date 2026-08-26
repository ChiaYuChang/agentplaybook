package cli

import (
	"fmt"
	"io"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	"github.com/spf13/cobra"
)

type ruleSummaryItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
}

// NewRuleCmd creates the `workflow rule` parent command and subcommands.
func NewRuleCmd(k *knowledge.Knowledge) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rule [command]",
		Short: "Inspect operational policies, invariants, and protocols",
		RunE: func(cmd *cobra.Command, args []string) error {
			return printRuleDiscovery(cmd.OutOrStdout())
		},
	}

	cmd.AddCommand(newRuleListCmd(k))
	cmd.AddCommand(newRuleExplainCmd(k))

	return cmd
}

func newRuleListCmd(k *knowledge.Knowledge) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available workflow rules and summaries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rules := k.Rules()
			items := make([]ruleSummaryItem, 0, len(rules))
			for _, r := range rules {
				items = append(items, ruleSummaryItem{
					ID:       r.ID,
					Title:    r.Title,
					Category: r.Category,
					Summary:  r.Summary,
				})
			}
			return writeJSON(cmd.OutOrStdout(), items)
		},
	}
}

func newRuleExplainCmd(k *knowledge.Knowledge) *cobra.Command {
	return &cobra.Command{
		Use:   "explain <id>...",
		Short: "Show full details and guidelines for specific rule IDs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results := make([]knowledge.Rule, 0, len(args))
			for _, id := range args {
				r, ok := k.Rule(id)
				if !ok {
					return fmt.Errorf("unknown rule %q (run 'agentplaybook rule list' to see available rules)", id)
				}
				results = append(results, r)
			}
			return writeJSON(cmd.OutOrStdout(), results)
		},
	}
}

func printRuleDiscovery(w io.Writer) error {
	catalog := `agentplaybook rule - Operational Policies and Behavioral Invariants

Commands:
  list              List all available workflow rules and summaries
  explain <id>...   Show full details and guidelines for specific rule IDs

Usage:
  agentplaybook rule [command]`
	return writeText(w, catalog)
}
