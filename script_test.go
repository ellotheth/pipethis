/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/
package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ScriptTest struct {
	suite.Suite
}

func (s *ScriptTest) TestAuthorUsesSavedName() {
	script := Script{author: "foo"}

	author, err := script.Author()
	s.NoError(err)
	s.Equal("foo", author)
}

func (s *ScriptTest) TestAuthorParsesFileForPattern() {
	f, err := ioutil.TempFile("", "pipethis-test-")
	if err != nil {
		s.Fail("Failed creating the test file")
	}

	filename := f.Name()

	authorTest := func(expectedAuthor string, expectedError bool, input string) {
		ioutil.WriteFile(filename, []byte(input), os.ModePerm)

		script := Script{filename: filename}
		author, err := script.Author()
		s.Equal(expectedAuthor, author)
		if expectedError {
			s.Error(err)
		} else {
			s.NoError(err)
		}
	}

	for _, contents := range providerTestAuthorValid() {
		authorTest(contents[0], false, contents[1])
	}
	for _, contents := range providerTestAuthorInvalid() {
		authorTest(contents[0], true, contents[1])
	}

	os.Remove(filename)

}
func providerTestAuthorInvalid() [][]string {
	return [][]string{
		[]string{"", ``},
		[]string{"", `no author here yo`},
		[]string{"", `
stuff things
more stuff
# comments to ignore
// reasons bar_STUFF_123 is my name but it should only pick up the first word
# more comments
things and stuff
		`},
	}
}
func providerTestAuthorValid() [][]string {
	return [][]string{
		[]string{`bar`, `PIPETHIS_AUTHOR bar`},
		[]string{`bar`, `// PIPETHIS_AUTHOR bar         `},
		[]string{`bar`, `# PIPETHIS_AUTHOR bar         `},
		[]string{`bar`, `# PIPETHIS_AUTHOR		bar				   `},
		[]string{`bar_STUFF_123`, `
stuff things
more stuff
# comments to ignore
// reasons PIPETHIS_AUTHOR bar_STUFF_123 is my name but it should only pick up the first word
# more comments
things and stuff
		`},
		[]string{`bar_STUFF_123`, `
// PIPETHIS_AUTHOR bar_STUFF_123 should take this one
// PIPETHIS_AUTHOR other_author should ignore this one
		`},
	}
}

func TestScriptTest(t *testing.T) {
	suite.Run(t, new(ScriptTest))
}
