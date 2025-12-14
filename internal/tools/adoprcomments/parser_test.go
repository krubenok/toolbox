package adoprcomments

import "testing"

func TestParsePRURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    *ParsedPR
		wantErr bool
	}{
		{
			name:   "dev.azure.com format",
			rawURL: "https://dev.azure.com/org/project/_git/repo/pullrequest/123",
			want: &ParsedPR{
				Organization: "org",
				Project:      "project",
				Repository:   "repo",
				PRID:         "123",
			},
		},
		{
			name:   "visualstudio.com format",
			rawURL: "https://org.visualstudio.com/project/_git/repo/pullrequest/123",
			want: &ParsedPR{
				Organization: "org",
				Project:      "project",
				Repository:   "repo",
				PRID:         "123",
			},
		},
		{
			name:   "path-unescapes segments",
			rawURL: "https://dev.azure.com/org/my%20project/_git/my%20repo/pullrequest/123",
			want: &ParsedPR{
				Organization: "org",
				Project:      "my project",
				Repository:   "my repo",
				PRID:         "123",
			},
		},
		{
			name:    "unsupported host",
			rawURL:  "https://example.com/org/project/_git/repo/pullrequest/123",
			wantErr: true,
		},
		{
			name:    "invalid path format",
			rawURL:  "https://dev.azure.com/org/project/_git/repo/pullrequests/123",
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

			got, err := ParsePRURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParsePRURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if *got != *tt.want {
				t.Fatalf("ParsePRURL() got %+v, want %+v", got, tt.want)
			}
		})
	}
}
