package gomod

import (
	"testing"
)

func TestParseRefFromVersion(t *testing.T) {
	testCases := []struct {
		rawVersion  string
		expectedErr bool
		expectedRef *VersionRef
	}{
		{
			rawVersion:  "",
			expectedErr: true,
		},
		{
			rawVersion:  "v1.2.1",
			expectedRef: &VersionRef{ref: "v1.2.1", isRev: false},
		},
		{
			rawVersion:  "v0.0.0-20170418151526-7e4b007599d4",
			expectedRef: &VersionRef{ref: "7e4b007599d4", isRev: true},
		},
		{
			rawVersion:  "v0.11.12-beta1.0.20190227065421-fc531f54a878",
			expectedRef: &VersionRef{ref: "fc531f54a878", isRev: true},
		},
		{
			rawVersion:  "v4.2.1+incompatible",
			expectedRef: &VersionRef{ref: "v4.2.1", isRev: false},
		},
	}

	for _, tc := range testCases {
		ref, err := ParseRefFromVersion(tc.rawVersion)
		if tc.expectedErr {
			if err == nil {
				t.Fatalf("Expected %q to return error, none given.", tc.rawVersion)
			}
			continue
		}

		if err != nil {
			t.Fatalf("Parsing %q failed: %s", tc.rawVersion, err)
		}

		if *ref != *tc.expectedRef {
			t.Fatalf("Expected %q, given: %q", tc.expectedRef, ref)
		}
	}
}
