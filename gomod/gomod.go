package gomod

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/tools/go/vcs"
)

type VersionRef struct {
	ref   string
	isRev bool
}

func (vr *VersionRef) String() string {
	return vr.ref
}

func (vr *VersionRef) IsRevision() bool {
	return vr.isRev
}

func ParseRefFromVersion(rawVersion string) (*VersionRef, error) {
	if rawVersion == "" {
		return nil, fmt.Errorf("Expected version, empty string given")
	}

	if strings.HasPrefix(rawVersion, "v0.0.0-") {
		parts := strings.Split(rawVersion, "-")
		if len(parts) != 3 {
			return nil, fmt.Errorf("Unexpected version format (%q)", rawVersion)
		}
		return &VersionRef{parts[2], true}, nil
	}

	rawVersion = strings.TrimSuffix(rawVersion, "+incompatible")

	return &VersionRef{rawVersion, false}, nil
}

func GoModWhy(importPath string) ([][]string, string, error) {
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

func repoRootForImportPath(importPath string) (string, error) {
	rr, err := vcs.RepoRootForImportPath(importPath, false)
	if err != nil {
		return "", err
	}
	return rr.Root, nil
}
