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

func (s *ScriptTest) TestDetachSigReturnsUnsignedContents() {
	script := Script{}

	contents, err := script.detachSignature([]byte("foo"))
	s.NoError(err)
	s.Equal([]byte("foo"), contents)
}

func (s *ScriptTest) TestDetachSigReturnsContentsWithoutSig() {
	expectedContents := `# PIPETHIS_AUTHOR 5AA6F296

echo this is my file and there will be one match`
	expectedSig := `-----BEGIN PGP SIGNATURE-----

wsFcBAEBCAAQBQJW7YbQCRALxruWWqbylgAAAwcQACrkjS8FVy6T5+pdrcYQToo3
41ppFNARFLor94lgeDzzuR4XkfNR4Cms8saceKTLfkan5nxnS/w6c/koFPskDHyS
2mZL0dDia7KUNMPt2rZGC6CLkKXs83eWAFZr4S2Y4gUYQGatAzvhTc9kgFtPRyBn
DA6FT4zj5JK4CcmxHITHeKokCwOdje80kGYlX2u3e5bqQpTbKLQ2oLMMZFfEXrDJ
Y3QOOKmBBKrS1TXVBReauojXVNNjADPZHYQzGoxgZ0GjTESkEjzrCjQnnTMcIN4C
nlskH9q28xyeDWBj+H7gNOpQZ2B3fs0cUs05Ucce/xBZeHaXqaW3GFmfdCbv1J9A
CjgVMGojAjTZf47y1mmHL9yh9gkXaLTYyO37MNku+cR0ntKIi3VyIHopiljYPGOG
r3EFxZOHg40QalMezFIfUG0S2MLpJ9+d5cvgdzHHHZYeL49L17U6eePnbt++xmGy
RrnWb1C/OXxYfCveB42v1/gg9novYYZ8/n/OLCsOL37v+b8rwEjNufmv+7G7DqOU
ejljGRd27WQSBaYQMovWGpgLmMyCiW6wnUbYFivlOTcnMvOnRsBXqCJt0jdpFEBp
hfZr8sPYSgWIDnkt7DWIwd8/eap5mgMkC5j+Q81Lcv01OfDmppRlMWRf+a4BpHFd
FiV3SWOrHc2hIbLugeNf
=tnov
-----END PGP SIGNATURE-----`

	script := Script{filename: "fixtures/signed.attached"}
	sigFile := script.filename + ".sig"

	body, err := getLocal(script.filename)
	s.NoError(err)

	contents, err := ioutil.ReadAll(body)
	body.Close()
	s.NoError(err)

	contents, err = script.detachSignature(contents)
	s.NoError(err)
	s.True(script.clearsigned)
	s.Equal([]byte(expectedContents), contents)

	// check the (now detached) signature contents
	body, err = getLocal(sigFile)
	s.NoError(err)

	contents, err = ioutil.ReadAll(body)
	body.Close()
	s.NoError(err)

	s.Equal([]byte(expectedSig), contents)

	err = os.Remove(sigFile)
	s.NoError(err)
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
		{"", ``},
		{"", `no author here yo`},
		{"", `
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
		{`bar`, `PIPETHIS_AUTHOR bar`},
		{`bar`, `// PIPETHIS_AUTHOR bar         `},
		{`bar`, `# PIPETHIS_AUTHOR bar         `},
		{`bar`, `# PIPETHIS_AUTHOR		bar				   `},
		{`bar_STUFF_123`, `
stuff things
more stuff
# comments to ignore
// reasons PIPETHIS_AUTHOR bar_STUFF_123 is my name but it should only pick up the first word
# more comments
things and stuff
		`},
		{`bar_STUFF_123`, `
// PIPETHIS_AUTHOR bar_STUFF_123 should take this one
// PIPETHIS_AUTHOR other_author should ignore this one
		`},
	}
}

func TestScriptTest(t *testing.T) {
	suite.Run(t, new(ScriptTest))
}
