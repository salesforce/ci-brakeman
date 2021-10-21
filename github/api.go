/*
 * Copyright (c) 2021, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */

package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4"
	gitHttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// JWTToken holds the access token used to auth to Github
var JWTToken string

// GithubToken holds the access token for GitHub when not using a bearer token
var GithubToken string

const githubAPIHost = "https://api.github.com"

func makeGetRequest(path string) (resp []byte, statusCode int, err error) {
	return makeRequest(path, "GET", false, nil)
}

func makePostRequest(path string, data io.Reader) (resp []byte, statusCode int, err error) {
	return makeRequest(path, "POST", false, data)
}

func makeAuthRequest(path string) (resp []byte, statusCode int, err error) {
	return makeRequest(path, "POST", true, nil)
}

func makeRequest(path, method string, isAuth bool, data io.Reader) (resp []byte, statusCode int, err error) {
	var request *http.Request
	url := fmt.Sprintf("%s%s", githubAPIHost, path)
	request, err = http.NewRequest(method, url, data)

	if err != nil {
		//logger.Error(err)
		return
	}
	var response *http.Response

	// setting header to vnd.github.machine-man-preview since we are a github app
	// only for AUTH
	if isAuth {
		request.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")
		request.Header.Add("Authorization", "bearer "+JWTToken)
	} else {
		request.Header.Add("Accept", "application/vnd.github.symmetra-preview+json")
		request.Header.Add("Authorization", "token "+GithubToken)
	}

	response, err = (&http.Client{}).Do(request)

	if err != nil {
		return
	}

	defer response.Body.Close()

	if err != nil {
		return
	}

	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	// return status code so that api function knows what the outcome was
	statusCode = response.StatusCode
	return
}

// GetAccessToken returns a Access token for interacting with Github
func GetAccessToken(installationID string) (err error) {
	var accessToken TokenResponse

	path := fmt.Sprintf("/app/installations/%s/access_tokens", installationID)
	body, status, err := makeAuthRequest(path)

	if err != nil {
		return
	}

	if status != 201 {
		return fmt.Errorf("Auth failed with status code: %d", status)
	}

	if err = json.Unmarshal(body, &accessToken); err != nil {
		return
	}
	GithubToken = accessToken.Token
	return
}

// GetCommit returns a specific commit
// Github API docs: https://developer.github.com/v3/repos/commits/#get-a-single-commit
func GetCommit(owner, repo, sha string) (commit *RepoCommit, resp *Response, err error) {
	path := fmt.Sprintf("/repos/%v/%v/commits/%v", owner, repo, sha)
	data, status, err := makeGetRequest(path)

	if err != nil {
		return
	}

	if status != 200 {
		if err = json.Unmarshal(data, &resp); err != nil {
			return nil, nil, fmt.Errorf("Fetch failed with status code: %d and couldn't unmarshal response: %s", status, err)
		}
		return nil, resp, fmt.Errorf("Fetch failed with status code: %d", status)
	}

	if err = json.Unmarshal(data, &commit); err != nil {
		return nil, nil, fmt.Errorf("Couldn't unmarshal response: %s", err)
	}

	return
}

// GetContents returns the contents of a file
// Github API docs: https://developer.github.com/v3/repos/contents/#get-contents
func GetContents(owner, repo, path, ref string) (content *Content, resp *Response, err error) {
	// ref can be empty
	p := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		p = fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, ref)
	}
	data, status, err := makeGetRequest(p)

	if err != nil {
		return
	}

	if status != 200 {
		if err = json.Unmarshal(data, &resp); err != nil {
			return nil, nil, fmt.Errorf("Fetch failed with status code: %d and couldn't unmarshal response: %s", status, err)
		}
		return nil, resp, fmt.Errorf("Fetch failed with status code: %d", status)
	}

	if err = json.Unmarshal(data, &content); err != nil {
		return nil, nil, fmt.Errorf("Couldn't unmarshal response: %s", err)
	}

	return
}

// GetFileFromTree returns the Blob content of a given file from a tree in a repository
// https://developer.github.com/v3/git/trees/
// https://developer.github.com/v3/git/blobs/
func GetFileFromTree(owner, repo, fp, ref string) (content *Blob, resp *Response, err error) {

	// if path is a file in a sub-dir, we have to walk the tree to get to the file
	fpath, fname := path.Split(fp)
	downloadPath := ""

	if err != nil {
		return
	}
	var tree *Tree

	nxsha := ref // start walking from the user supplied ref
	for _, k := range strings.Split(fpath, "/") {
		// get the tree for current ref
		tree, resp, err = GetTree(owner, repo, nxsha)
		if err != nil {
			fmt.Println(resp)
		}
		// parse all entries in the tree to find the ref to
		// the path we are trying to walk
		for _, t := range tree.Entries {
			// check if we've found the ref for the next tree to walk
			if *t.Path == k {
				nxsha = *t.SHA
				break
			}
		}
	}
	// reached the end of the tree and now we want to
	// get the download path to the actual file to download
	for _, t := range tree.Entries {
		if *t.Path == fname {
			downloadPath = *t.URL
			break
		}
	}
	// replace prefix in the downloadPath since makeGetRequest prepends that by default
	downloadPath = strings.Replace(downloadPath, "https://api.github.com", "", -1)

	data, status, err := makeGetRequest(downloadPath)
	if err != nil {
		return
	}

	// unmarshal the data, this is a Blob. The blob contains the base64 encoded data of the file
	if status != 200 {
		if err = json.Unmarshal(data, &resp); err != nil {
			return nil, nil, fmt.Errorf("Fetch failed with status code: %d and couldn't unmarshal response: %s", status, err)
		}
		return nil, resp, fmt.Errorf("Fetch failed with status code: %d", status)
	}

	if err = json.Unmarshal(data, &content); err != nil {
		return nil, nil, fmt.Errorf("Couldn't unmarshal response: %s", err)
	}
	return
}

