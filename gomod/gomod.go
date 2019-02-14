package gomod

import (
	"fmt"
	"strings"
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
