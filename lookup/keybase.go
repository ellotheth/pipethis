/*
pipethis: Stop piping the internet into your shell
Copyright 2016 Ellotheth

Use of this source code is governed by the GNU Public License version 2
(GPLv2). You should have received a copy of the GPLv2 along with your copy of
the source. If not, see http://www.gnu.org/licenses/gpl-2.0.html.
*/

package lookup

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"

	"golang.org/x/crypto/openpgp"
)

type keybaseResponse struct {
	Status struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"status"`
	Users []struct {
		Details struct {
			Username    keybaseUserVal   `json:"username"`
			Fingerprint keybaseUserVal   `json:"key_fingerprint"`
			FullName    keybaseUserVal   `json:"full_name"`
			Twitter     keybaseUserVal   `json:"twitter"`
			GitHub      keybaseUserVal   `json:"github"`
			HackerNews  keybaseUserVal   `json:"hackernews"`
			Reddit      keybaseUserVal   `json:"reddit"`
			Sites       []keybaseUserVal `json:"websites"`
		} `json:"components"`
	} `json:"completions"`
}

type keybaseUserVal struct {
	Value string `json:"val"`
}

// KeybaseService implements the KeyService interface for https://keybase.io
type KeybaseService struct{}

func (k KeybaseService) lookup(query string) ([]byte, error) {
	if matches, _ := regexp.MatchString(`^[a-zA-Z0-9_\-\.]+$`, query); !matches {
		return nil, errors.New("Invalid user requested")
	}

	resp, err := http.Get("https://keybase.io/_/api/1.0/user/autocomplete.json?q=" + query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (k KeybaseService) parse(body []byte) (*keybaseResponse, error) {
	lookup := &keybaseResponse{}
	if err := json.Unmarshal(body, lookup); err != nil {
		return nil, err
	}
	if lookup.Status.Code != 0 {
		return nil, errors.New("Bad status code: " + lookup.Status.Name)
	}

	return lookup, nil
}

// Matches finds all the Keybase users that match query in any of their details
// (username, Twitter identity, Github identity, public key fingerprint,
// etc.). At most 10 matches will be found. If no matches are found, Matches
// returns an error.
func (k KeybaseService) Matches(query string) ([]User, error) {
	results, err := k.lookup(query)
	if err != nil {
		return nil, err
	}

	lookup, err := k.parse(results)
	if err != nil {
		return nil, err
	}

	matches := []User{}

	for _, match := range lookup.Users {
		user := User{
			Username:    match.Details.Username.Value,
			Fingerprint: match.Details.Fingerprint.Value,
			GitHub:      match.Details.GitHub.Value,
			Twitter:     match.Details.Twitter.Value,
			HackerNews:  match.Details.HackerNews.Value,
			Reddit:      match.Details.Reddit.Value,
		}

		for _, site := range match.Details.Sites {
			user.Sites = append(user.Sites, site.Value)
		}

		matches = append(matches, user)
	}

	return matches, nil
}

// Key finds the PGP public key for one Keybase user by Keybase username and
// returns the key ring representation of the key. If the Keybase username is
// invalid, or the key itself is missing or invalid, Key returns an error.
func (k KeybaseService) Key(user User) (openpgp.EntityList, error) {

	// I think I set this up to match Keybase's own username pattern. I think.
	if matches, _ := regexp.MatchString(`^[a-zA-Z0-9_\-\.]+$`, user.Username); !matches {
		return nil, errors.New("Invalid user requested")
	}

	resp, err := http.Get("https://keybase.io/" + user.Username + "/key.asc")
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	ring, err := openpgp.ReadArmoredKeyRing(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(ring) != 1 {
		return nil, errors.New("More than one key returned, not sure what to do")
	}

	return ring, nil
}
