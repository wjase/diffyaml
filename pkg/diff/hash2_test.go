package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/corbym/gocrest/then"
)

var assertThat = then.AssertThat

func TestAllDiffs(t *testing.T) {
	makeDiffs()
}

func makeDiffs() error {
	spec1Path := FixturePath("allyaml/allfields.yml")
	spec1, err := ReadYAMLFile(spec1Path)
	if err != nil {
		return err
	}
	hashNode := HashNode(spec1)

	WalkNode(hashNode, func(n *HashedNode) bool {
		if len(n.Children) == 0 {
			fmt.Printf("%s\n", n.GetPath())
		}
		return true
	})

	return nil
}

func WalkNode(eachNode *HashedNode, visitorFn func(node *HashedNode) bool) bool {
	shouldContinue := visitorFn(eachNode)
	if shouldContinue {
		for _, child := range eachNode.Children {
			if shouldContinue := WalkNode(child, visitorFn); !shouldContinue {
				return false
			}
		}
	}
	return shouldContinue

}

func FixturePath(file string, parts ...string) string {
	path, _ := os.Getwd()
	fmt.Printf("%s\n", path)
	rootPath := strings.Split(path, "pkg")[0]
	return filepath.Join(rootPath, "fixtures", strings.Join(append([]string{file}, parts...), ""))
}
