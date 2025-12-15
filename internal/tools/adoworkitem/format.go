package adoworkitem

import (
	"html"
	"net/url"
	"regexp"
	"strconv"
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

type SimplifiedWorkItem struct {
	ID          int    `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	UIURL       string `json:"uiUrl,omitempty"`
	Rev         int    `json:"rev,omitempty"`
	Title       string `json:"title,omitempty"`
	Type        string `json:"type,omitempty"`
	State       string `json:"state,omitempty"`
	AssignedTo  string `json:"assignedTo,omitempty"`
	Description string `json:"description,omitempty"`

	Discussion  []SimplifiedComment    `json:"discussion"`
	Children    []SimplifiedChildLink  `json:"children"`
	Attachments []SimplifiedAttachment `json:"attachments"`
}

type SimplifiedComment struct {
	ID       int    `json:"id,omitempty"`
	Author   string `json:"author,omitempty"`
	Created  string `json:"created,omitempty"`
	Modified string `json:"modified,omitempty"`
	Text     string `json:"text,omitempty"`
}

type SimplifiedChildLink struct {
	ID    int    `json:"id,omitempty"`
	URL   string `json:"url,omitempty"`
	UIURL string `json:"uiUrl,omitempty"`
}

type SimplifiedAttachment struct {
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}

func SimplifyWorkItem(parsed *ParsedWorkItem, wi WorkItemResponse, comments []WorkItemComment, baseURL string) SimplifiedWorkItem {
	s := SimplifiedWorkItem{
		ID:    wi.ID,
		URL:   wi.URL,
		UIURL: UIWorkItemURL(baseURL, parsed.Organization, parsed.Project, wi.ID),
		Rev:   wi.Rev,

		Title:      getStringField(wi.Fields, "System.Title"),
		Type:       getStringField(wi.Fields, "System.WorkItemType"),
		State:      getStringField(wi.Fields, "System.State"),
		AssignedTo: getIdentityDisplayName(wi.Fields, "System.AssignedTo"),

		Description: normalizeContent(getStringField(wi.Fields, "System.Description")),
		Discussion:  make([]SimplifiedComment, 0, len(comments)),
		Children:    extractChildren(baseURL, parsed.Organization, parsed.Project, wi.Relations),
		Attachments: extractAttachments(baseURL, parsed.Organization, parsed.Project, wi.Relations),
	}

	for _, c := range comments {
		sc := SimplifiedComment{
			ID:       c.ID,
			Created:  c.CreatedDate,
			Modified: c.ModifiedDate,
			Text:     normalizeContent(c.Text),
		}
		if c.CreatedBy != nil {
			sc.Author = c.CreatedBy.DisplayName
		}
		s.Discussion = append(s.Discussion, sc)
	}

	return s
}

func getStringField(fields map[string]any, key string) string {
	if fields == nil {
		return ""
	}
	v, ok := fields[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func getIdentityDisplayName(fields map[string]any, key string) string {
	if fields == nil {
		return ""
	}
	v, ok := fields[key]
	if !ok || v == nil {
		return ""
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if s, ok := m["displayName"].(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func extractChildren(baseURL, org, project string, relations []WorkItemRelation) []SimplifiedChildLink {
	var children []SimplifiedChildLink
	for _, r := range relations {
		if !isChildRelation(r.Rel) {
			continue
		}
		id := extractWorkItemIDFromRelationURL(r.URL)
		if id == 0 {
			continue
		}
		children = append(children, SimplifiedChildLink{
			ID:    id,
			URL:   r.URL,
			UIURL: UIWorkItemURL(baseURL, org, project, id),
		})
	}
	if children == nil {
		return []SimplifiedChildLink{}
	}
	return children
}

func extractAttachments(baseURL, org, project string, relations []WorkItemRelation) []SimplifiedAttachment {
	var attachments []SimplifiedAttachment
	for _, r := range relations {
		if r.Rel != "AttachedFile" {
			continue
		}
		name := getStringAttr(r.Attributes, "name")
		downloadURL := attachmentDownloadURL(baseURL, org, project, r.URL, name)
		attachments = append(attachments, SimplifiedAttachment{
			Name:        name,
			URL:         r.URL,
			DownloadURL: downloadURL,
		})
	}
	if attachments == nil {
		return []SimplifiedAttachment{}
	}
	return attachments
}

func isChildRelation(rel string) bool {
	return rel == "System.LinkTypes.Hierarchy-Forward"
}

func getStringAttr(attrs map[string]any, key string) string {
	if attrs == nil {
		return ""
	}
	if v, ok := attrs[key]; ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func extractWorkItemIDFromRelationURL(relURL string) int {
	if relURL == "" {
		return 0
	}

	// vstfs:///WorkItemTracking/WorkItem/{id}
	if strings.HasPrefix(relURL, "vstfs:///") {
		parts := strings.Split(relURL, "/")
		if len(parts) == 0 {
			return 0
		}
		last := parts[len(parts)-1]
		id, _ := strconv.Atoi(last)
		return id
	}

	u, err := url.Parse(relURL)
	if err != nil {
		return 0
	}
	parts := splitPath(u.Path)
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.EqualFold(parts[i], "workitems") || strings.EqualFold(parts[i], "workItems") {
			if i+1 < len(parts) {
				id, _ := strconv.Atoi(parts[i+1])
				return id
			}
		}
	}

	// Fallback: last segment as id.
	if len(parts) > 0 {
		id, _ := strconv.Atoi(parts[len(parts)-1])
		return id
	}
	return 0
}

func attachmentDownloadURL(baseURL, org, project, relURL, name string) string {
	if relURL == "" {
		return ""
	}
	// Expect: vstfs:///WorkItemTracking/Attachment/{guid}
	if !strings.Contains(relURL, "/Attachment/") {
		return ""
	}
	guid := relURL[strings.LastIndex(relURL, "/")+1:]
	if guid == "" {
		return ""
	}

	u := baseURL + "/" + url.PathEscape(org) + "/" + url.PathEscape(project) + "/_apis/wit/attachments/" + url.PathEscape(guid)
	if name != "" {
		u += "?fileName=" + url.QueryEscape(name) + "&api-version=7.1-preview.3"
	} else {
		u += "?api-version=7.1-preview.3"
	}
	return u
}

// WorkItemToMap converts a SimplifiedWorkItem to a map based on output config.
func WorkItemToMap(w SimplifiedWorkItem, cfg *OutputConfig) map[string]any {
	m := make(map[string]any)

	if shouldInclude(cfg, "id", w.ID != 0) {
		m["id"] = w.ID
	}
	if shouldInclude(cfg, "url", w.URL != "") {
		m["url"] = w.URL
	}
	if shouldInclude(cfg, "uiUrl", w.UIURL != "") {
		m["uiUrl"] = w.UIURL
	}
	if shouldInclude(cfg, "rev", w.Rev != 0) {
		m["rev"] = w.Rev
	}
	if shouldInclude(cfg, "title", w.Title != "") {
		m["title"] = w.Title
	}
	if shouldInclude(cfg, "type", w.Type != "") {
		m["type"] = w.Type
	}
	if shouldInclude(cfg, "state", w.State != "") {
		m["state"] = w.State
	}
	if shouldInclude(cfg, "assignedTo", w.AssignedTo != "") {
		m["assignedTo"] = w.AssignedTo
	}
	if shouldInclude(cfg, "description", w.Description != "") {
		m["description"] = w.Description
	}

	if shouldInclude(cfg, "discussion", len(w.Discussion) > 0) {
		comments := make([]map[string]any, 0, len(w.Discussion))
		for _, c := range w.Discussion {
			comments = append(comments, CommentToMap(c, cfg))
		}
		m["discussion"] = comments
	}

	if shouldInclude(cfg, "children", len(w.Children) > 0) {
		children := make([]map[string]any, 0, len(w.Children))
		for _, c := range w.Children {
			children = append(children, ChildToMap(c, cfg))
		}
		m["children"] = children
	}

	if shouldInclude(cfg, "attachments", len(w.Attachments) > 0) {
		attachments := make([]map[string]any, 0, len(w.Attachments))
		for _, a := range w.Attachments {
			attachments = append(attachments, AttachmentToMap(a, cfg))
		}
		m["attachments"] = attachments
	}

	return m
}

func CommentToMap(c SimplifiedComment, cfg *OutputConfig) map[string]any {
	m := make(map[string]any)
	if shouldInclude(cfg, "commentId", c.ID != 0) {
		m["commentId"] = c.ID
	}
	if shouldInclude(cfg, "commentAuthor", c.Author != "") {
		m["commentAuthor"] = c.Author
	}
	if shouldInclude(cfg, "commentCreated", c.Created != "") {
		m["commentCreated"] = c.Created
	}
	if shouldInclude(cfg, "commentModified", c.Modified != "") {
		m["commentModified"] = c.Modified
	}
	if shouldInclude(cfg, "commentText", c.Text != "") {
		m["commentText"] = c.Text
	}
	return m
}

func ChildToMap(c SimplifiedChildLink, cfg *OutputConfig) map[string]any {
	m := make(map[string]any)
	if shouldInclude(cfg, "childId", c.ID != 0) {
		m["childId"] = c.ID
	}
	if shouldInclude(cfg, "childUrl", c.URL != "") {
		m["childUrl"] = c.URL
	}
	if shouldInclude(cfg, "childUiUrl", c.UIURL != "") {
		m["childUiUrl"] = c.UIURL
	}
	return m
}

func AttachmentToMap(a SimplifiedAttachment, cfg *OutputConfig) map[string]any {
	m := make(map[string]any)
	if shouldInclude(cfg, "attachmentName", a.Name != "") {
		m["attachmentName"] = a.Name
	}
	if shouldInclude(cfg, "attachmentUrl", a.URL != "") {
		m["attachmentUrl"] = a.URL
	}
	if shouldInclude(cfg, "attachmentDownloadUrl", a.DownloadURL != "") {
		m["attachmentDownloadUrl"] = a.DownloadURL
	}
	return m
}

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

func htmlToMarkdownish(input string) string {
	result := input

	result = reHTMLBr.ReplaceAllString(result, "\n")
	result = reHTMLCloseP.ReplaceAllString(result, "\n")
	result = reHTMLCloseDiv.ReplaceAllString(result, "\n")
	result = reHTMLCloseLi.ReplaceAllString(result, "\n- ")
	result = reHTMLCloseTr.ReplaceAllString(result, "\n")
	result = reHTMLCloseTh.ReplaceAllString(result, ": ")
	result = reHTMLCloseTd.ReplaceAllString(result, " ")

	result = reHTMLStripTags.ReplaceAllString(result, "")
	result = html.UnescapeString(result)

	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	result = strings.Join(lines, "\n")
	result = reManyNewlines.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
