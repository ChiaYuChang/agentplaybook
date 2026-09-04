package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	"github.com/spf13/cobra"
)

// WriteLivingMemoryTemplate writes the canonical living memory template to the provided writer.
// If minimal is true, MinimalLivingMemoryTemplate is written; otherwise DefaultLivingMemoryTemplate is written.
func WriteLivingMemoryTemplate(w io.Writer, minimal bool) error {
	var content string
	if minimal {
		content = MinimalLivingMemoryTemplate()
	} else {
		content = DefaultLivingMemoryTemplate()
	}
	_, err := io.WriteString(w, content)
	return err
}

// NewInitCmd constructs the 'init' command for scaffolding AGENTS.md living memory.
func NewInitCmd(k *knowledge.Knowledge) *cobra.Command {
	var (
		filePath string
		force    bool
		minimal  bool
	)

	cmd := &cobra.Command{
		Use:   "init [flags]",
		Short: "Initialize a standard AGENTS.md living memory file with anti-compaction governance",
		Long: `Initialize a standard AGENTS.md living memory file with anti-compaction governance.

AgentPlaybook remains an evidence-based, read-only guidance manual; the 'agentplaybook init'
command streams the standard AGENTS.md template to stdout by default (zero filesystem writes).
When invoked with --file/-f, it acts as an explicit opt-in local scaffolding utility executed
strictly upon operator invocation to generate baseline AGENTS.md, with zero background
mutation, network downloads, or daemon processes. Use --minimal/-m for an ultra-compact
telegraphic Caveman-style template optimized for minimal context budget.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if force && filePath == "" {
				return fmt.Errorf("--force requires --file")
			}

			// If no target file is specified, stream template directly to stdout (zero disk writes)
			if filePath == "" {
				return WriteLivingMemoryTemplate(cmd.OutOrStdout(), minimal)
			}

			// Preflight collision check
			if _, err := os.Stat(filePath); err == nil {
				if !force {
					return fmt.Errorf("%s already exists; use --force to overwrite", filePath)
				}
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check %s: %w", filePath, err)
			}

			// Ensure parent directories exist
			parentDir := filepath.Dir(filePath)
			if parentDir != "" && parentDir != "." {
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
				}
			}

			// Open file for write (create, write-only, truncate)
			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", filePath, err)
			}

			// Explicitly enforce 0644 permissions (OpenFile mode only applies to created files)
			if err := os.Chmod(filePath, 0644); err != nil {
				_ = f.Close()
				return fmt.Errorf("failed to set permissions on %s: %w", filePath, err)
			}

			if err := WriteLivingMemoryTemplate(f, minimal); err != nil {
				_ = f.Close()
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}

			if err := f.Close(); err != nil {
				return fmt.Errorf("failed to close %s: %w", filePath, err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Initialized standard AGENTS.md at %s\n", filePath)
			return err
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Target file path (if omitted, template is streamed to stdout)")
	cmd.Flags().BoolVarP(&force, "force", "F", false, "Overwrite target file if it already exists (requires --file)")
	cmd.Flags().BoolVarP(&minimal, "minimal", "m", false, "Use ultra-compact telegraphic Caveman-style living memory template")

	return cmd
}
