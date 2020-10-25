package array

// This is a simple DSL for diffing arrays

// DiffType enum identifying the type of change
type DiffType int

const (
	// DeleteItem - item was deleted
	DeleteItem DiffType = iota // remove item
	// AddItem Item was added
	AddItem
	// Unknown - Neither add nor delete
	Unknown
)

// DiffTypeLabels used for printing changes
var DiffTypeLabels = []string{"deleted", "added", "unknown"}

// String implement the Stringer interface
func (d DiffType) String() string {
	return DiffTypeLabels[d]
}

// Diff details a change from one collection to another
type Diff struct {
	Code      DiffType
	FromIndex int
	// ToIndex the position in the updated index
	ToIndex   int
	FromValue string
	ToValue   string
}

// FromArrayStruct utility struct to encompass diffing of string arrays
type FromArrayStruct struct {
	from []string
}

// Max returns the larger of x or y.
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// FromStringArray starts a fluent diff expression
func FromStringArray(from []string) FromArrayStruct {
	return FromArrayStruct{from}
}

// DiffsTo returns a set of diffs from the original array to the
// specified array
func (f FromArrayStruct) DiffsTo(toArray []string) []Diff {

	diffs := []Diff{}
	fromLen := len(f.from)
	toLen := len(toArray)
	if fromLen == 0 && toLen == 0 {
		return diffs
	}

	if fromLen == 0 {
		return allAdded(toArray)
	}

	if toLen == 0 {
		return allDeleted(f.from)
	}

	return diff(f.from, toArray)

}

func allDeleted(ary []string) []Diff {
	diffs := make([]Diff, len(ary))
	for index, item := range ary {
		diffs[index] = Diff{Code: DeleteItem, FromIndex: index, FromValue: item}
	}
	return diffs
}

func allAdded(ary []string) []Diff {
	diffs := make([]Diff, len(ary))
	for index, item := range ary {
		diffs[index] = Diff{Code: AddItem, ToIndex: index, ToValue: item}
	}
	return diffs
}

// FromMapStruct utility struct to encompass diffing of string arrays
type FromMapStruct struct {
	srcMap map[string]interface{}
}

// FromStringMap starts a comparison by declaring a source map
func FromStringMap(srcMap map[string]interface{}) FromMapStruct {
	return FromMapStruct{srcMap}
}

// Pair stores a pair of items which share a key in two maps
type Pair struct {
	First  interface{}
	Second interface{}
}

// DiffsTo - generates diffs for a comparison
func (f FromMapStruct) DiffsTo(destMap map[string]interface{}) (added, deleted, common map[string]interface{}) {
	added = make(map[string]interface{})
	deleted = make(map[string]interface{})
	common = make(map[string]interface{})

	inSrc := 1
	inDest := 2

	m := make(map[string]int)

	// enter values for all items in the source array
	for key := range f.srcMap {
		m[key] = inSrc
	}

	// now either set or 'boolean or' a new flag if in the second collection
	for key := range destMap {
		if _, ok := m[key]; ok {
			m[key] |= inDest
		} else {
			m[key] = inDest
		}
	}
	// finally inspect the values and generate the left,right and shared collections
	// for the shared items, store both values in case there's a diff
	for key, val := range m {
		switch val {
		case inSrc:
			deleted[key] = f.srcMap[key]
		case inDest:
			added[key] = destMap[key]
		default:
			common[key] = Pair{f.srcMap[key], destMap[key]}
		}
	}
	return added, deleted, common
}

//  Returns a minimal list of differences between 2 lists e and f
//  requring O(min(len(e),len(f))) space and O(min(len(e),len(f)) * D)
//  worst-case execution time where D is the number of differences.
func diff(e, f []string) []Diff {
	diffs := recDiff(e, f, 0, 0)
	for ind, eachDiff := range diffs {
		if eachDiff.Code == AddItem {
			eachDiff.ToValue = f[eachDiff.ToIndex]
		}
		if eachDiff.Code == DeleteItem {
			eachDiff.FromValue = e[eachDiff.FromIndex]
		}
		diffs[ind] = eachDiff
	}
	return diffs
}

// Min returns the lesser of x or y.
func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// Used in the port from Python to imitate the mod behaviour
// of python
func pyMod(d, m int) int {
	var res int = d % m
	if (res < 0 && m > 0) || (res > 0 && m < 0) {
		return res + m
	}
	return res
}

func recDiff(list1, list2 []string, i, j int) []Diff {
	diffs := []Diff{}
	//  Documented at http://blog.robertelder.org/diff-algorithm/
	N, M, L, Z := len(list1), len(list2), len(list1)+len(list2), 2*Min(len(list1), len(list2))+2
	if N > 0 && M > 0 {
		w, g, p := N-M, make([]int, Z), make([]int, Z)
		// for h := 0; h < (L/2+(L%2))+1; h++ {
		for h := 0; h < (L/2+(pyMod(L, 2)))+1; h++ {
			for r := 0; r < 2; r++ {
				c, d, o, m := g, p, 1, 1
				if r != 0 {
					c, d, o, m = p, g, 0, -1
				}
				// for k in range(-(h-2*max(0,h-M)), h-2*max(0,h-N)+1, 2):
				for k := -(h - 2*Max(0, h-M)); k < h-2*Max(0, h-N)+1; k = k + 2 {
					a := c[pyMod((k-1), Z)] + 1
					if k == -h || k != h && c[pyMod((k-1), Z)] < c[pyMod((k+1), Z)] {
						a = c[pyMod((k+1), Z)]
					}
					b := a - k
					s, t := a, b
					// while a<N and b<M and e[(1-o)*N+m*a+(o-1)]==f[(1-o)*M+m*b+(o-1)]:
					for a < N && b < M && list1[(1-o)*N+m*a+(o-1)] == list2[(1-o)*M+m*b+(o-1)] {
						a, b = a+1, b+1
					}
					c[pyMod(k, Z)] = a
					z := -(k - w)
					if L%2 == o && z >= -(h-o) && z <= h-o && c[pyMod(k, Z)]+d[pyMod(z, Z)] >= N {
						D, x, y, u, v := 2*h, N-a, M-b, N-s, M-t
						if o == 1 {
							D, x, y, u, v = 2*h-1, s, t, a, b
						}
						if D > 1 || (x != u && y != v) {
							diffs = append(diffs, recDiff(list1[0:x], list2[0:y], i, j)...)
							diffs = append(diffs, recDiff(list1[u:N], list2[v:M], i+u, j+v)...)
							return diffs
						} else if M > N {
							return recDiff([]string{}, list2[N:M], i+N, j+N)
						} else if M < N {
							return recDiff(list1[M:N], []string{}, i+M, j+M)
						}
						return diffs
					}
				}
			}
		}
	} else if N > 0 {
		for n := 0; n < N; n++ {
			diffs = append(diffs, Diff{Code: DeleteItem, FromIndex: i + n})
		}
	} else {
		//#  Modify the return statements below if you want a different edit script format
		for n := 0; n < M; n++ {
			diffs = append(diffs, Diff{Code: AddItem, FromIndex: i, ToIndex: j + n})
		}
	}
	return diffs
}
