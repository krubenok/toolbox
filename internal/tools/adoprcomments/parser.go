package adoprcomments

import (
	"fmt"
	"net/url"
	"strings"
)

// ParsedPR contains the extracted components from an Azure DevOps PR URL.
type ParsedPR struct {
	Organization string
	Project      string
	Repository   string
	PRID         string
}

// ParsePRURL parses an Azure DevOps PR URL and extracts its components.
// Supports both dev.azure.com and *.visualstudio.com formats.
func ParsePRURL(rawURL string) (*ParsedPR, error) {
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

// parseDevAzureCom parses dev.azure.com/{org}/{project}/_git/{repo}/pullrequest/{id}
func parseDevAzureCom(parts []string) (*ParsedPR, error) {
	// Expected: [org, project, _git, repo, pullrequest, id]
	if len(parts) < 6 || parts[2] != "_git" || parts[4] != "pullrequest" {
		return nil, fmt.Errorf("PR URL path does not match expected dev.azure.com format")
	}
	return &ParsedPR{
		Organization: parts[0],
		Project:      parts[1],
		Repository:   parts[3],
		PRID:         parts[5],
	}, nil
}

// parseVisualStudioCom parses {org}.visualstudio.com/{project}/_git/{repo}/pullrequest/{id}
func parseVisualStudioCom(parts []string, org string) (*ParsedPR, error) {
	// Expected: [project, _git, repo, pullrequest, id]
	if len(parts) < 5 || parts[1] != "_git" || parts[3] != "pullrequest" {
		return nil, fmt.Errorf("PR URL path does not match expected {org}.visualstudio.com format")
	}
	return &ParsedPR{
		Organization: org,
		Project:      parts[0],
		Repository:   parts[2],
		PRID:         parts[4],
	}, nil
}
