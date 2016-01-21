package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorUsesSavedName(t *testing.T) {
	s := Script{author: "foo"}

	author, err := s.Author()
	assert.NoError(t, err)
	assert.Equal(t, "foo", author)
}

func TestAuthorParsesFileForPattern(t *testing.T) {
	f, err := ioutil.TempFile("", "pipethis-test-")
	if err != nil {
		assert.Fail(t, "Couldn't open a temporary file")
	}
	filename := f.Name()

	authorTest := func(expectedAuthor string, expectedError bool, input string) {
		ioutil.WriteFile(filename, []byte(input), os.ModePerm)

		s := Script{filename: filename}
		author, err := s.Author()
		assert.Equal(t, expectedAuthor, author)
		if expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
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
