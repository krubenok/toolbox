package adoworkitem

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParsedWorkItem contains the extracted components from an Azure DevOps work item URL.
type ParsedWorkItem struct {
	Organization string
	Project      string
	ID           int
}

// ParseWorkItemURL parses an Azure DevOps work item URL and extracts its components.
// Supports both dev.azure.com and *.visualstudio.com formats.
func ParseWorkItemURL(rawURL string) (*ParsedWorkItem, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", rawURL)
	}

	parts := splitPath(u.Path)
	host := strings.ToLower(u.Host)

	if host == "dev.azure.com" {
		return parseDevAzureCom(parts)
	}

	if strings.HasSuffix(host, ".visualstudio.com") {
		org := strings.Split(host, ".")[0]
		return parseVisualStudioCom(parts, org)
	}

	return nil, fmt.Errorf("unsupported Azure DevOps host: %s", host)
}

// splitPath splits a URL path into decoded segments.
func splitPath(path string) []string {
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p == "" {
			continue
		}
		decoded, err := url.PathUnescape(p)
		if err != nil {
			decoded = p
		}
		parts = append(parts, decoded)
	}
	return parts
}

// parseDevAzureCom parses dev.azure.com/{org}/{project}/_workitems/edit/{id}
func parseDevAzureCom(parts []string) (*ParsedWorkItem, error) {
	// Expected: [org, project, _workitems, edit, id]
	if len(parts) < 5 || parts[2] != "_workitems" || parts[3] != "edit" {
		return nil, fmt.Errorf("work item URL path does not match expected dev.azure.com format")
	}
	id, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid work item id: %s", parts[4])
	}
	return &ParsedWorkItem{
		Organization: parts[0],
		Project:      parts[1],
		ID:           id,
	}, nil
}

// parseVisualStudioCom parses {org}.visualstudio.com/{project}/_workitems/edit/{id}
func parseVisualStudioCom(parts []string, org string) (*ParsedWorkItem, error) {
	// Expected: [project, _workitems, edit, id]
	if len(parts) < 4 || parts[1] != "_workitems" || parts[2] != "edit" {
		return nil, fmt.Errorf("work item URL path does not match expected {org}.visualstudio.com format")
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid work item id: %s", parts[3])
	}
	return &ParsedWorkItem{
		Organization: org,
		Project:      parts[0],
		ID:           id,
	}, nil
}
