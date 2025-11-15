package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
)

func main() {
	// url_gh_test := "https://api.github.com/octocat"
	url_gh_followers := "https://api.github.com/user/followers?per_page=100"
	url_gh_following := "https://api.github.com/user/following?per_page=100"
	gh_pat, err := os.ReadFile(os.ExpandEnv("$KEYS/gh-PAT"))
	if err != nil {
		gh_pat = []byte(os.Getenv("GH_PAT"))
	}
	token := strings.TrimSpace(string(gh_pat))
	if token == "" {
		panic("GitHub token not found in environment variable GH_PAT or in $KEYS/gh-PAT file")
	}
	// GITHUB API REQ
	req, _ := http.NewRequest("GET", url_gh_followers, nil)
	req1, _ := http.NewRequest("GET", url_gh_following, nil)
	// "GET" = HTTP method
	// nil = no request body (GET requests don't have bodies)

	// SET HEADER AUTHORIZATION
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token)) // "Hey, GH, here's my token"
	req1.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	// ^ MUST BE: "token <token>"

	// SET HEADER ACCEPTANCE
	// Most REST APIs use: `req.Header.Set("Accept", "application/json")`
	// GitHub API uses custom media types to version their APIs:
	// req.Header.Set("Accept", "application/vnd.github.v3+json") <- github docs specify this
	// vnd = vendor-specific
	//   GitHub API Version History
	// 	- https://docs.github.com/en/rest
	//  - v1 & v2: Deprecated/removed years ago
	//  - v3: Current REST API (most used)
	//  - v4: GraphQL API (different syntax, very flexible/complex)
	//
	// MIME ("Content-Type") structuring = type/subtype
	// types: [application[json, xml, pdf, zip, octet-stream...], text[plain, html, css, csv],
	// more types:image[png, jpg, gif, svg], audio[mpeg, wav], video/[mpeg, mp4, webm], multipart/form-data]

	// Most common
	req.Header.Set("Accept", "application/json")
	req1.Header.Set("Accept", "application/json")

	// Create an HTTP client which is the actual requester
	client := &http.Client{}
	resp, err := client.Do(req)
	resp1, err1 := client.Do(req1)
	if err != nil {
		panic(err)
	}
	if err1 != nil {
		panic(err1)
	}
	defer resp.Body.Close() // Tells to close the response body when done, but not immediately
	defer resp1.Body.Close()
	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	body1, err1 := io.ReadAll(resp1.Body)
	if err1 != nil {
		panic(err1)
	}

	var followers []map[string]any
	var followers_array []string
	raw_followers := (string(body))
	json.Unmarshal([]byte(raw_followers), &followers)

	var following []map[string]any
	var following_array []string
	raw_following := (string(body1))
	json.Unmarshal([]byte(raw_following), &following)

	for _, follower_obj := range followers {
		follower_name := follower_obj["login"].(string)
		followers_array = append(followers_array, follower_name)
	}

	for _, following_obj := range following {
		following_name := following_obj["login"].(string)
		following_array = append(following_array, following_name)
	}

	fmt.Println("Followers:", len(followers_array))
	fmt.Println("Following:", len(following_array))

	sort.Strings(followers_array)
	sort.Strings(following_array)

	// fmt.Println("\nFollowers List:")
	// for _, follower := range followers_array {
	// 	fmt.Println("-", follower)
	// }

	// fmt.Println("\nFollowing List:")
	// for _, following := range following_array {
	// 	fmt.Println("-", following)
	// }

	follow_back_array := []string{}
	//unfollow_array := []string{}
	// Determine who is following me that I am not following back
	for _, follower := range followers_array {
		follower_found := false
		for _, following := range following_array {
			if !follower_found && follower == following {
				follower_found = true
				break
			}
		}
		if !follower_found && follower != "sphinxzerd" { // exclude bot accounts
			fmt.Println("-", follower)
			follow_back_array = append(follow_back_array, follower)
		}
	}

	// follow back
	for _, follow_back := range follow_back_array {
		follow_req, follow_req_err := http.NewRequest("PUT", "https://api.github.com/user/following/"+follow_back, nil)
		follow_req.Header.Set("Content-Type", "application/json")
		if follow_req_err != nil {
			panic(follow_req_err)
		}
		follow_req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		follow_resp, follow_resp_err := client.Do(follow_req)
		if follow_resp_err != nil {
			panic(follow_resp_err)
		}
		defer follow_resp.Body.Close()
		if follow_resp.StatusCode == 204 {
			fmt.Println("Followed back:", follow_back)
		} else {
			fmt.Println("Failed to follow back:", follow_back, "Status Code:", follow_resp.StatusCode)
		}
	}

	unfollow_array := []string{}
	// Determine who I am following that is not following me back
	fmt.Println("\nUsers not following back:")
	for _, following := range following_array {
		following_found := false
		for _, follower := range followers_array {
			if !following_found && following == follower {
				following_found = true
				break
			}
		}
		if !following_found {
			fmt.Println("-", following)
			unfollow_array = append(unfollow_array, following)
		}
	}

	exempt_users_array := []string{"BenjaminX", "academind", "angelabauer", "mschwarzmueller"}
	// unfollow
	for _, unfollow := range unfollow_array {
		exempted := slices.Contains(exempt_users_array, unfollow)
		if exempted {
			fmt.Println("Exempted from unfollowing:", unfollow)
			continue
		}
		unfollow_req, unfollow_req_err := http.NewRequest("DELETE", "https://api.github.com/user/following/"+unfollow, nil)
		unfollow_req.Header.Set("Content-Type", "application/json")
		if unfollow_req_err != nil {
			panic(unfollow_req_err)
		}
		unfollow_req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		unfollow_resp, unfollow_resp_err := client.Do(unfollow_req)
		if unfollow_resp_err != nil {
			panic(unfollow_resp_err)
		}
		defer unfollow_resp.Body.Close()
		if unfollow_resp.StatusCode == 204 {
			fmt.Println("Unfollowed:", unfollow)
		} else {
			fmt.Println("Failed to unfollow:", unfollow, "Status Code:", unfollow_resp.StatusCode)
		}
	}
}
