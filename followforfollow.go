package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
)

func loadToken() (string, error) {
	gh_pat, err := os.ReadFile(os.ExpandEnv("$KEYS/gh-PAT"))
	if err != nil {
		gh_pat = []byte(os.Getenv("GH_PAT"))
	}
	token := strings.TrimSpace(string(gh_pat))
	if token == "" {
		return token, fmt.Errorf("GitHub token not found in environment variable GH_PAT or in $KEYS/gh-PAT file")
	}
	return token, nil
}

type GitHubClient struct {
	http  *http.Client
	token string
}

func (c *GitHubClient) fetch_user_list(url string) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var users []map[string]any
	var users_array []string
	raw_users := string(body)
	json.Unmarshal([]byte(raw_users), &users)

	for _, user_obj := range users {
		user_name := user_obj["login"].(string)
		users_array = append(users_array, user_name)
	}

	return users_array, nil
}

func (c *GitHubClient) follow_user(username string) error {
	follow_req, follow_req_err := http.NewRequest("PUT", "https://api.github.com/user/following/"+username, nil)
	if follow_req_err != nil {
		return follow_req_err
	}
	follow_req.Header.Set("Content-Type", "application/json")
	follow_req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))

	follow_resp, follow_resp_err := c.http.Do(follow_req)
	if follow_resp_err != nil {
		return follow_resp_err
	}
	defer follow_resp.Body.Close()

	if follow_resp.StatusCode == 204 {
		fmt.Println("Followed back:", username)
	} else {
		fmt.Println("Failed to follow back:", username, "Status Code:", follow_resp.StatusCode)
	}
	return nil
}

func (c *GitHubClient) unfollow_user(username string) error {
	unfollow_req, unfollow_req_err := http.NewRequest("DELETE", "https://api.github.com/user/following/"+username, nil)
	if unfollow_req_err != nil {
		return unfollow_req_err
	}
	unfollow_req.Header.Set("Content-Type", "application/json")
	unfollow_req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))

	unfollow_resp, unfollow_resp_err := c.http.Do(unfollow_req)
	if unfollow_resp_err != nil {
		return unfollow_resp_err
	}
	defer unfollow_resp.Body.Close()

	if unfollow_resp.StatusCode == 204 {
		fmt.Println("Unfollowed:", username)
	} else {
		fmt.Println("Failed to unfollow:", username, "Status Code:", unfollow_resp.StatusCode)
	}
	return nil
}

func GetGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		http:  &http.Client{},
		token: token,
	}
}

func main() {
	// GITHUB API URLS
	url_gh_followers := "https://api.github.com/user/followers?per_page=100"
	url_gh_following := "https://api.github.com/user/following?per_page=100"
	token, err := loadToken() // load token from file or env var
	if err != nil {
		log.Fatal(err)
	}
	client := GetGitHubClient(token)

	followers_array, _ := client.fetch_user_list(url_gh_followers)
	following_array, _ := client.fetch_user_list(url_gh_following)

	fmt.Println("Followers:", len(followers_array))
	fmt.Println("Following:", len(following_array))

	sort.Strings(followers_array)
	sort.Strings(following_array)

	// Determine who is following me that I am not following back
	follow_back_array := []string{}
	for _, user_that_follows_me := range followers_array {
		i_follow_this_user := false
		for _, user_that_i_follow := range following_array {
			if !i_follow_this_user && user_that_follows_me == user_that_i_follow {
				i_follow_this_user = true
				break
			}
		}
		if !i_follow_this_user && user_that_follows_me != "sphinxzerd" { // exclude bot accounts
			fmt.Println("-", user_that_follows_me)
			follow_back_array = append(follow_back_array, user_that_follows_me)
		}
	}

	// follow back
	for _, follow_back := range follow_back_array {
		err := client.follow_user(follow_back)
		if err != nil {
			panic(err)
		}
	}

	// Determine who I am following that is not following me back
	unfollow_array := []string{}
	fmt.Println("\nUsers not following back:")
	for _, user_that_i_follow := range following_array {
		user_follows_me := false
		for _, user_that_follows_me := range followers_array {
			if !user_follows_me && user_that_i_follow == user_that_follows_me {
				user_follows_me = true
				break
			}
		}
		if !user_follows_me {
			fmt.Println("-", user_that_i_follow)
			unfollow_array = append(unfollow_array, user_that_i_follow)
		}
	}

	exempt_users_array := []string{"BenjaminX", "academind", "angelabauer", "mschwarzmueller"}
	for _, unfollow := range unfollow_array {
		exempted := slices.Contains(exempt_users_array, unfollow)
		if exempted {
			fmt.Println("Exempted from unfollowing:", unfollow)
			continue
		}
		err := client.unfollow_user(unfollow)
		if err != nil {
			panic(err)
		}
	}
}
