package diff

import (
	"os"

	"github.com/wjase/diffyaml/pkg/array"
	"gopkg.in/yaml.v3"
)

// GetYamlFileChanges loads the specs and compares them
func GetYamlFileChanges(oldSpec, newSpec string) (ChangeLogEntries, error) {
	spec1, err := ReadYAMLFile(oldSpec)
	if err != nil {
		return nil, err
	}
	spec2, err := ReadYAMLFile(newSpec)
	if err != nil {
		return nil, err
	}
	changes, err := GetYamlNodeChanges(spec1, spec2)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// GetYamlNodeChanges returns the changes between the two yaml documents
func GetYamlNodeChanges(doc1, doc2 *yaml.Node) (ChangeLogEntries, error) {
	changes := ChangeLogEntries{}
	hashed1 := HashNode(doc1)
	hashed2 := HashNode(doc2)

	changes = diffNode(hashed1, hashed2)
	return changes, nil
}

func diffNode(node1, node2 *HashedNode) ChangeLogEntries {
	changes := ChangeLogEntries{}
	switch node1.Node.Kind {
	case yaml.DocumentNode:
		childChanges := diffSequenceChildren(node1.Children, node2.Children)
		changes = append(changes, childChanges...)
	case yaml.SequenceNode:
		childChanges := diffSequenceChildren(node1.Children, node2.Children)
		changes = append(changes, childChanges...)
	case yaml.MappingNode:
		childChanges := diffMappedChildren(node1.Children, node2.Children)
		changes = append(changes, childChanges...)
	case yaml.ScalarNode:
		if string(node1.Hash) != string(node2.Hash) {
			changes = append(changes, ChangeLogEntry{
				Path:       node2.GetPath().String(),
				ChangeType: Changed,
				From:       node1.Node,
				To:         node2.Node,
			})
		}
	case yaml.AliasNode:
		// don't handle these yet
	}
	changes = prune(changes)
	return changes
}

func prune(changes ChangeLogEntries) ChangeLogEntries {
	pruned := make(ChangeLogEntries, 0, len(changes))
	for _, change := range changes {
		if change.ChangeType != NoChange {
			pruned = append(pruned, change)
		}
	}
	return pruned
}

// SequenceChangeLogEntry a changelog entry for a SequenceNode
type SequenceChangeLogEntry struct {
	Path               string
	ChangeType         ChangeType `yaml:"type"`
	From               *HashedNode
	To                 *HashedNode
	FromIndex, ToIndex int
}

func diffSequenceChildren(seq1, seq2 HashedNodes) ChangeLogEntries {

	if len(seq1) == 0 && len(seq2) == 0 {
		return ChangeLogEntries{}
	}
	if len(seq1) == 1 && len(seq2) == 1 {
		return diffNode(seq1[0], seq2[0])
	}

	if seq1[0].IsScalar() && seq2[0].IsScalar() {
		return diffScalarSequence(seq1, seq2)
	}

	// its either a sequence of sequences or a sequence of mapping nodes
	if seq1[0].Node.Kind == yaml.MappingNode && seq2[0].Node.Kind == yaml.MappingNode {
		return diffSequenceOfMappingNodes(seq1, seq2)
	}

	//diff by hash values
	changes := []SequenceChangeLogEntry{}

	hashSeq1 := hashList(seq1)
	hashSeq2 := hashList(seq2)

	hashDiffs := array.FromStringArray(hashSeq1).DiffsTo(hashSeq2)

	// report changes
	for _, item := range hashDiffs {
		var hashedItem *HashedNode
		entry := SequenceChangeLogEntry{}

		if item.Code == array.DeleteItem {
			hashedItem = seq1[item.FromIndex]
			entry.Path = hashedItem.GetPath().String()
			entry.ChangeType = Deleted
			entry.From = hashedItem
			entry.FromIndex = item.FromIndex
			entry.ToIndex = item.ToIndex
			changes = append(changes, entry)
		}
		if item.Code == array.AddItem {
			hashedItem = seq2[item.ToIndex]
			entry.Path = hashedItem.GetPath().String()
			entry.ChangeType = Added
			entry.To = hashedItem
			entry.FromIndex = item.FromIndex
			entry.ToIndex = item.ToIndex
			changes = append(changes, entry)
		}
	}
	mergedChanges := mergeMovedChanges(mergeUpdateChanges(changes))

	return toChangeLog(mergedChanges)
}

func diffSequenceOfMappingNodes(children1, children2 HashedNodes) ChangeLogEntries {
	seq1Map := map[string]interface{}{}
	seq2Map := map[string]interface{}{}
	for _, child := range children1 {
		seq1Map[child.Key] = child
	}
	for _, child := range children2 {
		seq2Map[child.Key] = child
	}
	changes := ChangeLogEntries{}
	added, deleted, common := array.FromStringMap(seq1Map).DiffsTo(seq2Map)
	for eachKey := range added {
		addedIndex, hashedNode := children2.Find(eachKey)
		changes = append(changes, ChangeLogEntry{
			ChangeType: Added,
			Path:       hashedNode.GetPath().String(),
			ToIndex:    &addedIndex,
			To:         hashedNode.Node,
		})
	}
	for eachKey := range deleted {
		deletedIndex, hashedNode := children1.Find(eachKey)
		changes = append(changes, ChangeLogEntry{
			ChangeType: Deleted,
			Path:       hashedNode.GetPath().String(),
			FromIndex:  &deletedIndex,
			From:       hashedNode.Node,
		})
	}

	for key := range common {
		item1 := seq1Map[key].(*HashedNode)
		item2 := seq2Map[key].(*HashedNode)
		if item1.Hash != item2.Hash {
			changes = append(changes, diffNode(item1, item2)...)
		}
	}

	return changes
}

func diffScalarSequence(children1, children2 HashedNodes) ChangeLogEntries {
	seq1Values := make([]string, len(children1))
	seq2Values := make([]string, len(children2))

	for index, item := range children1 {
		seq1Values[index] = item.Hash
	}

	for index, item := range children2 {
		seq2Values[index] = item.Hash
	}
	diffs := array.FromStringArray(seq1Values).DiffsTo(seq2Values)

	changes := make(ChangeLogEntries, len(diffs))
	for index, diff := range diffs {

		change := diffToChange(diff)

		if change.ToIndex != nil {
			change.Path = children2[*change.ToIndex].GetPath().String()
			change.To = children2[*change.ToIndex].Node
		}
		if change.FromIndex != nil {
			change.Path = children1[*change.FromIndex].GetPath().String()
			change.From = children1[*change.FromIndex].Node
		}
		changes[index] = change
	}

	// add + delete same value different index => move
	for addedIndex, added := range changes {
		if added.ChangeType == Added {
			for deletedIndex, deleted := range changes {
				if deleted.ChangeType == Deleted {
					if added.To.Value == deleted.From.Value {
						item := changes[addedIndex]
						item.Path = deleted.Path
						item.ChangeType = Moved
						item.FromIndex = deleted.FromIndex
						item.ToIndex = added.ToIndex
						item.To = nil
						changes[deletedIndex].ChangeType = NoChange
						changes[addedIndex] = item
					}
				}
			}
		}
	}

	return changes
}

func diffToChange(diff array.Diff) ChangeLogEntry {
	change := ChangeLogEntry{}
	switch diff.Code {
	case array.AddItem:
		to := diff.ToIndex
		change.ChangeType = Added
		change.ToIndex = &to
		return change
	case array.DeleteItem:
		change.ChangeType = Deleted
		from := diff.FromIndex
		change.FromIndex = &from
		return change
	default:
		change.ChangeType = NoChange
		return change
	}
}

func toChangeLog(seqChanges []SequenceChangeLogEntry) ChangeLogEntries {
	changes := make(ChangeLogEntries, 0, len(seqChanges))
	for _, c := range seqChanges {
		if c.ChangeType == Changed {
			if !c.From.IsScalar() {
				changes = append(changes, diffNode(c.From, c.To)...)
				continue
			}
		}
		if c.ChangeType == Moved {
			if c.From.Hash != c.To.Hash {
				changes = append(changes, diffNode(c.From, c.To)...)
				continue
			}
		}
		entry := ChangeLogEntry{Path: c.Path, ChangeType: c.ChangeType}
		if c.From != nil {
			entry.From = c.From.Node
		}
		if c.To != nil {
			entry.To = c.To.Node
		}
		if c.ChangeType == Moved {
			from := c.FromIndex
			entry.FromIndex = &from
			to := c.ToIndex
			entry.ToIndex = &to
		}

		changes = append(changes, entry)
	}
	return changes
}

func mergeMovedChanges(changes []SequenceChangeLogEntry) []SequenceChangeLogEntry {
	deletions := map[int]bool{}
	updateChanges := []SequenceChangeLogEntry{}
	for delIndex, eachDelete := range changes {
		if eachDelete.ChangeType == Deleted {
			for addIndex, eachAdd := range changes {
				if eachAdd.ChangeType == Added {
					if eachAdd.To.Hash == eachDelete.From.Hash {
						updateChanges = append(updateChanges,
							SequenceChangeLogEntry{
								ChangeType: Moved,
								Path:       eachDelete.Path,
								FromIndex:  eachDelete.FromIndex,
								ToIndex:    eachAdd.ToIndex,
								From:       eachDelete.From,
								To:         eachAdd.To,
							})
						deletions[addIndex] = true
						deletions[delIndex] = true
					}
				}
			}
		}
	}
	mergedChanges := updateChanges
	for ind, eachChange := range changes {
		if _, exists := deletions[ind]; !exists {
			mergedChanges = append(mergedChanges, eachChange)
		}
	}
	return mergedChanges
}

func isSameSequenceIdentity(node1, node2 *HashedNode) bool {
	if node1.Node.Kind != node2.Node.Kind {
		return false
	}
	// assume that scalars at the same pos in old and new are the same item
	if node1.IsScalar() {
		return true
	}

	if node1.Node.Kind == yaml.MappingNode {
		// empty map
		if len(node1.Children) == 0 && len(node2.Children) == 0 {
			return true
		}
		// assume the first key of a sequence of mapping nodes denotes identity
		if len(node1.Children) > 0 && len(node2.Children) > 0 {
			return node1.Children[0].Hash == node2.Children[0].Hash
		}
	}
	return node1.Hash == node2.Hash
}

func mergeUpdateChanges(changes []SequenceChangeLogEntry) []SequenceChangeLogEntry {
	deletions := map[int]bool{}
	updateChanges := []SequenceChangeLogEntry{}
	var currentAddIndex = -1
	var currentDelIndex = -1
	for index, change := range changes {
		if change.ChangeType == Added {
			currentAddIndex = index
		}
		if change.ChangeType == Deleted {
			currentDelIndex = index
		}
		if currentAddIndex >= 0 && currentDelIndex >= 0 {
			eachDelete := changes[currentDelIndex]
			eachAdd := changes[currentAddIndex]
			if isSameSequenceIdentity(eachAdd.To, eachDelete.From) &&
				eachAdd.To.Hash != eachDelete.From.Hash {
				updateChanges = append(updateChanges, SequenceChangeLogEntry{ChangeType: Changed, Path: eachAdd.Path, From: eachDelete.From, To: eachAdd.To})
				deletions[currentAddIndex] = true
				deletions[currentDelIndex] = true
				currentAddIndex = -1
				currentDelIndex = -1
			}
		}
	}
	mergedChanges := updateChanges
	for ind, eachChange := range changes {
		if _, exists := deletions[ind]; !exists {
			mergedChanges = append(mergedChanges, eachChange)
		}
	}
	return mergedChanges
}

func hashList(seq HashedNodes) []string {
	hashSeq := make([]string, len(seq))
	for ind, item := range seq {
		hashSeq[ind] = string(item.Hash)
	}
	return hashSeq
}

func toHashedMap(nodes HashedNodes) map[string]HashedNode {
	mappedNodes := map[string]HashedNode{}
	for _, node := range nodes {
		mappedNodes[node.Key] = *node
	}
	return mappedNodes
}

func diffMappedChildren(children1, children2 HashedNodes) ChangeLogEntries {
	changes := ChangeLogEntries{}
	map1 := toHashedMap(children1)
	map2 := toHashedMap(children2)

	for k, item1 := range map1 {
		if item2, ok := map2[k]; !ok {
			changes = append(changes, ChangeLogEntry{Path: item1.GetPath().String(), ChangeType: Deleted, From: item1.Node})
		} else {
			if item2.Hash != item1.Hash {
				changes = append(changes, diffNode(&item1, &item2)...)
			}
		}
	}
	for k, item2 := range map2 {
		if _, ok := map1[k]; !ok {
			changes = append(changes, ChangeLogEntry{Path: item2.GetPath().String(), ChangeType: Added, To: item2.Node})
		}
	}
	return changes
}

// ReadYAMLFile Reads a YAML file into a yaml.Node
func ReadYAMLFile(filename string) (*yaml.Node, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(f)

	var cf yaml.Node

	err = decoder.Decode(&cf)
	if err != nil {
		return nil, err
	}
	return &cf, nil
}
