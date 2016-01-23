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
	"io"
	"os"

	"golang.org/x/crypto/openpgp"
)

// Signature represents the PGP signature to be verified against a key and
// Script.
type Signature struct {
	key      openpgp.KeyRing
	script   *Script
	filename string
	source   string
}

// NewSignature loads a key ring and Script into a new Signature.
func NewSignature(key openpgp.KeyRing, script *Script, source string) *Signature {
	sig := &Signature{key: key, script: script, source: source}
	sig.filename = script.Name() + ".sig"

	return sig
}

// Name is the name of the temporary file holding the signature.
func (s Signature) Name() string {
	return s.filename
}

// Source is the original location of the signature file. It defaults to
// <script source>.sig.
func (s *Signature) Source() string {
	if s.source != "" {
		return s.source
	}

	s.source = s.script.Source() + ".sig"

	return s.source
}

// Download saves the signature to a temporary file.
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

// Body opens Signature.Name() for reading, downloading it first if necessary.
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

// Verify checks Signature.Name() against the public key and script file, and
// returns an error if the signature cannot be verified.
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
