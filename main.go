package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/govendor/vendorfile"
	"github.com/mitchellh/colorstring"
	"github.com/radeksimko/go-mod-diff/go-src/cmd/go/_internal/modfile"
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

	// Compare both and print out differences
	matches := 0
	notFound := 0
	diffRevs := 0
	for _, r := range f.Require {
		mv := r.Mod
		govendorPkgs := findGoVendorPkgs(oldPackages, mv.Path)

		goModRev := parseRevision(mv.Version)

		if len(govendorPkgs) == 1 && goModRev != "" && strings.HasPrefix(govendorPkgs[0].Revision, goModRev) {
			matches++
			continue
		} else if len(govendorPkgs) > 0 {
			diffRevs++
			colorstring.Printf("[bold]%s[reset]\n - go modules: %s\n", mv.Path, mv.Version)
			fmt.Print(" - govendor: [\n")
			for _, gvp := range govendorPkgs {
				if goModRev != "" && strings.HasPrefix(gvp.Revision, goModRev) {
					colorstring.Printf("       [green]%s (%s)\n", gvp.Revision, gvp.RevisionTime)
				} else {
					fmt.Printf("       %s (%s)\n", gvp.Revision, gvp.RevisionTime)
				}
			}
			fmt.Print("   ]\n")
		} else {
			notFound++
			colorstring.Printf("[bold]%s[reset]\n - go modules: %s\n", mv.Path, mv.Version)
			colorstring.Print(" - govendor: [red]Not found\n")
		}
	}

	total := len(f.Require) - matches
	colorstring.Printf("\n\nMatched package revisions: [bold][green]%d[reset] of %d.\n"+
		"[bold]%d[reset] to check ([bold][red]%d[reset] not found and [bold][yellow]%d[reset] different revs).\n",
		matches, len(f.Require), total, notFound, diffRevs)
}

func parseRevision(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) == 3 {
		return parts[2]
	}
	return ""
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
