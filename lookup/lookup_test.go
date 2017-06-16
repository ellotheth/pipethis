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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LookupTest struct {
	home string
	suite.Suite
}

func (s *LookupTest) SetupSuite() {
	s.home = os.Getenv("HOME")
	os.Setenv("HOME", strings.TrimRight(os.TempDir(), "/"))
}

func (s *LookupTest) TeardownSuite() {
	os.Setenv("HOME", s.home)
}

func (s *LookupTest) TestUserString() {
	expected := `     Identifier: me
        Twitter: 
         Github: me
    Hacker News: 
         Reddit: 
    Fingerprint: foobitybar
           Site: google.com
           Site: yahoo.com
          Email: me@myself
          Email: you@yourself
`
	user := User{
		Username:    "me",
		Fingerprint: "foobitybar",
		GitHub:      "me",
		Sites:       []string{"google.com", "yahoo.com"},
		Emails:      []string{"me@myself", "you@yourself"},
	}

	s.Equal(expected, user.String())
}

func (s *LookupTest) TestNewKeyServiceBailsOnUnrecognizedType() {
	service, err := NewKeyService("foo", false)
	s.Nil(service)
	s.EqualError(err, "Unrecognized key service")
}

func (s *LookupTest) TestNewKeyServiceAcceptsKeybaseWithoutPipe() {
	service, err := NewKeyService("keybase", false)

	s.NoError(err)
	s.IsType(&KeybaseService{}, service)
}

func (s *LookupTest) TestNewKeyServiceForcesLocalWithPipe() {
	_, err := NewKeyService("keybase", true)
	s.Error(err)

	perr, ok := err.(*os.PathError)
	s.True(ok)
	s.Equal(os.Getenv("HOME")+"/.gnupg/pubring.gpg", perr.Path)
}

func (s *LookupTest) TestChooseSingleMatchBailsWithoutMatches() {
	user, err := chooseSingleMatch([]User{})

	s.Error(err)
	s.Equal(User{}, user)
}

func (s *LookupTest) TestChooseSingleMatchBailsWithMoreThanOneMatch() {
	user, err := chooseSingleMatch([]User{{}, {}})

	s.Error(err)
	s.Equal(User{}, user)
}

func (s *LookupTest) TestChooseSingleMatchReturnsSingleMatch() {
	user, err := chooseSingleMatch([]User{{Username: "foo"}})

	s.NoError(err)
	s.Equal("foo", user.Username)
}

func TestLookupTest(t *testing.T) {
	suite.Run(t, new(LookupTest))
}
