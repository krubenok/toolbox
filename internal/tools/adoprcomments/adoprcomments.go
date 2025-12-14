package adoprcomments

import (
	"encoding/json"

	"github.com/kyrubeno/toolbox/internal/auth"
	toon "github.com/toon-format/toon-go"
)

// Options configures the PR comments fetcher.
type Options struct {
	PRURL      string
	ActiveOnly bool
	OutputJSON bool // Output JSON instead of toon
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

	// Load and compile filter (unless disabled)
	var filter *CompiledFilter
	if !opts.NoFilter {
		cfg, err := LoadConfig()
		if err != nil {
			if opts.Debug && opts.DebugLog != nil {
				opts.DebugLog("Warning: failed to load config: " + err.Error())
			}
			cfg = &Config{Filter: DefaultFilterConfig()}
		}

		filter, err = cfg.Filter.Compile()
		if err != nil {
			if opts.Debug && opts.DebugLog != nil {
				opts.DebugLog("Warning: failed to compile filter: " + err.Error())
			}
		}
	}

	// Create client and fetch threads
	client := NewClient(azAuth, opts.Debug, opts.DebugLog)
	threads, err := client.FetchThreads(parsed)
	if err != nil {
		return nil, err
	}

	// Filter if needed
	if opts.ActiveOnly {
		threads = FilterActiveThreads(threads)
	}

	// Simplify threads
	simplified := SimplifyThreads(threads, filter)

	// Serialize output
	var output string
	if opts.OutputJSON {
		jsonBytes, err := json.MarshalIndent(simplified, "", "  ")
		if err != nil {
			return nil, err
		}
		output = string(jsonBytes)
	} else {
		toonStr, err := toon.MarshalString(simplified)
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
