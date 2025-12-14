package adoprcomments

import (
	"html"
	"regexp"
	"strings"
)

var (
	reHTMLTag       = regexp.MustCompile(`(?i)</?[a-z][a-z0-9]*[\s>/]`)
	reHTMLBr        = regexp.MustCompile(`(?i)<\s*br\s*/?\s*>`)
	reHTMLCloseP    = regexp.MustCompile(`(?i)</\s*p\s*>`)
	reHTMLCloseDiv  = regexp.MustCompile(`(?i)</\s*div\s*>`)
	reHTMLCloseLi   = regexp.MustCompile(`(?i)</\s*li\s*>`)
	reHTMLCloseTr   = regexp.MustCompile(`(?i)</\s*tr\s*>`)
	reHTMLCloseTh   = regexp.MustCompile(`(?i)</\s*th\s*>`)
	reHTMLCloseTd   = regexp.MustCompile(`(?i)</\s*td\s*>`)
	reHTMLStripTags = regexp.MustCompile(`<[^>]+>`)
	reManyNewlines  = regexp.MustCompile(`\n{3,}`)
)

// SimplifiedThread represents a simplified view of a PR thread.
type SimplifiedThread struct {
	FilePath  string              `json:"filePath,omitempty"`
	LineStart *int                `json:"lineStart,omitempty"`
	LineEnd   *int                `json:"lineEnd,omitempty"`
	Status    string              `json:"status,omitempty"`
	Comments  []SimplifiedComment `json:"comments"`
}

// SimplifiedComment represents a simplified view of a comment.
type SimplifiedComment struct {
	Author    string `json:"author,omitempty"`
	Published string `json:"published,omitempty"`
	Updated   string `json:"updated,omitempty"`
	Type      string `json:"type,omitempty"`
	Content   string `json:"content,omitempty"`
}

// SimplifyThreads converts raw API threads to simplified format.
func SimplifyThreads(threads []Thread, filter *CompiledFilter) []SimplifiedThread {
	result := make([]SimplifiedThread, 0, len(threads))

	for _, thread := range threads {
		simplified := SimplifiedThread{
			Status:   thread.Status,
			Comments: make([]SimplifiedComment, 0, len(thread.Comments)),
		}

		// Get file path from context or properties
		if thread.ThreadContext != nil && thread.ThreadContext.FilePath != "" {
			simplified.FilePath = thread.ThreadContext.FilePath
		} else if thread.Properties != nil && thread.Properties.FilePath != nil {
			simplified.FilePath = thread.Properties.FilePath.Value
		}

		// Get line numbers
		if thread.ThreadContext != nil {
			if thread.ThreadContext.RightFileStart != nil {
				line := thread.ThreadContext.RightFileStart.Line
				simplified.LineStart = &line
			}
			if thread.ThreadContext.RightFileEnd != nil {
				line := thread.ThreadContext.RightFileEnd.Line
				simplified.LineEnd = &line
			}
		}

		// Simplify comments
		for _, comment := range thread.Comments {
			sc := SimplifiedComment{
				Published: comment.PublishedDate,
				Updated:   comment.LastUpdatedDate,
				Type:      comment.CommentType,
			}

			if comment.Author != nil {
				sc.Author = comment.Author.DisplayName
			}

			content := normalizeContent(comment.Content)
			if filter != nil && filter.ShouldFilter(sc.Author) {
				content = filter.Apply(content)
			}
			sc.Content = content

			simplified.Comments = append(simplified.Comments, sc)
		}

		result = append(result, simplified)
	}

	return result
}

