package array

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	listA := []string{"abc", "def", "ghi", "jkl"}
	emptyList := []string{}
	deletedDEF := []string{"abc", "ghi", "jkl"}
	addedXYZ := []string{"abc", "def", "xyz", "ghi", "jkl"}
	movedABC := []string{"def", "ghi", "jkl", "abc"}
	deletedDEFmovedABC := []string{"ghi", "abc", "jkl"}
	changedDEFtoXYZ := []string{"abc", "xyz", "ghi", "jkl"}
	changedJKLtoXYZ := []string{"abc", "def", "ghi", "xyz"}
	changedJKLtoXYZAndAddedTYU := []string{"abc", "def", "ghi", "xyz", "tyu"}

	testCases := []struct {
		desc     string
		from     []string
		to       []string
		expected []Diff
	}{
		{
			desc:     "equal lists",
			from:     listA,
			to:       listA,
			expected: []Diff{},
		},
		{
			desc:     "all added",
			from:     emptyList,
			to:       listA,
			expected: allAdded(listA),
		},
		{
			desc:     "all deleted",
			from:     listA,
			to:       emptyList,
			expected: allDeleted(listA),
		},
		{
			desc:     "one deleted",
			from:     listA,
			to:       deletedDEF,
			expected: []Diff{{Code: DeleteItem, FromIndex: 1, FromValue: "def"}},
		},
		{
			desc:     "one added",
			from:     listA,
			to:       addedXYZ,
			expected: []Diff{{Code: AddItem, FromIndex: 2, ToIndex: 2, ToValue: "xyz"}},
		},
		{
			desc: "movedABC",
			from: listA,
			to:   movedABC,
			expected: []Diff{
				{Code: DeleteItem, FromIndex: 0, ToIndex: 0, FromValue: "abc"},
				{Code: AddItem, FromIndex: 4, ToIndex: 3, ToValue: "abc"}},
		},
		{
			desc: "deletedDEFmovedABC",
			from: listA,
			to:   deletedDEFmovedABC,
			expected: []Diff{
				{Code: AddItem, FromIndex: 0, ToIndex: 0, FromValue: "", ToValue: "ghi"},
				{Code: DeleteItem, FromIndex: 1, ToIndex: 0, FromValue: "def", ToValue: ""},
				{Code: DeleteItem, FromIndex: 2, ToIndex: 0, FromValue: "ghi", ToValue: ""}},
		},
		{
			desc: "changedDEFtoXYZ",
			from: listA,
			to:   changedDEFtoXYZ,
			expected: []Diff{
				{Code: DeleteItem, FromIndex: 1, ToIndex: 0, FromValue: "def", ToValue: ""},
				{Code: AddItem, FromIndex: 2, ToIndex: 1, FromValue: "", ToValue: "xyz"},
			},
		},
		{
			desc: "changedJKLtoXYZ",
			from: listA,
			to:   changedJKLtoXYZ,
			expected: []Diff{
				{Code: DeleteItem, FromIndex: 3, ToIndex: 0, FromValue: "jkl", ToValue: ""},
				{Code: AddItem, FromIndex: 4, ToIndex: 3, FromValue: "", ToValue: "xyz"},
			},
		},
		{
			desc: "changedJKLtoXYZAndAddedTYU",
			from: listA,
			to:   changedJKLtoXYZAndAddedTYU,
			expected: []Diff{
				{Code: AddItem, FromIndex: 3, ToIndex: 3, FromValue: "", ToValue: "xyz"},
				{Code: AddItem, FromIndex: 3, ToIndex: 4, FromValue: "", ToValue: "tyu"},
				{Code: DeleteItem, FromIndex: 3, ToIndex: 0, FromValue: "jkl", ToValue: ""},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			diffs := FromStringArray(tC.from).DiffsTo(tC.to)
			require.Equal(t, tC.expected, diffs)
		})
	}
}

func TestMapDiff(t *testing.T) {
	mapA := map[string]interface{}{"abc": 1, "def": 2, "ghi": 3, "jkl": 4}
	added, deleted, common := FromStringMap(mapA).DiffsTo(mapA)
	require.Equal(t, map[string]interface{}{}, added)
	require.Equal(t, map[string]interface{}{}, deleted)

	commonDiffs := map[string]interface{}{"abc": Pair{1, 1}, "def": Pair{2, 2}, "ghi": Pair{3, 3}, "jkl": Pair{4, 4}}
	require.Equal(t, commonDiffs, common)

	mapB := map[string]interface{}{"abc": 2, "ghi": 3, "jkl": 4, "xyz": 5, "fgh": 6}
	added, deleted, common = FromStringMap(mapA).DiffsTo(mapB)
	require.Equal(t, map[string]interface{}{"xyz": 5, "fgh": 6}, added)
	require.Equal(t, map[string]interface{}{"def": 2}, deleted)
	commonDiffs = map[string]interface{}{"abc": Pair{1, 2}, "ghi": Pair{3, 3}, "jkl": Pair{4, 4}}
	require.Equal(t, commonDiffs, common)
}
