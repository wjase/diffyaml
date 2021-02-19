package diff

import "gopkg.in/yaml.v3"

// ChangeType describes the change
type ChangeType int

const (
	// NoChange zero value = no change
	NoChange ChangeType = iota
	//Added  key or item was added
	Added
	//Deleted key or item was deleted
	Deleted
	//Moved item moved in a sequence
	Moved
	//Changed item was modified
	Changed
)

// ChangeTypeLabels used for printing changes
var ChangeTypeLabels = []string{"no-change", "added", "deleted", "moved", "changed"}

// String implement the Stringer interface
func (d ChangeType) String() string {
	return ChangeTypeLabels[d]
}

// MarshalYAML custom marshal function
func (d ChangeType) MarshalYAML() (interface{}, error) {
	return ChangeTypeLabels[d], nil
}

// ChangeLogEntry info on a changed node
type ChangeLogEntry struct {
	Path       string
	ChangeType ChangeType `yaml:"type,omitempty"`
	From       *yaml.Node `yaml:"from,omitempty"`
	To         *yaml.Node `yaml:"to,omitempty"`
	FromIndex  *int       `yaml:"from-index,omitempty"`
	ToIndex    *int       `yaml:"to-index,omitempty"`
	Line       *int       `yaml:"line,omitempty"`
	Column     *int       `yaml:"column,omitempty"`
}

// ChangeLogEntries custom collection type
type ChangeLogEntries []ChangeLogEntry

func (l ChangeLogEntries) Len() int {
	return len(l)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (l ChangeLogEntries) Less(i, j int) bool {
	return l[i].Path < l[j].Path
}

// Swap swaps the elements with indexes i and j.
func (l ChangeLogEntries) Swap(i, j int) {
	var item = l[i]
	l[i] = l[j]
	l[j] = item

}
