/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/clearsign"
)

// Script represents a shell script to be inspected, verified, and run.
type Script struct {
	author      string
	source      string
	filename    string
	clearsigned bool
}

// NewScript copies the shell script specified in location (which may be local
// or remote) to a temporary file and loads it into a Script.
func NewScript(location string) (*Script, error) {
	script := &Script{source: location}

	body, err := getFile(location)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	file, err := ioutil.TempFile("", "pipethis-")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	script.filename = file.Name()

	contents, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	contents, err = script.detachSignature(contents)
	if err != nil {
		return nil, err
	}

	_, err = file.Write(contents)
	if err != nil {
		return nil, err
	}

	return script, nil
}

func (s *Script) detachSignature(contents []byte) ([]byte, error) {
	block, _ := clearsign.Decode(contents)

	// if the signature is not attached, return the contents without
	// modification
	if block == nil {
		return contents, nil
	}

	s.clearsigned = true

	// get the raw script, without the signature or PGP headers, and without
	// the CRs
	contents = bytes.Replace(block.Bytes, []byte{0x0d}, nil, -1)

	// create a file for the armored signature
	sig, err := os.Create(s.filename + ".sig")
	if err != nil {
		return nil, err
	}
	defer sig.Close()

	// write the armored signature
	sigWriter, err := armor.Encode(sig, "PGP SIGNATURE", block.ArmoredSignature.Header)
	if err != nil {
		return nil, err
	}
	defer sigWriter.Close()
	io.Copy(sigWriter, block.ArmoredSignature.Body)

	return contents, nil
}

func (s Script) IsPiped() bool {
	return s.source == ""
}

// Name is the name of the temporary file holding the shell script.
func (s Script) Name() string {
	return s.filename
}

// Source is the original location of the shell script (local or remote).
func (s Script) Source() string {
	return s.source
}

// Body opens Script.Name() for reading.
func (s Script) Body() (ReadSeekCloser, error) {
	return os.Open(s.Name())
}

// Author parses Script.Body() for the PIPETHIS_AUTHOR token, and saves it if
// it's found.
func (s *Script) Author() (string, error) {
	if s.author != "" {
		return s.author, nil
	}

	file, err := s.Body()
	if err != nil {
		return "", err
	}
	defer file.Close()

	if author := parseToken(`.*PIPETHIS_AUTHOR\s+(\w+)`, file); author != "" {
		s.author = author

		return s.author, nil
	}

	return "", errors.New("Author not found")
}

// Run creates a new process, running Script.Name() with target and any
// additional arguments from the command line. It returns the result of the
// process.
func (s Script) Run(target string, args ...string) error {
	log.Println("Running", s.Name(), "with", target)

	// the first argument is the script source location. replace it with the
	// temporary filename.
	args[0] = s.Name()

	cmd := exec.Command(target, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s Script) Echo() {
	log.Println("Sending", s.Name(), "to STDOUT for more processing")
	body, _ := s.Body()
	defer body.Close()

	io.Copy(os.Stdout, body)
}

// Inspect checks whether an inspection was requested, and sends Script.Name()
// to editor if so. When editor exits, Inspect prompts the user to continue
// processing, and returns true to continue or false to stop.
func (s Script) Inspect(inspect bool, editor string) bool {
	if !inspect || s.IsPiped() {
		return true
	}

	log.Println("Opening", s.Name(), "in", editor)

	cmd := exec.Command(editor, s.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()

	runScript := "y"
	fmt.Print("Continue processing ", s.Name(), "? (Y/n) ")
	fmt.Scanf("%s", &runScript)

	return strings.ToLower(runScript) == "y"
}