// GetTree returns the contents of a Tree on GitHub
// https://developer.github.com/v3/git/trees/
func GetTree(owner, repo, ref string) (tree *Tree, resp *Response, err error) {

	p := fmt.Sprintf("/repos/%s/%s/git/trees/%s", owner, repo, ref)

	data, status, err := makeGetRequest(p)

	if err != nil {
		return
	}

	if status != 200 {
		if err = json.Unmarshal(data, &resp); err != nil {
			return nil, nil, fmt.Errorf("Fetch failed with status code: %d and couldn't unmarshal response: %s", status, err)
		}
		return nil, resp, fmt.Errorf("Fetch failed with status code: %d", status)
	}

	if err = json.Unmarshal(data, &tree); err != nil {
		return nil, nil, fmt.Errorf("Couldn't unmarshal response: %s", err)
	}

	return
}

// GetPullRequestFiles gets the files for a particular pull request
func GetPullRequestFiles(owner, repo, pullNumber string) (pullReqResp []PullRequestFile, err error) {
	path := fmt.Sprintf("/repos/%v/%v/pulls/%v/files", owner, repo, pullNumber)
	url := fmt.Sprintf("%s%s", githubAPIHost, path)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+GithubToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &pullReqResp); err != nil {
		return nil, fmt.Errorf("Couldn't unmarshal response: %s", err)
	}

	fmt.Println("[GetPullRequestFiles]: response Status:", resp.Status, "repo: ", repo, "owner: ", owner, "pull request number: ", pullNumber)

	return
}

//PostCommentToGit posts comments to Pull requests
func PostCommentToGit(owner string, repo string, pullNumber string, commentBody string) (err error) {
	path := fmt.Sprintf("/repos/%v/%v/issues/%v/comments", owner, repo, pullNumber)

	var jsonStr = []byte(commentBody)
	url := fmt.Sprintf("%s%s", githubAPIHost, path)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "token "+GithubToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("[PostCommentToGit] response Status: ", resp.Status, "repo: ", repo, "owner: ", owner, "pull request number: ", pullNumber)

	return
}

// CreateGitCheckRun creates a PR Check
func CreateGitCheckRun(owner string, repo string, commitSHA string) (checkRunID string) {
	fmt.Println(GithubToken)
	path := fmt.Sprintf("/repos/%v/%v/check-runs", owner, repo)
	body := `{
		"name": "Brakeman Scan (Security)",
		"head_sha": "` + commitSHA + `",
		"status": "in_progress",
		"external_id": "03",
		"started_at": "` + time.Now().Format(time.RFC3339) + `",
		"output": {
			"title": "Brakeman Scan",
			"summary": "Scanning the commits in this pull request for potential security vulnerabilities",
			"text": ""
		}
	}`

	var jsonStr = []byte(body)
	url := fmt.Sprintf("%s%s", githubAPIHost, path)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "token "+GithubToken)
	req.Header.Set("Content-Type", "application/vnd.github.antiope-preview+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("[CreateGitCheckRun] response Status: ", resp.Status, "repo: ", repo, "owner: ", owner, "commit SHA: ", commitSHA)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyMap := make(map[string]json.RawMessage)
	e := json.Unmarshal(bodyBytes, &bodyMap)
	if e != nil {
		panic(err)
	}
	checkRunID = string(bodyMap["id"])

	return checkRunID
}

// CompleteGitCheckRun completes the PR check
func CompleteGitCheckRun(owner string, repo string, commitSHA string, checkRunID string, scanOutputString string, conclusion string) (err error) {
	path := fmt.Sprintf("/repos/%v/%v/check-runs/%v", owner, repo, checkRunID)
	body := `{
		"name": "Brakeman Scan (Security)",
		"status": "completed",
		"conclusion": "` + conclusion + `",
		"completed_at": "` + time.Now().Format(time.RFC3339) + `",
		"output": {
		  "title": "Brakeman Scan Report for the Pull Request",
		  "summary": "",
		  "text": "` + scanOutputString + `",
		  "annotations": [
		
		  ],
		  "images": [
		  ]
		},
		"actions": [
		
		]
	  }`
	var jsonStr = []byte(body)
	url := fmt.Sprintf("%s%s", githubAPIHost, path)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "token "+GithubToken)
	req.Header.Set("Content-Type", "application/vnd.github.antiope-preview+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	fmt.Println("[CompleteGitCheckRun] response Status: ", resp.Status, "repo: ", repo, "owner: ", owner, "commit SHA: ", commitSHA)

	return

}

func CloneGitRepository(repoURL string, dir string) (err error) {
	fmt.Println("[CloneGitRepository] repo= ", repoURL)
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
		Auth: &gitHttp.BasicAuth{
			Username: "abc123", // anything except an empty string (yes, it can be any string :D)
			Password: GithubToken,
		},
	})

	if err != nil {
		panic(err)
	}
	fmt.Println("[CloneGitRepository] Successful")
	return

}
