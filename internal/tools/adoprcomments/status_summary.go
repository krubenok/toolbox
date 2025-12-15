package adoprcomments

import (
	"fmt"
	"sort"
	"strings"
)

var preferredStatusOrder = []string{
	"active",
	"pending",
	"fixed",
	"wontFix",
	"byDesign",
	"closed",
}

// CountThreadsByStatus returns a map of thread status -> count.
func CountThreadsByStatus(threads []Thread) map[string]int {
	counts := make(map[string]int)
	for _, t := range threads {
		counts[t.Status]++
	}
	return counts
}

// EmptyStatusFilterSummary returns a summary line when a status filter yields 0 results but other statuses exist.
// Returns an empty string when no summary is applicable.
func EmptyStatusFilterSummary(statuses []string, allStatusCounts map[string]int, filteredCount int) string {
	if len(statuses) == 0 {
		return ""
	}

	total := 0
	for _, n := range allStatusCounts {
		total += n
	}
	if total == 0 {
		return ""
	}

	if filteredCount != 0 {
		return ""
	}

	statusSet := make(map[string]struct{}, len(statuses))
	statuses = dedupePreserveOrder(statuses)
	for _, s := range statuses {
		statusSet[s] = struct{}{}
	}

	otherCounts := make(map[string]int)
	otherTotal := 0
	for status, n := range allStatusCounts {
		if n == 0 {
			continue
		}
		if _, ok := statusSet[status]; ok {
			continue
		}
		otherCounts[status] = n
		otherTotal += n
	}
	if otherTotal == 0 {
		return ""
	}

	var filterLabel string
	if len(statuses) == 1 {
		filterLabel = statuses[0]
	} else {
		filterLabel = strings.Join(statuses, ",")
	}

	parts := make([]string, 0, len(otherCounts))
	for _, status := range orderedStatusKeys(otherCounts) {
		parts = append(parts, fmt.Sprintf("%s=%d", status, otherCounts[status]))
	}

	otherLabel := "comment threads"
	otherVerb := "have"
	if otherTotal == 1 {
		otherLabel = "comment thread"
		otherVerb = "has"
	}

	return fmt.Sprintf(
		"0 comment threads matched status filter (%s); %d %s %s other statuses: %s",
		filterLabel,
		otherTotal,
		otherLabel,
		otherVerb,
		strings.Join(parts, ", "),
	)
}

func orderedStatusKeys(counts map[string]int) []string {
	seen := make(map[string]struct{}, len(counts))
	var keys []string

	for _, status := range preferredStatusOrder {
		if counts[status] <= 0 {
			continue
		}
		keys = append(keys, status)
		seen[status] = struct{}{}
	}

	var remaining []string
	for status, n := range counts {
		if n <= 0 {
			continue
		}
		if _, ok := seen[status]; ok {
			continue
		}
		remaining = append(remaining, status)
	}
	sort.Strings(remaining)

	return append(keys, remaining...)
}

func dedupePreserveOrder(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
