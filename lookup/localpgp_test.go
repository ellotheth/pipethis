/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/

package lookup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LocalPGPTest struct {
	suite.Suite
}

func (s *LocalPGPTest) TestIsMatchMatchesOnFingerprint() {
	local := LocalPGPService{}
	user := User{Fingerprint: "foobar"}

	s.True(local.isMatch("oba", user))
	s.True(local.isMatch("foo", user))
	s.True(local.isMatch("bar", user))
	s.True(local.isMatch("FOOBAR", user))
}

func (s *LocalPGPTest) TestIsMatchMatchesOnEmails() {
	local := LocalPGPService{}
	user := User{Emails: []string{"foobar", "bizbaz", "THINGS"}}

	s.True(local.isMatch("oba", user))
	s.True(local.isMatch("foo", user))
	s.True(local.isMatch("bar", user))
	s.True(local.isMatch("FOOBAR", user))
	s.True(local.isMatch("zba", user))
	s.True(local.isMatch("thin", user))
}

func (s *LocalPGPTest) TestIsMatchFailsWithoutMatches() {
	local := LocalPGPService{}
	user := User{}

	s.False(local.isMatch("foo", user))
}

func (s *LocalPGPTest) TestGnupgHomeOverride() {
	os.Setenv("GNUPGHOME", "/foo")
	_, err := NewLocalPGPService()
	s.EqualError(err, "stat /foo/pubring.gpg: no such file or directory")
	os.Unsetenv("GNUPGHOME")
}

func (s *LocalPGPTest) TestBuildRingfileName() {
	cases := []struct {
		home      string
		gnupghome string
		expected  string
	}{
		{"/foo/", "", "/foo/.gnupg/pubring.gpg"},
		{"/foo", "", "/foo/.gnupg/pubring.gpg"},
		{"foo", "", "foo/.gnupg/pubring.gpg"},
		{"foo", "/things", "/things/pubring.gpg"},
		{"foo", "/things/", "/things/pubring.gpg"},
		{"foo", "things/", "things/pubring.gpg"},
		{"", "/things/", "/things/pubring.gpg"},
		{"", "", ".gnupg/pubring.gpg"},
	}

	for _, c := range cases {
		os.Setenv("HOME", c.home)
		os.Setenv("GNUPGHOME", c.gnupghome)
		local := LocalPGPService{}
		local.buildRingfileName()
		s.Equal(c.expected, local.ringfile)
	}
}

func TestLocalPGPTest(t *testing.T) {
	suite.Run(t, new(LocalPGPTest))
}
