package adoprcomments

import (
	"os"
	"regexp"
	"strings"

	"github.com/kyrubeno/toolbox/internal/config"
)

const configFile = "ado-pr-comments.json"

// Config holds all configuration for the ado-pr-comments tool.
type Config struct {
	Filter *FilterConfig `json:"filter,omitempty"`
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

// CompiledFilter holds compiled regex patterns for efficient filtering.
type CompiledFilter struct {
	cutPatterns    []*regexp.Regexp
	scrubPatterns  []*regexp.Regexp
	authorPatterns []*regexp.Regexp
}

// DefaultFilterConfig returns the default filter config.
// These are generic patterns for common PR comment noise, not specific to any bot.
func DefaultFilterConfig() *FilterConfig {
	return &FilterConfig{
		// Empty by default - no content is cut or scrubbed unless configured
		CutPatterns:   []string{},
		ScrubPatterns: []string{},
		// Empty means filter applies to all authors
		AuthorPatterns: []string{},
	}
}

// LoadConfig loads the ado-pr-comments config from ~/.toolbox/ado-pr-comments.json.
// Falls back to defaults if file doesn't exist.
func LoadConfig() (*Config, error) {
	var cfg Config
	err := config.Load(configFile, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Filter: DefaultFilterConfig()}, nil
		}
		return nil, err
	}

	// Use defaults if filter not specified
	if cfg.Filter == nil {
		cfg.Filter = DefaultFilterConfig()
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
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
