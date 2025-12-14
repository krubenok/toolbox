package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kyrubeno/toolbox/internal/tools/adoprcomments"
)

var adoPRCommentsCmd = &cobra.Command{
	Use:   "ado-pr-comments <PR_URL>",
	Short: "Fetch pull request comments from Azure DevOps",
	Long: `Fetch and display pull request comments from Azure DevOps.

Auth:
  - Uses Azure CLI login if available (Bearer token for Azure DevOps).
  - Otherwise set AZDO_PAT or ADO_PAT (Code -> Read) for Basic auth.

Output:
  By default, output is in TOON format (token-optimized notation).
  Use --json for standard JSON output.

Content Filtering:
  Configure regex patterns to strip boilerplate from comments.
  Create ~/.toolbox/ado-pr-comments.json:

  {
    "filter": {
      "cutPatterns": ["(?i)pattern to cut at"],
      "scrubPatterns": ["(?i)pattern to remove"],
      "authorPatterns": ["(?i)^botname$"]
    }
  }

  - cutPatterns: content after first match is removed
  - scrubPatterns: all matches are removed
  - authorPatterns: only filter comments from matching authors (empty = all)

  See examples/ado-pr-comments.json for a complete example with bot patterns.

Status Filtering:
  By default, all statuses are included. Configure default statuses in config:

  {
    "status": {
      "include": ["active"]
    }
  }

  Override with --status flag (comma-separated or repeated):
    --status active
    --status active,fixed
    --status active --status pending

  Valid statuses: active, fixed, closed, byDesign, pending, wontFix

Examples:
  toolbox ado-pr-comments https://dev.azure.com/org/project/_git/repo/pullrequest/123
  toolbox ado-pr-comments https://org.visualstudio.com/project/_git/repo/pullrequest/123 --status active
  toolbox ado-pr-comments <PR_URL> --json
  toolbox ado-pr-comments <PR_URL> --no-filter`,
	Args: cobra.ExactArgs(1),
	RunE: runAdoPRComments,
}

var (
	adoPRStatuses   []string
	adoPROutputJSON bool
	adoPRDebug      bool
	adoPRNoFilter   bool
)

func init() {
	rootCmd.AddCommand(adoPRCommentsCmd)

	adoPRCommentsCmd.Flags().StringSliceVar(&adoPRStatuses, "status", nil, "Filter by thread status (comma-separated or repeated, e.g., --status active,fixed)")
	adoPRCommentsCmd.Flags().BoolVar(&adoPROutputJSON, "json", false, "Output JSON instead of TOON format")
	adoPRCommentsCmd.Flags().BoolVar(&adoPRDebug, "debug", false, "Print debug info to stderr")
	adoPRCommentsCmd.Flags().BoolVar(&adoPRNoFilter, "no-filter", false, "Disable content filtering")
}

func runAdoPRComments(cmd *cobra.Command, args []string) error {
	opts := adoprcomments.Options{
		Ctx:        cmd.Context(),
		PRURL:      args[0],
		Statuses:   adoPRStatuses,
		OutputJSON: adoPROutputJSON,
		Debug:      adoPRDebug,
		NoFilter:   adoPRNoFilter,
		DebugLog: func(msg string) {
			fmt.Fprintln(os.Stderr, msg)
		},
	}

	result, err := adoprcomments.Run(opts)
	if err != nil {
		return err
	}

	fmt.Println(result.Output)
	return nil
}
