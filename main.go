package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kardianos/govendor/vendorfile"
	"github.com/mitchellh/colorstring"
	"github.com/radeksimko/go-mod-diff/github"
	"github.com/radeksimko/go-mod-diff/go-src/cmd/go/_internal/modfile"
	"github.com/radeksimko/go-mod-diff/gomod"
	"golang.org/x/tools/go/vcs"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Parse go modules file
	gomodPath := filepath.Join(cwd, "go.mod")
	data, err := ioutil.ReadFile(gomodPath)
	if err != nil {
		log.Fatal(err)
	}
	f, err := modfile.Parse(gomodPath, data, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Parse govendor file
	vendorJsonPath := os.Args[1]
	src, err := os.Open(vendorJsonPath)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()
	vf := &vendorfile.File{}
	err = vf.Unmarshal(io.Reader(src))
	if err != nil {
		log.Fatal(err)
	}

	oldPackages := vf.Package

	gh := github.NewGitHub()
	if os.Getenv("GITHUB_TOKEN") != "" {
		gh = github.NewGitHubWithToken(os.Getenv("GITHUB_TOKEN"))
	}

	// Compare both and print out differences
	matches := 0
	notFound := 0
	diffRevs := 0
	for _, r := range f.Require {
		mv := r.Mod
		govendorPkgs := findGoVendorPkgs(oldPackages, mv.Path)

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
	mts, stderr, err := goModWhy(path)
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
			version := getVersionForModule(requires, t)
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

func goModWhy(importPath string) ([][]string, string, error) {
	cmd := exec.Command("go", "mod", "why", "-m", importPath)
	var stdout, stderr bytes.Buffer
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, stderr.String(), err
	}

	capture := false
	moduleTrees := make([][]string, 0)
	tree := make([]string, 0)
	i, j := 0, 0
	lastCapturedRoot := ""
	for {
		line, err := stdout.ReadString('\n')

		// beginning of tree
		if strings.HasPrefix(line, "# ") {
			capture = true
			j = 0
			continue
		}

		// end of tree
		if err != nil || line == "\n" || strings.Contains(line, "module does not need") {
			capture = false
			lastCapturedRoot = ""
			// Save tree, if there's anything to save
			if len(tree) > 0 {
				moduleTrees = append(moduleTrees, tree)
				tree = make([]string, 0)
				i++
			}
			if err != nil {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)

		if capture {
			if j == 0 {
				tree = append(tree, line)
			} else {
				repoRoot, _ := repoRootForImportPath(line)
				if repoRoot != lastCapturedRoot {
					// log.Printf("[%d] %q", j, repoRoot)
					tree = append(tree, repoRoot)
					lastCapturedRoot = repoRoot
				}
			}
			j++
		}
	}

	return moduleTrees, "", nil
}

func getVersionForModule(modules []*modfile.Require, modPath string) string {
	for _, m := range modules {
		if m.Mod.Path == modPath {
			return m.Mod.Version
		}
	}
	return ""
}

func repoRootForImportPath(importPath string) (string, error) {
	rr, err := vcs.RepoRootForImportPath(importPath, false)
	if err != nil {
		return "", err
	}
	return rr.Root, nil
}

func findGoVendorPkgs(packages []*vendorfile.Package, importPath string) []*vendorfile.Package {
	var pkgs []*vendorfile.Package
	for _, p := range packages {
		if strings.HasPrefix(p.Path, importPath) {
			p.Path = importPath
			if !govendorPkgExists(pkgs, p) {
				pkgs = append(pkgs, p)
			}
		}
	}
	return pkgs
}

func govendorPkgExists(packages []*vendorfile.Package, pkg *vendorfile.Package) bool {
	for _, p := range packages {
		if p.Revision == pkg.Revision {
			return true
		}
	}
	return false
}
