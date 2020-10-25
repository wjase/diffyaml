package diff_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/wjase/diffyam/pkg/diff"
	"gopkg.in/yaml.v3"
)

func TestHash(t *testing.T) {
	var basicYamlReader = LinesInFile(t, fixturePath("basic.yaml"))
	basicYaml, err := ioutil.ReadAll(basicYamlReader)
	basicYamlStr := string(basicYaml)

	var doc1 yaml.Node
	err = yaml.Unmarshal(basicYaml, &doc1)
	assertThat(t, err, is.Nil())
	node := diff.HashNode(&doc1)
	assertThat(t, node, not(is.Nil()))
	nodeAgain := diff.HashNode(&doc1)
	assertThat(t, nodeAgain, not(is.Nil()))

	// hash same thing should get same result
	assertThat(t, nodeAgain.Hash, is.EqualTo(node.Hash))

	// now change something and hash should be different
	var doc2 yaml.Node
	newYaml := strings.Replace(basicYamlStr, "Denver", "Rio", 1)
	err = yaml.Unmarshal([]byte(newYaml), &doc2)
	assertThat(t, err, is.Nil())
	node2 := diff.HashNode(&doc2)
	assertThat(t, string(node2.Hash), not(is.EqualTo(string(node.Hash))))
	assertThat(t, len(node.Children), is.EqualTo(1))
	node = node.Children[0]
	assertThat(t, node, not(is.Nil()))
}
