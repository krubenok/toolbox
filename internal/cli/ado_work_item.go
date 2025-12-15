package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/krubenok/toolbox/internal/tools/adoworkitem"
)

var adoWorkItemCmd = &cobra.Command{
	Use:   "ado-work-item <WORK_ITEM_URL>",
	Short: "Fetch work item details from Azure DevOps",
	Long: `Fetch and display work item details from Azure DevOps.

Auth:
  - Uses Azure CLI login if available (Bearer token for Azure DevOps).
  - Otherwise set AZDO_PAT or ADO_PAT (Work Items -> Read) for Basic auth.

Output:
  By default, output is in TOON format (token-optimized notation).
  Use --json for standard JSON output.

Included By Default:
  - Description
  - Discussion (work item comments)
  - Child links
  - Attachment links

Examples:
  toolbox ado-work-item https://dev.azure.com/org/project/_workitems/edit/1144734
  toolbox ado-work-item https://org.visualstudio.com/project/_workitems/edit/1144734 --json
  toolbox ado-work-item <WORK_ITEM_URL> --no-discussion
  toolbox ado-work-item <WORK_ITEM_URL> --max-comments 50`,
	Args: cobra.ExactArgs(1),
	RunE: runAdoWorkItem,
}

var (
	adoWIDOutputJSON bool
	adoWIDDebug      bool

	adoWIDNoDescription bool
	adoWIDNoDiscussion  bool
	adoWIDNoChildren    bool
	adoWIDNoAttachments bool
	adoWIDMaxComments   int
)

func init() {
	rootCmd.AddCommand(adoWorkItemCmd)

	adoWorkItemCmd.Flags().BoolVar(&adoWIDOutputJSON, "json", false, "Output JSON instead of TOON format")
	adoWorkItemCmd.Flags().BoolVar(&adoWIDDebug, "debug", false, "Print debug info to stderr")

	adoWorkItemCmd.Flags().BoolVar(&adoWIDNoDescription, "no-description", false, "Do not include the work item description")
	adoWorkItemCmd.Flags().BoolVar(&adoWIDNoDiscussion, "no-discussion", false, "Do not include work item comments/discussion")
	adoWorkItemCmd.Flags().BoolVar(&adoWIDNoChildren, "no-children", false, "Do not include child work item links")
	adoWorkItemCmd.Flags().BoolVar(&adoWIDNoAttachments, "no-attachments", false, "Do not include attachment links")
	adoWorkItemCmd.Flags().IntVar(&adoWIDMaxComments, "max-comments", 0, "Maximum number of discussion comments to fetch (0 = no limit)")
}

func runAdoWorkItem(cmd *cobra.Command, args []string) error {
	opts := adoworkitem.Options{
		Ctx:         cmd.Context(),
		WorkItemURL: args[0],

		IncludeDescription: !adoWIDNoDescription,
		IncludeDiscussion:  !adoWIDNoDiscussion,
		IncludeChildren:    !adoWIDNoChildren,
		IncludeAttachments: !adoWIDNoAttachments,
		MaxComments:        adoWIDMaxComments,

		OutputJSON: adoWIDOutputJSON,
		Debug:      adoWIDDebug,
		DebugLog: func(msg string) {
			fmt.Fprintln(os.Stderr, msg)
		},
	}

	result, err := adoworkitem.Run(opts)
	if err != nil {
		return err
	}

	fmt.Println(result.Output)
	return nil
}
