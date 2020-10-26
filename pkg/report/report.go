package report

import (
	"io"
	"sort"

	"github.com/wjase/diffyaml/pkg/diff"
	"gopkg.in/yaml.v3"
)

// WriteChanges reports the changes to the specified Writer
func WriteChanges(changes []diff.ChangeLogEntry, w io.Writer) error {
	reportChanges := make(diff.ChangeLogEntries, len(changes))

	for index, change := range changes {
		if change.ChangeType == diff.NoChange {
			continue
		}
		copiedChange := change
		switch {
		case copiedChange.ChangeType == diff.Deleted:
			if copiedChange.From.Kind != yaml.ScalarNode {
				copiedChange.From = nil
			}
		case copiedChange.ChangeType == diff.Added:
			if copiedChange.To.Kind != yaml.ScalarNode {
				copiedChange.To = nil
			}
		case copiedChange.ChangeType == diff.Moved:
			copiedChange.To = nil
			copiedChange.From = nil
		}
		reportChanges[index] = copiedChange
	}
	sort.Sort(reportChanges)
	changeReport, err := yaml.Marshal(reportChanges)
	if err != nil {
		return err
	}
	w.Write(changeReport)
	return nil
}
