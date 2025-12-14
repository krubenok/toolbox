package adoprcomments

import (
	"os"
	"regexp"
	"strings"

	"github.com/kyrubeno/toolbox/internal/config"
)

const configFile = "ado-pr-comments.json"

// FieldMode specifies when a field should be included in output.
type FieldMode string

const (
	FieldModeAlways   FieldMode = "always"   // Always include the field
	FieldModeNotEmpty FieldMode = "notEmpty" // Only include if not empty/null (default)
	FieldModeNever    FieldMode = "never"    // Never include the field
)

// Config holds all configuration for the ado-pr-comments tool.
type Config struct {
	Filter *FilterConfig `json:"filter,omitempty"`
	Output *OutputConfig `json:"output,omitempty"`
	Status *StatusConfig `json:"status,omitempty"`
}

// StatusConfig controls which thread statuses are included in output.
type StatusConfig struct {
	// Include is a list of statuses to include (e.g., "active", "fixed", "closed", "byDesign", "pending", "wontFix").
	// Empty means include all statuses.
	Include []string `json:"include,omitempty"`
}

// FilterConfig defines patterns to strip from PR comments.
type FilterConfig struct {
	// CutPatterns - content after first match is removed
	CutPatterns []string `json:"cutPatterns"`
	// ScrubPatterns - all matches are removed
	ScrubPatterns []string `json:"scrubPatterns"`
	// AuthorPatterns - only apply filters to comments from matching authors (empty = all)
	AuthorPatterns []string `json:"authorPatterns"`
}

// OutputConfig controls which fields are included in output.
type OutputConfig struct {
	// Thread fields
	FilePath  FieldMode `json:"filePath,omitempty"`
	LineStart FieldMode `json:"lineStart,omitempty"`
	LineEnd   FieldMode `json:"lineEnd,omitempty"`
	Status    FieldMode `json:"status,omitempty"`

	// Comment fields
	Author    FieldMode `json:"author,omitempty"`
	Published FieldMode `json:"published,omitempty"`
	Updated   FieldMode `json:"updated,omitempty"`
	Type      FieldMode `json:"type,omitempty"`
	Content   FieldMode `json:"content,omitempty"`
}

// CompiledFilter holds compiled regex patterns for efficient filtering.
type CompiledFilter struct {
	cutPatterns    []*regexp.Regexp
	scrubPatterns  []*regexp.Regexp
	authorPatterns []*regexp.Regexp
}

// DefaultFilterConfig returns the default filter config.
func DefaultFilterConfig() *FilterConfig {
	return &FilterConfig{
		CutPatterns:    []string{},
		ScrubPatterns:  []string{},
		AuthorPatterns: []string{},
	}
}

// DefaultStatusConfig returns the default status config.
// Empty Include means all statuses are included.
func DefaultStatusConfig() *StatusConfig {
	return &StatusConfig{
		Include: []string{},
	}
}

// DefaultOutputConfig returns the default output config.
// All fields default to "notEmpty".
func DefaultOutputConfig() *OutputConfig {
	return &OutputConfig{
		FilePath:  FieldModeNotEmpty,
		LineStart: FieldModeNotEmpty,
		LineEnd:   FieldModeNotEmpty,
		Status:    FieldModeNotEmpty,
		Author:    FieldModeNotEmpty,
		Published: FieldModeNotEmpty,
		Updated:   FieldModeNotEmpty,
		Type:      FieldModeNotEmpty,
		Content:   FieldModeNotEmpty,
	}
}

// GetFieldMode returns the mode for a field, defaulting to notEmpty if not set.
func (oc *OutputConfig) GetFieldMode(field string) FieldMode {
	if oc == nil {
		return FieldModeNotEmpty
	}

	var mode FieldMode
	switch field {
	case "filePath":
		mode = oc.FilePath
	case "lineStart":
		mode = oc.LineStart
	case "lineEnd":
		mode = oc.LineEnd
	case "status":
		mode = oc.Status
	case "author":
		mode = oc.Author
	case "published":
		mode = oc.Published
	case "updated":
		mode = oc.Updated
	case "type":
		mode = oc.Type
	case "content":
		mode = oc.Content
	}

	if mode == "" {
		return FieldModeNotEmpty
	}
	return mode
}

// LoadConfig loads the ado-pr-comments config from ~/.toolbox/ado-pr-comments.json.
// Falls back to defaults if file doesn't exist.
func LoadConfig() (*Config, error) {
	var cfg Config
	err := config.Load(configFile, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Filter: DefaultFilterConfig(),
				Output: DefaultOutputConfig(),
			}, nil
		}
		return nil, err
	}

	// Use defaults if not specified
	if cfg.Filter == nil {
		cfg.Filter = DefaultFilterConfig()
	}
	if cfg.Output == nil {
		cfg.Output = DefaultOutputConfig()
	}
	if cfg.Status == nil {
		cfg.Status = DefaultStatusConfig()
	}

	return &cfg, nil
}

// Compile compiles the filter patterns into regex.
func (fc *FilterConfig) Compile() (*CompiledFilter, error) {
	cf := &CompiledFilter{}

	for _, p := range fc.CutPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		cf.cutPatterns = append(cf.cutPatterns, re)
	}

	for _, p := range fc.ScrubPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		cf.scrubPatterns = append(cf.scrubPatterns, re)
	}

	for _, p := range fc.AuthorPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		cf.authorPatterns = append(cf.authorPatterns, re)
	}

	return cf, nil
}

// ShouldFilter returns true if the author matches the filter patterns.
func (cf *CompiledFilter) ShouldFilter(author string) bool {
	if len(cf.authorPatterns) == 0 {
		return true
	}
	for _, re := range cf.authorPatterns {
		if re.MatchString(author) {
			return true
		}
	}
	return false
}

// Apply applies the filter to the given text.
func (cf *CompiledFilter) Apply(text string) string {
	result := text

	// Cut at first matching pattern
	for _, re := range cf.cutPatterns {
		loc := re.FindStringIndex(result)
		if loc != nil {
			result = result[:loc[0]]
			break
		}
	}

	// Scrub all matching patterns
	for _, re := range cf.scrubPatterns {
		result = re.ReplaceAllString(result, "")
	}

	// Clean up extra newlines
	result = reManyNewlines.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
