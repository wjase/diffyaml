package diff

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/corbym/gocrest"
)

//MatchesFileContentsExcludingWhitepaces - load the file and compare to actual string
//Returns a gocrest matcher.
func MatchesFileContentsExcludingWhitepaces(expectedFilePath interface{}) *gocrest.Matcher {
	match := new(gocrest.Matcher)
	match.Describe = fmt.Sprintf("Matches contents of file <%v>", expectedFilePath)
	match.Matches = func(actual interface{}) bool {
		f, err := os.Open(expectedFilePath.(string))
		if err != nil {
			match.Reasonf("%v", err)
			match.Actual = "not compared"
			return false
		}
		expectedStr, err := ioutil.ReadAll(f)
		if err != nil {
			match.Reasonf("%v", err)
			match.Actual = "not compared"
			return false
		}

		match.Describe = match.Describe + "\n<\n" + string(expectedStr) + ">\n"
		r := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "")
		switch v := actual.(type) {
		case []byte:
			match.Actual = "\n" + string(v)
		case string:
			match.Actual = "\n" + v
		case fmt.Stringer:
			match.Actual = "\n" + v.String()
		default:
			match.Actual = fmt.Sprintf("Unable to compare %v", v)
		}
		str1 := r.Replace(string(expectedStr))
		str2 := r.Replace(match.Actual)
		return str1 == str2
	}

	return match
}
