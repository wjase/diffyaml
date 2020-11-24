package diff_test

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/stretchr/testify/require"
	"github.com/wjase/diffyaml/pkg/diff"
	"github.com/wjase/diffyaml/pkg/report"
)

var assertThat = then.AssertThat
var not = is.Not

var slasher = strings.NewReplacer("\\", "/")

type testCaseData struct {
	name          string
	oldSpec       string
	newSpec       string
	expectedError bool
	expectedFile  string
}

func makeTestCases(toFiles []string) []testCaseData {
	testCases := make([]testCaseData, 0, len(toFiles)+2)
	for _, eachFile := range toFiles {
		pathUnixSlashes := slasher.Replace(eachFile)
		namePart := path.Base(pathUnixSlashes)
		parentPart := path.Base(path.Dir(pathUnixSlashes))

		testCases = append(
			testCases, testCaseData{
				name:         parentPart + "/" + namePart,
				oldSpec:      strings.Replace(eachFile, "to.yaml", "from.yaml", 1),
				newSpec:      eachFile,
				expectedFile: strings.Replace(eachFile, "to.yaml", "diffs.yaml", 1),
			})
	}

	return testCases
}

func LinesInFile(t testing.TB, fileName string) io.ReadCloser {
	file, err := os.Open(fileName)
	assertThat(t, err, is.Nil())
	return file
}

func fixturePath(file string, parts ...string) string {
	return filepath.Join("..", "..", "fixtures", strings.Join(append([]string{file}, parts...), ""))
}

// TestDiffForVariousCombinations - computes the diffs for a number
// of scenarios and compares the computed diff with expected diffs
func TestDiffFiles(t *testing.T) {

	takeSnapshot := false
	_, present := os.LookupEnv("SNAPSHOT")
	if present {
		takeSnapshot = true
	}

	fileSets := []string{
		"simple/*.to.yaml",
		"swagger/*.to.yaml",
		// uncomment this to test individual cases
		// "simple/sequence-moved-item.to.yaml",
	}

	for _, suitePath := range fileSets {
		// pattern := fixturePath("simple/*.to.yaml")
		pattern := fixturePath(suitePath)
		// pattern := fixturePath("simple/sequence-added-changed-nesteditem.to.yaml")

		// To filter cases for debugging poke an individual case here eg "path", "enum" etc
		// see the test cases in fixtures/diff
		// Don't forget to remove it once you're done.
		// (There's a test at the end to check al cases were run)
		allToFiles, err := filepath.Glob(pattern)
		assertThat(t, err, is.Nil())
		assertThat(t, len(allToFiles), is.GreaterThan(0))

		testCases := makeTestCases(allToFiles)

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				changes, err := diff.GetYamlFileChanges(tc.oldSpec, tc.newSpec)

				if tc.expectedError {
					// edge cases with error
					assertThat(t, err, is.Not(is.Nil()))
					return
				}
				require.NoError(t, err)
				buffer := strings.Builder{}
				report.WriteChanges(changes, &buffer)
				changeReport := buffer.String()

				if takeSnapshot {
					ioutil.WriteFile(tc.expectedFile, []byte(changeReport), os.FileMode(0775))
				}
				assertThat(t, err, is.Nil())
				assertThat(t, buffer.String(), diff.MatchesFileContentsExcludingWhitepaces(tc.expectedFile))
			})
		}

	}

}
