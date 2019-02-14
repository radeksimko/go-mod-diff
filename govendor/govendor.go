package govendor

import (
	"io"
	"os"
	"strings"

	"github.com/kardianos/govendor/vendorfile"
)

func ParseFile(path string) (*vendorfile.File, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	vf := &vendorfile.File{}
	err = vf.Unmarshal(io.Reader(src))
	if err != nil {
		return nil, err
	}
	return vf, nil
}

func FindPackages(packages []*vendorfile.Package, importPath string) []*vendorfile.Package {
	var pkgs []*vendorfile.Package
	for _, p := range packages {
		if strings.HasPrefix(p.Path, importPath) {
			p.Path = importPath
			if !packageExists(pkgs, p) {
				pkgs = append(pkgs, p)
			}
		}
	}
	return pkgs
}

func packageExists(packages []*vendorfile.Package, pkg *vendorfile.Package) bool {
	for _, p := range packages {
		if p.Revision == pkg.Revision {
			return true
		}
	}
	return false
}
