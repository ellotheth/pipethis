package lookup

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/crypto/openpgp"
)

type KeyService interface {
	Matches(query string) ([]User, error)
	Key(user string) (openpgp.EntityList, error)
}

type User struct {
	Username    string
	Fingerprint string
	FullName    string
	Twitter     string
	GitHub      string
	HackerNews  string
	Reddit      string
	Sites       []string
}

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

	return s
}

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
	ring, err := service.Key(match.Username)
	if err != nil {
		return nil, err
	}
	log.Printf("Using %v (%v)", match.Username, ring[0].PrimaryKey.KeyIdShortString())

	return ring, nil
}
