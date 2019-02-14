package diff

import (
	"fmt"
	"strings"

	"github.com/kardianos/govendor/vendorfile"
	"github.com/radeksimko/go-mod-diff/github"
	"github.com/radeksimko/go-mod-diff/go-src/cmd/go/_internal/modfile"
	"github.com/radeksimko/go-mod-diff/gomod"
	"github.com/radeksimko/go-mod-diff/govendor"
)

type Diff struct {
	Matched   []*DiffEntry
	NotFound  []*DiffEntry
	Different []*DiffEntry
	Errored   []*DiffEntry
}

type DiffEntry struct {
	ModulePath       string
	GoModVersion     *Version
	GithubVersion    *Version
	GoVendorVersions []*Version
	Error            error
}

type Version struct {
	Version    string
	Revision   string
	Time       string
	isRevision bool
}

func (v *Version) String() string {
	output := v.Version
	if v.Revision != "" && v.Version != v.Revision {
		output += fmt.Sprintf(" / %s", v.Revision)
	}
	if v.Time != "" {
		output += fmt.Sprintf(" (%s)", v.Time)
	}

	return output
}

// IsRevision returns true if v.Version is a revision
func (v *Version) IsRevision() bool {
	return v.isRevision
}

func (v *Version) IsEqual(cv *Version) bool {
	if cv == nil {
		return false
	}

	if v.Revision != "" && cv.Revision != "" {
		return strings.HasPrefix(v.Revision, cv.Revision) || strings.HasPrefix(cv.Revision, v.Revision)
	}

	return v.Version == cv.Version
}

func CompareGoModWithGovendor(goModFile *modfile.File, gvFile *vendorfile.File, gh *github.GitHub) (*Diff, error) {
	d := &Diff{
		Matched:   make([]*DiffEntry, 0),
		NotFound:  make([]*DiffEntry, 0),
		Different: make([]*DiffEntry, 0),
		Errored:   make([]*DiffEntry, 0),
	}

	for _, r := range goModFile.Require {
		mv := r.Mod

		diffEntry := &DiffEntry{
			ModulePath: mv.Path,
			GoModVersion: &Version{
				Version: mv.Version,
			},
		}

		ref, err := gomod.ParseRefFromVersion(mv.Version)
		if err != nil {
			diffEntry.Error = err
			d.Errored = append(d.Errored, diffEntry)
			continue
		}
		diffEntry.GoModVersion.isRevision = ref.IsRevision()
		if ref.IsRevision() {
			diffEntry.GoModVersion.Revision = ref.String()
		}

		repo, err := github.ParseRepositoryURL(mv.Path)
		isGithubURL := (err == nil)

		govendorPkgs := govendor.FindPackages(gvFile.Package, mv.Path)

		if len(govendorPkgs) == 1 && ref.IsRevision() && strings.HasPrefix(govendorPkgs[0].Revision, ref.String()) {
			diffEntry.GoVendorVersions = govendorVersions(govendorPkgs)
			d.Matched = append(d.Matched, diffEntry)
			continue
		} else if len(govendorPkgs) > 0 {
			if !ref.IsRevision() && isGithubURL {
				// Try converting reference to a revision via GitHub and compare
				githubSHA, err := gh.GetCommitSHA(repo, ref.String())
				if err != nil {
					diffEntry.Error = fmt.Errorf("Failed to get ref SHA from GitHub: %s", err)
					d.Errored = append(d.Errored, diffEntry)
					continue
				}

				diffEntry.GithubVersion = &Version{
					Version:    githubSHA,
					Revision:   githubSHA,
					Time:       "", // TODO: Add timestamp
					isRevision: true,
				}

				if len(govendorPkgs) == 1 && govendorPkgs[0].Revision == githubSHA {
					diffEntry.GoVendorVersions = govendorVersions(govendorPkgs)
					d.Matched = append(d.Matched, diffEntry)
					continue
				}
			}

			diffEntry.GoVendorVersions = govendorVersions(govendorPkgs)
			d.Different = append(d.Different, diffEntry)
			continue
		}

		d.NotFound = append(d.NotFound, diffEntry)
	}

	return d, nil
}

func govendorVersions(pkgs []*vendorfile.Package) []*Version {
	versions := make([]*Version, 0)
	for _, pkg := range pkgs {
		versions = append(versions, &Version{
			Version:    pkg.Revision,
			Revision:   pkg.Revision,
			Time:       pkg.RevisionTime,
			isRevision: true,
		})
	}
	return versions
}
