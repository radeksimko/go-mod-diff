package diff

import (
	"testing"
)

func TestVersionIsEqual(t *testing.T) {
	testCases := []struct {
		first, second *Version
		expectedEqual bool
	}{
		{
			&Version{Version: "v0.0.0-20170808112155-b176d7def5d7", Revision: "b176d7def5d7", isRevision: true},
			&Version{Version: "b176d7def5d71bdd214203491f89843ed217f420", Revision: "b176d7def5d71bdd214203491f89843ed217f420", isRevision: true, Time: "2017-07-23T04:49:35Z"},
			true,
		},
	}

	for _, tc := range testCases {
		isEqual := tc.first.IsEqual(tc.second)
		if isEqual && !tc.expectedEqual {
			t.Fatalf("Expected %s to NOT equal %s", tc.first, tc.second)
		}
		if !isEqual && tc.expectedEqual {
			t.Fatalf("Expected %s to equal %s", tc.first, tc.second)
		}
	}
}
