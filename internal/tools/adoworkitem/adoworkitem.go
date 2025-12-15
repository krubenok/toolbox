package adoworkitem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krubenok/toolbox/internal/auth"
	toon "github.com/toon-format/toon-go"
)

type Options struct {
	Ctx         context.Context
	WorkItemURL string

	IncludeDescription bool
	IncludeDiscussion  bool
	IncludeChildren    bool
	IncludeAttachments bool
	MaxComments        int

	OutputJSON bool
	Debug      bool
	DebugLog   func(string)
}

type Result struct {
	WorkItem SimplifiedWorkItem
	Output   string
}

func Run(opts Options) (*Result, error) {
	ctx := opts.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	parsed, err := ParseWorkItemURL(opts.WorkItemURL)
	if err != nil {
		return nil, err
	}

	azAuth, err := auth.GetAzureAuth()
	if err != nil {
		return nil, err
	}
	if opts.Debug && opts.DebugLog != nil {
		opts.DebugLog("Auth: " + azAuth.Scheme)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	client := NewClient(azAuth, opts.Debug, opts.DebugLog)

	wi, err := client.FetchWorkItem(ctx, parsed)
	if err != nil {
		return nil, err
	}

	var comments []WorkItemComment
	if opts.IncludeDiscussion {
		comments, err = client.FetchAllComments(ctx, parsed, opts.MaxComments)
		if err != nil {
			return nil, err
		}
	}

	simplified := SimplifyWorkItem(parsed, *wi, comments, defaultBaseURL)
	if !opts.IncludeDescription {
		simplified.Description = ""
	}
	if !opts.IncludeChildren {
		simplified.Children = []SimplifiedChildLink{}
	}
	if !opts.IncludeAttachments {
		simplified.Attachments = []SimplifiedAttachment{}
	}
	if !opts.IncludeDiscussion {
		simplified.Discussion = []SimplifiedComment{}
	}

	var output string
	if opts.OutputJSON {
		b, err := json.MarshalIndent(simplified, "", "  ")
		if err != nil {
			return nil, err
		}
		output = string(b)
	} else {
		m := WorkItemToMap(simplified, cfg.Output)
		toonStr, err := toon.MarshalString(m)
		if err != nil {
			if opts.Debug && opts.DebugLog != nil {
				opts.DebugLog("Warning: toon encoding failed, falling back to JSON: " + err.Error())
			}
			b, err := json.MarshalIndent(simplified, "", "  ")
			if err != nil {
				return nil, err
			}
			output = string(b)
		} else {
			output = toonStr
		}
	}

	return &Result{
		WorkItem: simplified,
		Output:   output,
	}, nil
}
