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

func (k KeybaseService) Key(user string) (openpgp.EntityList, error) {
	if matches, _ := regexp.MatchString(`^[a-zA-Z0-9_\-\.]+$`, user); !matches {
		return nil, errors.New("Invalid user requested")
	}

	resp, err := http.Get("https://keybase.io/" + user + "/key.asc")
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
