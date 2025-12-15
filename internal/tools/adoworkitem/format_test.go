package adoworkitem

import (
	"testing"
)

func TestNormalizeContent(t *testing.T) {
	t.Parallel()

	t.Run("non-html text with less-than", func(t *testing.T) {
		got := normalizeContent("a < b")
		if got != "a < b" {
			t.Fatalf("got %q, want %q", got, "a < b")
		}
	})

	t.Run("html content is stripped and unescaped", func(t *testing.T) {
		got := normalizeContent("<div>Hello<br/>World &amp; friends</div>")
		want := "Hello\nWorld & friends"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func TestExtractWorkItemIDFromRelationURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want int
	}{
		{
			name: "vstfs url",
			in:   "vstfs:///WorkItemTracking/WorkItem/12345",
			want: 12345,
		},
		{
			name: "api url",
			in:   "https://dev.azure.com/org/_apis/wit/workItems/12345",
			want: 12345,
		},
		{
			name: "api url with other path",
			in:   "https://dev.azure.com/org/project/_apis/wit/workItems/12345",
			want: 12345,
		},
		{
			name: "empty",
			in:   "",
			want: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractWorkItemIDFromRelationURL(tt.in)
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}
