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

type SigTest struct {
	suite.Suite
}

func (s *SigTest) TestSourceUsesSaved() {
	sig := Signature{source: "foosig"}
	s.Equal("foosig", sig.Source())
}

func (s *SigTest) TestSourceBuildsFromScript() {
	sig := Signature{script: &Script{source: "scriptsource"}}
	s.Equal("scriptsource.sig", sig.Source())
}

func (s *SigTest) TestDownloadFailsWithoutSource() {
	sig := Signature{}
	s.Error(sig.Download())
}

func (s *SigTest) TestDownloadFailsWithNonexistentSource() {
	sig := Signature{source: "not-a-real-file"}
	s.Error(sig.Download())
}

func (s *SigTest) TestDownloadFailsWithoutDestinationName() {
	f, err := ioutil.TempFile("", "pipethis-test-")
	if err != nil {
		s.Fail("Failed creating the test file")
	}

	sig := Signature{source: f.Name()}
	defer os.Remove(sig.Source())

	s.Error(sig.Download())
}

func (s *SigTest) TestDownloadCopiesFile() {
	f, err := ioutil.TempFile("", "pipethis-test-")
	if err != nil {
		s.Fail("Failed creating the test file")
	}

	sig := Signature{source: f.Name(), filename: "destination"}
	defer os.Remove(sig.Source())
	defer os.Remove(sig.Name())

	expected := []byte("file contents, wooo")
	ioutil.WriteFile(sig.Source(), expected, os.ModePerm)

	s.NoError(sig.Download())

	actual, err := ioutil.ReadFile(sig.Name())
	s.Equal(expected, actual)
}

func (s *SigTest) TestBodyFailsWithoutFiles() {
	sig := Signature{}
	_, err := sig.Body()
	s.Error(err)
}

func TestSignatureTest(t *testing.T) {
	suite.Run(t, new(SigTest))
}
