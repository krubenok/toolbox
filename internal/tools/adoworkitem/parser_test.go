package adoworkitem

import "testing"

func TestParseWorkItemURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    *ParsedWorkItem
		wantErr bool
	}{
		{
			name:   "dev.azure.com format",
			rawURL: "https://dev.azure.com/org/project/_workitems/edit/1144734",
			want: &ParsedWorkItem{
				Organization: "org",
				Project:      "project",
				ID:           1144734,
			},
		},
		{
			name:   "visualstudio.com format",
			rawURL: "https://org.visualstudio.com/project/_workitems/edit/1144734",
			want: &ParsedWorkItem{
				Organization: "org",
				Project:      "project",
				ID:           1144734,
			},
		},
		{
			name:   "path-unescapes segments",
			rawURL: "https://dev.azure.com/org/my%20project/_workitems/edit/1144734",
			want: &ParsedWorkItem{
				Organization: "org",
				Project:      "my project",
				ID:           1144734,
			},
		},
		{
			name:    "unsupported host",
			rawURL:  "https://example.com/org/project/_workitems/edit/1144734",
			wantErr: true,
		},
		{
			name:    "invalid path format",
			rawURL:  "https://dev.azure.com/org/project/_workitems/edits/1144734",
			wantErr: true,
		},
		{
			name:    "invalid id",
			rawURL:  "https://dev.azure.com/org/project/_workitems/edit/not-an-int",
			wantErr: true,
		},
		{
			name:    "invalid url",
			rawURL:  "://not a url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseWorkItemURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseWorkItemURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if *got != *tt.want {
				t.Fatalf("ParseWorkItemURL() got %+v, want %+v", got, tt.want)
			}
		})
	}
}
