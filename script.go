/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/
package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Script represents a shell script to be inspected, verified, and run.
type Script struct {
	author   string
	source   string
	filename string
	// piped if set signifies that the source is from stdin
	piped bool
}

// NewScript copies the shell script specified in location (which may be local
// or remote) to a temporary file and loads it into a Script.
func NewScript(location string) (script *Script, err error) {
	body, err := getFile(location)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	script, err = fnewScript(body)
	if err != nil {
		return
	}

	script.source = location
	return
}

func FNewScript(r io.Reader) (script *Script, err error) {
	script, err = fnewScript(r)
	if err != nil {
		return
	}
	script.piped = true
	return
}

func fnewScript(r io.Reader) (*Script, error) {
	// TODO: Detect if a reader is pipe-like
	// ie like a named piped that could infinitely
	// hang if content isn't read from it.
	if r == nil {
		return nil, fmt.Errorf("script body is nil")
	}

	file, err := ioutil.TempFile("", "pipethis-")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	n, err := io.Copy(file, r)
	if n < 1 && err != nil {
		return nil, err
	}

	script := &Script{
		filename: file.Name(),
	}

	return script, nil
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
	args[0] = s.Name()

	cmd := exec.Command(target, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Inspect checks whether an inspection was requested, and sends Script.Name()
// to editor if so. When editor exits, Inspect prompts the user to continue
// processing, and returns true to continue or false to stop.
func (s Script) Inspect(inspect bool, editor string) bool {
	if !inspect {
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

// Piped tells if a script was passed in from stdin
func (s Script) Piped() bool {
	return s.piped
}
