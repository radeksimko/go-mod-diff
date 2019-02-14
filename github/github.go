package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	githubSDK "github.com/google/go-github/v22/github"
	"golang.org/x/oauth2"
)

const ghHostname = "github.com"

type Repository struct {
	Owner string
	Name  string
}

type GitHub struct {
	ctx    context.Context
	client *githubSDK.Client
}

func (gh *GitHub) GetCommitSHA(r *Repository, ref string) (string, error) {
	rc, _, err := gh.client.Repositories.GetCommit(gh.ctx, r.Owner, r.Name, ref)
	if err != nil {
		return "", err
	}
	return *rc.SHA, nil
}

func NewGitHub() *GitHub {
	return &GitHub{
		ctx:    context.Background(),
		client: githubSDK.NewClient(nil),
	}
}

func NewGitHubWithToken(token string) *GitHub {
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	}))

	return &GitHub{
		ctx:    ctx,
		client: githubSDK.NewClient(tc),
	}
}

func NewGitHubWithURL(rawUrl string) *GitHub {
	ghClient := githubSDK.NewClient(nil)

	customURL, _ := url.Parse(rawUrl)

	if customURL.EscapedPath() == "" {
		customURL.Path = "/"
	}

	ghClient.BaseURL = customURL
	ghClient.UploadURL = customURL

	return &GitHub{
		ctx:    context.Background(),
		client: ghClient,
	}
}

func ParseRepositoryURL(rawUrl string) (*Repository, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return ParseRepositoryURL("https://" + rawUrl)
	}

	if u.Hostname() != ghHostname {
		return nil, fmt.Errorf("Invalid hostname (%q), expected %q.", u.Hostname(), ghHostname)
	}

	path := strings.TrimPrefix(u.EscapedPath(), "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("Invalid GitHub URL format (%q)", rawUrl)
	}

	return &Repository{
		Owner: pathParts[0],
		Name:  pathParts[1],
	}, nil
}

func TreeURL(repo *Repository, ref string) string {
	return fmt.Sprintf("https://%s/%s/%s/tree/%s",
		ghHostname, repo.Owner, repo.Name, ref)
}
