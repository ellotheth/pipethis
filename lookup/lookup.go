/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/

// Package lookup provides services that can look up identities and public keys
// for script authors.
package lookup

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/crypto/openpgp"
)

// KeyService defines the interface for third-party identity verification and
// public key services, like Keybase or Onename.
//
// Matches gets a list of identities that match a query string from the
// KeyService.
//
// Key gets the PGP public key for one user.
type KeyService interface {
	Matches(query string) ([]User, error)
	Key(user User) (openpgp.EntityList, error)
}

// User represents an author's identity.
type User struct {
	Username    string
	Fingerprint string
	FullName    string
	Twitter     string
	GitHub      string
	HackerNews  string
	Reddit      string
	Sites       []string
	Emails      []string
}

// String returns a representation of all the User's identity details.
func (u User) String() string {
	format := "%15s: %s\n"
	s := ""

	s = s + fmt.Sprintf(format, "Identifier", u.Username)
	s = s + fmt.Sprintf(format, "Twitter", u.Twitter)
	s = s + fmt.Sprintf(format, "Github", u.GitHub)
	s = s + fmt.Sprintf(format, "Hacker News", u.HackerNews)
	s = s + fmt.Sprintf(format, "Reddit", u.Reddit)
	s = s + fmt.Sprintf(format, "Fingerprint", u.Fingerprint)

	for _, site := range u.Sites {
		s = s + fmt.Sprintf(format, "Site", site)
	}

	for _, email := range u.Emails {
		s = s + fmt.Sprintf(format, "Email", email)
	}

	return s
}

// chooseMatch prints all the matches provided, prompts for a choice, and
// returns the chosen match.
func chooseMatch(matches []User) (User, error) {
	log.Println("I found", len(matches), "results:")
	fmt.Println()
	for idx, user := range matches {
		fmt.Printf("%d:\n\n", idx)
		fmt.Println(user)
		fmt.Println()
	}

	response := "q"
	fmt.Print("\nEnter the number to use, or 'q' to cancel: ")
	fmt.Scanf("%s", &response)

	if strings.ToLower(response) == "q" {
		return User{}, errors.New("No match selected")
	}

	n, err := strconv.Atoi(response)
	if err != nil {
		return User{}, err
	}

	if n < 0 || n >= len(matches) {
		return User{}, errors.New("Invalid match selected")
	}
	fmt.Println()

	return matches[n], nil
}

// Key looks up an author query in the provided KeyService, and prompts for a
// choice of matches. It returns an error if no matches were found, if no match
// was chosen, or if no PGP public was found.
func Key(service KeyService, query string) (openpgp.KeyRing, error) {
	// get possible matches from the key service
	matches, err := service.Matches(query)
	if err != nil {
		return nil, err
	}

	// verify that the author is who the user was expecting by showing all the
	// details (twitter handle, github handle, websites, etc.)
	match, err := chooseMatch(matches)
	if err != nil {
		return nil, err
	}

	// get the public key for the selected author
	ring, err := service.Key(match)
	if err != nil {
		return nil, err
	}
	log.Printf("Using %v (%v)", match.Username, ring[0].PrimaryKey.KeyIdShortString())

	return ring, nil
}
