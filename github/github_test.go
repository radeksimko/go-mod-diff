package github

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGitHubGetCommitSHA(t *testing.T) {
	ts := githubApiMockServer([]*githubResponse{
		{
			URI:         "/repos/hashicorp/terraform/commits/v0.11.11",
			ContentType: "application/json; charset=utf-8",
			Body: `{
  "sha": "ac4fff416318bf0915a0ab80e062a99ef3724334",
  "node_id": "MDY6Q29tbWl0MTc3MjgxNjQ6YWM0ZmZmNDE2MzE4YmYwOTE1YTBhYjgwZTA2MmE5OWVmMzcyNDMzNA==",
  "commit": {
    "message": "v0.11.11",
    "tree": {
      "sha": "4506f028b93d44c18af45bac8ed761256f733a01",
      "url": "https://api.github.com/repos/hashicorp/terraform/git/trees/4506f028b93d44c18af45bac8ed761256f733a01"
    },
    "url": "https://api.github.com/repos/hashicorp/terraform/git/commits/ac4fff416318bf0915a0ab80e062a99ef3724334"
  },
  "url": "https://api.github.com/repos/hashicorp/terraform/commits/ac4fff416318bf0915a0ab80e062a99ef3724334",
  "html_url": "https://github.com/hashicorp/terraform/commit/ac4fff416318bf0915a0ab80e062a99ef3724334",
  "comments_url": "https://api.github.com/repos/hashicorp/terraform/commits/ac4fff416318bf0915a0ab80e062a99ef3724334/comments"
}`,
		},
	})
	defer ts.Close()

	gh := NewGitHubWithURL(ts.URL)
	sha, err := gh.GetCommitSHA(&Repository{"hashicorp", "terraform"}, "v0.11.11")
	if err != nil {
		t.Fatal(err)
	}

	expectedSHA := "ac4fff416318bf0915a0ab80e062a99ef3724334"
	if sha != expectedSHA {
		t.Fatalf("Expected: %q, given: %q", expectedSHA, sha)
	}
}

func TestParseRepositoryURL(t *testing.T) {
	testCases := []struct {
		rawURL       string
		expectedErr  bool
		expectedRepo *Repository
	}{
		{
			rawURL:       "https://github.com/hashicorp/terraform",
			expectedErr:  false,
			expectedRepo: &Repository{"hashicorp", "terraform"},
		},
		{ // no protocol
			rawURL:       "github.com/hashicorp/terraform",
			expectedErr:  false,
			expectedRepo: &Repository{"hashicorp", "terraform"},
		},
		{
			rawURL:      "https://ghe.engineering/hashicorp/terraform",
			expectedErr: true,
		},
		{
			rawURL:      "https://random-hostname",
			expectedErr: true,
		},
		{
			rawURL:      "https://random/path",
			expectedErr: true,
		},
		{
			rawURL:      "https://another/random/path",
			expectedErr: true,
		},
		{
			rawURL:      "https://github.com/just-org",
			expectedErr: true,
		},
		{
			rawURL:      "https://github.com/org/repo/something-else",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		repo, err := ParseRepositoryURL(tc.rawURL)
		if tc.expectedErr {
			if err == nil {
				t.Fatalf("Expected %q to return error, none given.", tc.rawURL)
			}
			continue
		}

		if err != nil {
			t.Fatalf("Parsing %q failed: %s", tc.rawURL, err)
		}

		if !reflect.DeepEqual(*tc.expectedRepo, *repo) {
			t.Fatalf("Expected %q, given: %q", *tc.expectedRepo, *repo)
		}
	}
}

func TestTreeURL(t *testing.T) {
	testCases := []struct {
		repo        *Repository
		ref         string
		expectedUrl string
	}{
		{
			&Repository{"hashicorp", "terraform"},
			"v0.11",
			"https://github.com/hashicorp/terraform/tree/v0.11",
		},
		{
			&Repository{"hashicorp", "terraform"},
			"f9b62cb5fef70e9f24f6c421f8840b999d2b0bed",
			"https://github.com/hashicorp/terraform/tree/f9b62cb5fef70e9f24f6c421f8840b999d2b0bed",
		},
	}

	for _, tc := range testCases {
		url := TreeURL(tc.repo, tc.ref)
		if url != tc.expectedUrl {
			t.Fatalf("Expected %q, given: %q", tc.expectedUrl, url)
		}
	}
}

func githubApiMockServer(reponses []*githubResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Mock server received request to %q", r.RequestURI)
		for _, resp := range reponses {
			if r.RequestURI == resp.URI {
				w.Header().Set("Content-Type", resp.ContentType)
				fmt.Fprintln(w, resp.Body)
				w.WriteHeader(200)
				return
			}
		}
		w.WriteHeader(400)
	}))
}

type githubResponse struct {
	URI         string
	ContentType string
	Body        string
}
