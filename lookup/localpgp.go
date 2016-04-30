/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/

package lookup

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/crypto/openpgp"
)

// LocalPGPService implements the KeyService interface for a local GnuPG
// public keyring.
type LocalPGPService struct {
	ringfile string
	ring     openpgp.EntityList
}

// NewLocalPGPService creates a new LocalPGPService if it finds a local
// public keyring; otherwise it bails.
func NewLocalPGPService() (*LocalPGPService, error) {

	home := os.Getenv("HOME")

	// Check if an override for GNUPG home is set
	if os.Getenv("GNUPGHOME") != "" {
		home = os.Getenv("GNUPGHOME")
	}

	ringfile := path.Join(home, ".gnupg", "pubring.gpg")

	info, err := os.Stat(ringfile)
	if err != nil || info.Size() == 0 {
		return nil, err
	}

	return &LocalPGPService{ringfile: ringfile}, nil
}

// Ring loads the local public keyring so LocalPGPService can use it later. If
// it's already been loaded, Ring returns the existing version.
func (l *LocalPGPService) Ring() openpgp.EntityList {
	if l.ring != nil {
		return l.ring
	}

	reader, err := os.Open(l.ringfile)
	if err != nil {
		return nil
	}
	defer reader.Close()

	ring, err := openpgp.ReadKeyRing(reader)
	if err != nil {
		return nil
	}

	return ring
}

// Matches finds all the public keys that have a fingerprint or identity (name
// and email address) that match query. If no matches are found, Matches
// returns an error.
func (l *LocalPGPService) Matches(query string) ([]User, error) {
	users := []User{}

	ring := l.Ring()
	if ring == nil {
		return nil, errors.New("No key ring loaded")
	}

	// this is why LocalPGPService.ring has to be an EntityList instead of the
	// more generic KeyRing: can't iterate through the latter. Botheration.
	for _, key := range ring {
		user := User{
			Fingerprint: key.PrimaryKey.KeyIdString(),
		}

		for name := range key.Identities {
			user.Emails = append(user.Emails, name)
		}

		if l.isMatch(query, user) {
			users = append(users, user)
		}
	}

	if len(users) == 0 {
		return nil, errors.New("No matches")
	}

	return users, nil
}

func (l LocalPGPService) isMatch(query string, user User) bool {
	if strings.Contains(strings.ToUpper(user.Fingerprint), strings.ToUpper(query)) {
		return true
	}

	for _, email := range user.Emails {
		if strings.Contains(strings.ToUpper(email), strings.ToUpper(query)) {
			return true
		}
	}

	return false
}

// Key gets the PGP public key from the local public keyring for a user's
// fingerprint and returns the keyRing representation. If the fingerprint is
// invalid or more than one public key is found, Key returns an error.
func (l *LocalPGPService) Key(user User) (openpgp.EntityList, error) {
	id, err := strconv.ParseUint(user.Fingerprint, 16, 64)
	if err != nil {
		return nil, err
	}

	ring := l.Ring()
	keys := ring.KeysById(id)
	if len(keys) != 1 {
		return nil, errors.New("More than one key returned, not sure what to do")
	}

	list := []*openpgp.Entity{keys[0].Entity}

	return list, nil
}
