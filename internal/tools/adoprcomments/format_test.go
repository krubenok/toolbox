package adoprcomments

import "testing"

func TestFilterThreadsByStatus(t *testing.T) {
	t.Parallel()

	threads := []Thread{
		{ID: 1, Status: "active"},
		{ID: 2, Status: "fixed"},
		{ID: 3, Status: "active"},
	}

	t.Run("empty statuses returns all", func(t *testing.T) {
		got := FilterThreadsByStatus(threads, nil)
		if len(got) != len(threads) {
			t.Fatalf("len = %d, want %d", len(got), len(threads))
		}
	})

	t.Run("filters to matching statuses", func(t *testing.T) {
		got := FilterThreadsByStatus(threads, []string{"active"})
		if len(got) != 2 {
			t.Fatalf("len = %d, want %d", len(got), 2)
		}
		if got[0].ID != 1 || got[1].ID != 3 {
			t.Fatalf("unexpected thread IDs: got %v, %v", got[0].ID, got[1].ID)
		}
	})
}

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
