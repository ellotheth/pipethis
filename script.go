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

type Script struct {
	author   string
	source   string
	filename string
}

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

func (s Script) Name() string {
	return s.filename
}

func (s Script) Source() string {
	return s.source
}

func (s Script) Body() (ReadSeekCloser, error) {
	return os.Open(s.Name())
}

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

func (s Script) Run(target string, args ...string) error {
	args[0] = s.Name()

	cmd := exec.Command(target, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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
	fmt.Print("Continue to signature verification for ", s.Name(), " now? (Y/n) ")
	fmt.Scanf("%s", &runScript)

	return strings.ToLower(runScript) == "y"
}
