package adoprcomments

import "testing"

func TestEmptyStatusFilterSummary(t *testing.T) {
	t.Parallel()

	t.Run("no statuses returns empty", func(t *testing.T) {
		t.Parallel()
		got := EmptyStatusFilterSummary(nil, map[string]int{"fixed": 2}, 0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("no threads returns empty", func(t *testing.T) {
		t.Parallel()
		got := EmptyStatusFilterSummary([]string{"active"}, map[string]int{}, 0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("non-empty filtered results returns empty", func(t *testing.T) {
		t.Parallel()
		got := EmptyStatusFilterSummary([]string{"active"}, map[string]int{"active": 1, "fixed": 2}, 1)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty results with other statuses returns summary", func(t *testing.T) {
		t.Parallel()
		got := EmptyStatusFilterSummary([]string{"active"}, map[string]int{"fixed": 2, "closed": 3}, 0)
		want := "0 comment threads matched status filter (active); 5 comment threads have other statuses: fixed=2, closed=3"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("dedupes and preserves filter order", func(t *testing.T) {
		t.Parallel()
		got := EmptyStatusFilterSummary([]string{"active", "active", "pending"}, map[string]int{"fixed": 1}, 0)
		want := "0 comment threads matched status filter (active,pending); 1 comment thread has other statuses: fixed=1"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}
