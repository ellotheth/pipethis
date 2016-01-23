package main

import (
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/openpgp"
)

type Signature struct {
	key      openpgp.KeyRing
	script   *Script
	filename string
	source   string
}

func NewSignature(key openpgp.KeyRing, script *Script, source string) *Signature {
	sig := &Signature{key: key, script: script, source: source}
	sig.filename = script.Name() + ".sig"

	return sig
}

func (s Signature) Name() string {
	return s.filename
}

func (s *Signature) Source() string {
	if s.source != "" {
		return s.source
	}

	s.source = s.script.Source() + ".sig"

	return s.source
}

func (s *Signature) Download() error {
	source := s.Source()
	if source == "" {
		return errors.New("Signature source location not found")
	}

	body, err := getFile(source)
	if err != nil {
		return err
	}
	defer body.Close()

	file, err := os.Create(s.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	io.Copy(file, body)

	return nil
}

func (s *Signature) Body() (ReadSeekCloser, error) {
	info, err := os.Stat(s.Name())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) || info.Size() == 0 {
		err := s.Download()
		if err != nil {
			return nil, err
		}
	}

	return os.Open(s.Name())
}

func (s *Signature) Verify() error {
	signed, err := s.script.Body()
	if err != nil {
		return err
	}
	defer signed.Close()

	signature, err := s.Body()
	if err != nil {
		return err
	}
	defer signature.Close()

	if _, err := openpgp.CheckDetachedSignature(s.key, signed, signature); err == nil {
		return nil
	}

	signature.Seek(0, 0) // i'm sure there's a good reason i don't need to reset the script...
	if _, err := openpgp.CheckArmoredDetachedSignature(s.key, signed, signature); err == nil {
		return nil
	}

	return errors.New("Failed to verify signature")
}
