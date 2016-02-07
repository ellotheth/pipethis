package lookup

import (
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

func TestLocalPGPTest(t *testing.T) {
	suite.Run(t, new(LocalPGPTest))
}
