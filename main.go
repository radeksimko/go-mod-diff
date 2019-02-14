package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/colorstring"
	"github.com/radeksimko/go-mod-diff/github"
	"github.com/radeksimko/go-mod-diff/go-src/cmd/go/_internal/modfile"
	"github.com/radeksimko/go-mod-diff/gomod"
	"github.com/radeksimko/go-mod-diff/govendor"
)

func main() {
	// Setup GitHub connection
	gh := github.NewGitHub()
	if os.Getenv("GITHUB_TOKEN") != "" {
		gh = github.NewGitHubWithToken(os.Getenv("GITHUB_TOKEN"))
	}

	// Parse go modules file
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	f, err := gomod.ParseFile(filepath.Join(cwd, "go.mod"))

	// Parse govendor file
	vf, err := govendor.ParseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Compare both and print out differences
	matches := 0
	notFound := 0
	diffRevs := 0
	for _, r := range f.Require {
		mv := r.Mod
		govendorPkgs := govendor.FindPackages(vf.Package, mv.Path)

		ref, err := gomod.ParseRefFromVersion(mv.Version)
		if err != nil {
			log.Printf("Error: %s", err)
		}

		repo, err := github.ParseRepositoryURL(mv.Path)
		isGithubURL := (err == nil)

		if len(govendorPkgs) == 1 && ref.IsRevision() && strings.HasPrefix(govendorPkgs[0].Revision, ref.String()) {
			matches++
			continue
		} else if len(govendorPkgs) > 0 {
			gitHubSHA := ""
			githubRevSuffix := ""
			if !ref.IsRevision() && isGithubURL {
				sha, err := gh.GetCommitSHA(repo, ref.String())
				if err != nil {
					log.Printf("Error: %s", err)
				} else {
					gitHubSHA = sha
					githubRevSuffix = fmt.Sprintf(" (%s)", gitHubSHA)
					if len(govendorPkgs) == 1 && govendorPkgs[0].Revision == gitHubSHA {
						matches++
						continue
					}
				}
			}

			diffRevs++
			colorstring.Printf("\n[bold]%s[reset]\n - go modules: %s%s\n", mv.Path, mv.Version, githubRevSuffix)
			if isGithubURL {
				fmt.Printf(" - GitHub: %s\n", github.TreeURL(repo, ref.String()))
			}
			fmt.Print(" - govendor: [\n")
			for _, gvp := range govendorPkgs {
				revTime := ""
				if gvp.RevisionTime != "" {
					revTime = fmt.Sprintf(" (%s)", gvp.RevisionTime)
				}

				if ref.IsRevision() && strings.HasPrefix(gvp.Revision, ref.String()) || gitHubSHA != "" && gitHubSHA == gvp.Revision {
					colorstring.Printf("       [green]%s%s\n", gvp.Revision, revTime)
				} else {
					fmt.Printf("       %s%s\n", gvp.Revision, revTime)
				}
			}
			fmt.Print("   ]\n")
			printGoModWhy(mv.Path, f.Require)
		} else {
			notFound++
			colorstring.Printf("\n[bold]%s[reset]\n - go modules: %s\n", mv.Path, mv.Version)
			if isGithubURL {
				fmt.Printf(" - GitHub: %s\n", github.TreeURL(repo, ref.String()))
			}
			colorstring.Print(" - govendor: [red]Not found\n")
			printGoModWhy(mv.Path, f.Require)
		}
	}

	total := len(f.Require) - matches
	colorstring.Printf("\n\nMatched package revisions: [bold][green]%d[reset] of %d.\n"+
		"[bold]%d[reset] to check ([bold][red]%d[reset] not found and [bold][yellow]%d[reset] different revs).\n",
		matches, len(f.Require), total, notFound, diffRevs)
}

func printGoModWhy(path string, requires []*modfile.Require) {
	fmt.Printf(" - go mod why: ")
	mts, stderr, err := gomod.GoModWhy(path)
	if err != nil {
		colorstring.Printf("[bold][red]Failed to check (%s)[reset][red]\n%s", err, stderr)
		return
	}
	if len(mts) > 0 {
		fmt.Printf("[")
	} else {
		colorstring.Printf("[bold][red]Package not needed (try `go mod tidy`)\n")
	}
	for _, mt := range mts {
		for _, t := range mt {
			versionSuffix := ""
			version := gomod.GetVersionForModule(requires, t)
			if version != "" {
				versionSuffix = " @ " + version
			}

			fmt.Printf("\n     %s%s", t, versionSuffix)
		}
		fmt.Println("")
	}
	if len(mts) > 0 {
		fmt.Printf("   ]\n")
	}
}
