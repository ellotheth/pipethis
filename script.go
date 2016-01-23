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

	io.Copy(file, body)
	script.filename = file.Name()

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
