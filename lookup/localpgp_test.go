/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/

package lookup

import (
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
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
	s.True(err.Error() == "stat /foo/pubring.gpg: no such file or directory")
	os.Unsetenv("GNUPGHOME")
}

func TestLocalPGPTest(t *testing.T) {
	suite.Run(t, new(LocalPGPTest))
}
