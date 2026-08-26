package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	"github.com/spf13/cobra"
)

// NewRoleCmd creates the `workflow role [name]` command.
func NewRoleCmd(k *knowledge.Knowledge) *cobra.Command {
	var (
		flagDescription    bool
		flagResponsibility bool
		flagBoundary       bool
		flagCommunication  bool
	)

	cmd := &cobra.Command{
		Use:   "role [name]",
		Short: "Inspect participant identities, boundaries, and responsibilities",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			// Check selector count
			activeSelectors := 0
			if flagDescription {
				activeSelectors++
			}
			if flagResponsibility {
				activeSelectors++
			}
			if flagBoundary {
				activeSelectors++
			}
			if flagCommunication {
				activeSelectors++
			}

			if activeSelectors > 1 {
				return errors.New("cannot specify multiple section selectors simultaneously")
			}

			// Bare discovery
			if len(args) == 0 {
				if activeSelectors > 0 {
					return errors.New("cannot use section selectors without specifying a role name")
				}
				return printRolesCatalog(out, k.Roles())
			}

			roleName := args[0]
			role, ok := k.Role(roleName)
			if !ok {
				available := make([]string, 0, len(k.Roles()))
				for _, r := range k.Roles() {
					available = append(available, string(r.Name))
				}
				return fmt.Errorf("unknown role %q (available: %s)", roleName, strings.Join(available, ", "))
			}

			if flagDescription {
				return writeJSON(out, role.Description)
			}
			if flagResponsibility {
				return writeJSON(out, role.Responsibilities)
			}
			if flagBoundary {
				return writeJSON(out, role.Boundaries)
			}
			if flagCommunication {
				return writeJSON(out, role.Communication)
			}

			// Default: return full role JSON
			return writeJSON(out, role)
		},
	}

	cmd.Flags().BoolVar(&flagDescription, "description", false, "Show only role description")
	cmd.Flags().BoolVar(&flagResponsibility, "responsibility", false, "Show only role responsibilities")
	cmd.Flags().BoolVar(&flagBoundary, "boundary", false, "Show only role boundaries")
	cmd.Flags().BoolVar(&flagCommunication, "communication", false, "Show only role communication targets")

	return cmd
}

func printRolesCatalog(w io.Writer, roles []knowledge.RoleDefinition) error {
	var b strings.Builder
	b.WriteString("Available roles:\n\n")
	for _, r := range roles {
		fmt.Fprintf(&b, "  %-10s %s\n", r.Name, r.Description)
	}
	b.WriteString("\nUsage:\n  agentplaybook role <name> [flags]\n\nFlags:\n")
	b.WriteString("  --description       Show only role description\n")
	b.WriteString("  --responsibility    Show only role responsibilities\n")
	b.WriteString("  --boundary          Show only role boundaries\n")
	b.WriteString("  --communication     Show only role communication targets")

	return writeText(w, b.String())
}
