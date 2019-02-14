package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/colorstring"
	"github.com/radeksimko/go-mod-diff/diff"
	"github.com/radeksimko/go-mod-diff/github"
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
	goModFile, err := gomod.ParseFile(filepath.Join(cwd, "go.mod"))

	// Parse govendor file
	govendorFile, err := govendor.ParseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Compare both and print out differences
	d, err := diff.CompareGoModWithGovendor(goModFile, govendorFile, gh)
	if err != nil {
		log.Fatal(err)
	}

	printDifference(d, gomod.GetVersionForModule(goModFile))

	total := len(goModFile.Require) - len(d.Matched)

	colorstring.Printf("\n\nMatched package revisions: [bold][green]%d[reset] of %d.\n"+
		"[bold]%d[reset] to check ([bold][red]%d[reset] not found and [bold][yellow]%d[reset] different revs).\n",
		len(d.Matched), len(goModFile.Require), total, len(d.NotFound), len(d.Different))
}

func printDifference(d *diff.Diff, vlF gomod.VersionLookupFunc) {
	for _, entry := range d.Errored {
		printDiffEntry(entry, vlF)
	}

	for _, entry := range d.NotFound {
		printDiffEntry(entry, vlF)
	}

	for _, entry := range d.Different {
		printDiffEntry(entry, vlF)
	}
}

func printDiffEntry(de *diff.DiffEntry, vlF gomod.VersionLookupFunc) {
	colorstring.Printf("\n[bold]%s[reset]\n", de.ModulePath)

	colorstring.Printf(" - go modules: %s\n", de.GoModVersion.String())

	if de.Error != nil {
		colorstring.Printf(" - [bold][red]Error:[reset] [red]%s[reset]\n", de.Error.Error())
	}

	repo, err := github.ParseRepositoryURL(de.ModulePath)
	if err == nil {
		ref, err := gomod.ParseRefFromVersion(de.GoModVersion.Version)
		if err == nil {
			fmt.Printf(" - GitHub: %s\n", github.TreeURL(repo, ref.String()))
		}
	}

	if de.GithubVersion != nil {
		colorstring.Printf(" - GitHub rev: %s\n", de.GithubVersion.String())
	}

	fmt.Print(" - govendor: ")
	if len(de.GoVendorVersions) > 0 {
		fmt.Printf("[\n")
		for _, gvv := range de.GoVendorVersions {
			if gvv.IsEqual(de.GoModVersion) || gvv.IsEqual(de.GithubVersion) {
				colorstring.Printf("       [green]%s\n", gvv.String())
			} else {
				fmt.Printf("       %s\n", gvv.String())
			}
		}
		fmt.Print("   ]\n")
	} else {
		colorstring.Print("[red]not found\n")
	}

	printGoModWhy(de.ModulePath, vlF)
}

func printGoModWhy(path string, vlF gomod.VersionLookupFunc) {
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
			githubURL := ""

			version := vlF(t)
			if version != "" {
				versionSuffix = " @ " + version

				repo, err := github.ParseRepositoryURL(t)
				if err == nil {
					ref, err := gomod.ParseRefFromVersion(version)
					if err == nil {
						githubURL = fmt.Sprintf(" (%s)", github.TreeURL(repo, ref.String()))
					}
				}
			}

			fmt.Printf("\n     %s%s%s", t, versionSuffix, githubURL)
		}
		fmt.Println("")
	}
	if len(mts) > 0 {
		fmt.Printf("   ]\n")
	}
}
