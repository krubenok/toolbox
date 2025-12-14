package adoprcomments

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyrubeno/toolbox/internal/auth"
	toon "github.com/toon-format/toon-go"
)

// Options configures the PR comments fetcher.
type Options struct {
	Ctx        context.Context
	PRURL      string
	Statuses   []string // Filter to these statuses (empty = use config default, which may also be empty for all)
	OutputJSON bool     // Output JSON instead of toon
	Debug      bool
	NoFilter   bool // Disable content filtering
	DebugLog   func(string)
}

// Result contains the output from fetching PR comments.
type Result struct {
	Threads []SimplifiedThread
	Output  string // Formatted output (toon or JSON)
}

// Run fetches and processes PR comments from Azure DevOps.
func Run(opts Options) (*Result, error) {
	ctx := opts.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Parse the PR URL
	parsed, err := ParsePRURL(opts.PRURL)
	if err != nil {
		return nil, err
	}

	// Get authentication
	azAuth, err := auth.GetAzureAuth()
	if err != nil {
		return nil, err
	}

	if opts.Debug && opts.DebugLog != nil {
		opts.DebugLog("Auth: " + azAuth.Scheme)
	}

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Compile filter (unless disabled)
	var filter *CompiledFilter
	if !opts.NoFilter {
		filter, err = cfg.Filter.Compile()
		if err != nil {
			return nil, fmt.Errorf("compile filter config: %w", err)
		}
	}

	// Create client and fetch threads
	client := NewClient(azAuth, opts.Debug, opts.DebugLog)
	threads, err := client.FetchThreads(ctx, parsed)
	if err != nil {
		return nil, err
	}

	// Filter by status
	// If CLI provided statuses, use those; otherwise use config default
	statuses := opts.Statuses
	if len(statuses) == 0 && cfg.Status != nil {
		statuses = cfg.Status.Include
	}
	threads = FilterThreadsByStatus(threads, statuses)

	// Simplify threads
	simplified := SimplifyThreads(threads, filter)

	// Serialize output
	var output string
	if opts.OutputJSON {
		// JSON output uses structs with omitempty tags
		jsonBytes, err := json.MarshalIndent(simplified, "", "  ")
		if err != nil {
			return nil, err
		}
		output = string(jsonBytes)
	} else {
		// TOON output uses maps with configurable field inclusion
		maps := ThreadsToMaps(simplified, cfg.Output)
		toonStr, err := toon.MarshalString(maps)
		if err != nil {
			// Fall back to JSON if toon fails
			if opts.Debug && opts.DebugLog != nil {
				opts.DebugLog("Warning: toon encoding failed, falling back to JSON: " + err.Error())
			}
			jsonBytes, err := json.MarshalIndent(simplified, "", "  ")
			if err != nil {
				return nil, err
			}
			output = string(jsonBytes)
		} else {
			output = toonStr
		}
	}

	return &Result{
		Threads: simplified,
		Output:  output,
	}, nil
}
