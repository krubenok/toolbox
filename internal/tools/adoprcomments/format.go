package adoprcomments

import (
	"regexp"
	"strings"
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

// FilterActiveThreads returns only threads with status "active".
func FilterActiveThreads(threads []Thread) []Thread {
	var active []Thread
	for _, t := range threads {
		if t.Status == "active" {
			active = append(active, t)
		}
	}
	return active
}

// normalizeContent converts HTML content to plain text/markdown.
func normalizeContent(content string) string {
	if content == "" {
		return ""
	}

	// Check if content contains HTML
	if !strings.Contains(content, "<") {
		return strings.TrimSpace(content)
	}

	return htmlToMarkdownish(content)
}

// htmlToMarkdownish converts HTML to a markdown-like format.
func htmlToMarkdownish(input string) string {
	result := input

	// Convert block elements to newlines
	result = regexp.MustCompile(`(?i)<\s*br\s*/?\s*>`).ReplaceAllString(result, "\n")
	result = regexp.MustCompile(`(?i)</\s*p\s*>`).ReplaceAllString(result, "\n")
	result = regexp.MustCompile(`(?i)</\s*div\s*>`).ReplaceAllString(result, "\n")
	result = regexp.MustCompile(`(?i)</\s*li\s*>`).ReplaceAllString(result, "\n- ")
	result = regexp.MustCompile(`(?i)</\s*tr\s*>`).ReplaceAllString(result, "\n")
	result = regexp.MustCompile(`(?i)</\s*th\s*>`).ReplaceAllString(result, ": ")
	result = regexp.MustCompile(`(?i)</\s*td\s*>`).ReplaceAllString(result, " ")

	// Strip remaining tags
	result = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(result, "")

	// Decode HTML entities
	result = decodeEntities(result)

	// Clean up whitespace
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	result = strings.Join(lines, "\n")
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}

// decodeEntities decodes common HTML entities.
func decodeEntities(input string) string {
	replacements := map[string]string{
		"&lt;":   "<",
		"&gt;":   ">",
		"&amp;":  "&",
		"&quot;": `"`,
		"&#39;":  "'",
	}

	result := input
	for entity, char := range replacements {
		result = strings.ReplaceAll(result, entity, char)
	}
	return result
}