// ThreadToMap converts a SimplifiedThread to a map based on output config.
// This allows dynamic field inclusion for TOON output.
func ThreadToMap(t SimplifiedThread, cfg *OutputConfig) map[string]any {
	m := make(map[string]any)

	if shouldInclude(cfg, "filePath", t.FilePath != "") {
		m["filePath"] = t.FilePath
	}
	if shouldInclude(cfg, "lineStart", t.LineStart != nil) {
		m["lineStart"] = t.LineStart
	}
	if shouldInclude(cfg, "lineEnd", t.LineEnd != nil) {
		m["lineEnd"] = t.LineEnd
	}
	if shouldInclude(cfg, "status", t.Status != "") {
		m["status"] = t.Status
	}

	// Always include comments array, but filter comment fields
	comments := make([]map[string]any, 0, len(t.Comments))
	for _, c := range t.Comments {
		comments = append(comments, CommentToMap(c, cfg))
	}
	m["comments"] = comments

	return m
}

// CommentToMap converts a SimplifiedComment to a map based on output config.
func CommentToMap(c SimplifiedComment, cfg *OutputConfig) map[string]any {
	m := make(map[string]any)

	if shouldInclude(cfg, "author", c.Author != "") {
		m["author"] = c.Author
	}
	if shouldInclude(cfg, "published", c.Published != "") {
		m["published"] = c.Published
	}
	if shouldInclude(cfg, "updated", c.Updated != "") {
		m["updated"] = c.Updated
	}
	if shouldInclude(cfg, "type", c.Type != "") {
		m["type"] = c.Type
	}
	if shouldInclude(cfg, "content", c.Content != "") {
		m["content"] = c.Content
	}

	return m
}

// ThreadsToMaps converts a slice of threads to maps for TOON output.
func ThreadsToMaps(threads []SimplifiedThread, cfg *OutputConfig) []map[string]any {
	result := make([]map[string]any, 0, len(threads))
	for _, t := range threads {
		result = append(result, ThreadToMap(t, cfg))
	}
	return result
}

// shouldInclude determines if a field should be included based on config and value.
func shouldInclude(cfg *OutputConfig, field string, hasValue bool) bool {
	mode := cfg.GetFieldMode(field)
	switch mode {
	case FieldModeAlways:
		return true
	case FieldModeNever:
		return false
	case FieldModeNotEmpty:
		fallthrough
	default:
		return hasValue
	}
}

// FilterThreadsByStatus returns threads matching any of the given statuses.
// If statuses is empty, all threads are returned.
func FilterThreadsByStatus(threads []Thread, statuses []string) []Thread {
	if len(statuses) == 0 {
		return threads
	}

	// Build a set for O(1) lookup
	statusSet := make(map[string]bool, len(statuses))
	for _, s := range statuses {
		statusSet[s] = true
	}

	var filtered []Thread
	for _, t := range threads {
		if statusSet[t.Status] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// normalizeContent converts HTML content to plain text/markdown.
func normalizeContent(content string) string {
	if content == "" {
		return ""
	}

	if !looksLikeHTML(content) {
		return strings.TrimSpace(content)
	}

	return htmlToMarkdownish(content)
}

func looksLikeHTML(content string) bool {
	return reHTMLTag.MatchString(content)
}

// htmlToMarkdownish converts HTML to a markdown-like format.
func htmlToMarkdownish(input string) string {
	result := input

	// Convert block elements to newlines
	result = reHTMLBr.ReplaceAllString(result, "\n")
	result = reHTMLCloseP.ReplaceAllString(result, "\n")
	result = reHTMLCloseDiv.ReplaceAllString(result, "\n")
	result = reHTMLCloseLi.ReplaceAllString(result, "\n- ")
	result = reHTMLCloseTr.ReplaceAllString(result, "\n")
	result = reHTMLCloseTh.ReplaceAllString(result, ": ")
	result = reHTMLCloseTd.ReplaceAllString(result, " ")

	// Strip remaining tags
	result = reHTMLStripTags.ReplaceAllString(result, "")

	// Decode HTML entities
	result = html.UnescapeString(result)

	// Clean up whitespace
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	result = strings.Join(lines, "\n")
	result = reManyNewlines.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
