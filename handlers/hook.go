/*
 * Copyright (c) 2021, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */

// Package handlers - hook
// Contains the logic to deal with incoming git hooks
package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/ci-brakeman/github"
	"github.com/ci-brakeman/logger"
	"github.com/ci-brakeman/scanner"
	"github.com/tidwall/gjson"
)

//DIFF_MINUTES how many minutes between re-reporting the same issue
const DIFF_MINUTES = 60 // 15 minutes * 60

// defining a Header type to access response headers
type Header map[string][]string

// Catcher function handles the incoming git hook and passes off
// handling of notifications, logging etc to the relevant functions
func Catcher(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error!"))
		logger.Error(err)
		return
	}

	xsig := []byte(r.Header.Get("X-Hub-Signature"))

	if checkSignature(body, xsig) == false {
		w.WriteHeader(401)
		w.Write([]byte("Signature mismatch!"))
		logger.Message("Unauthorized, signature mismatch")
		return
	}
	respstatus := 404
	respbody := []byte("")
	// get event type
	event := r.Header.Get("X-GitHub-Event")
	// hand off to relevant handlers
	switch event {
	case "pull_request":
		// can't wait for the scan to finish since large scans will timeout
		// so send 200 response
		respstatus = 200
		respbody = []byte("received")

		// trigger handler for the event
		go pullReqEvent(body)

		break
	default:
		respstatus = 404
		respbody = []byte("unsupported event")
		fmt.Printf("Unsupported event %s\n", event)
	}

	w.WriteHeader(respstatus)
	if respbody != nil {
		_, err := w.Write(respbody)
		if err != nil {
			logger.Error(err)
		}
	}
}

// checkSignature verifies that the supplied message has been signed with our
// secret. Annoyingly can't do this in auth middle-ware as it would mean having to pass through the body
func checkSignature(body, xsignature []byte) bool {
	secret := []byte(os.Getenv("GITHUB_SECRET"))

	if len(secret) != 0 && len(xsignature) != 0 {
		// check if signature matches
		// https://developer.github.com/webhooks/securing/
		// The HMAC hex digest is generated using the sha1 hash function and the secret as the HMAC key
		return checkHMAC(body, xsignature, secret)
	}

	return false
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkHMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)

	// we need to prepend sha1= to the message signature, since the supplied header is sha1=AAAA
	encodedMAC := fmt.Sprintf("sha1=%s", hex.EncodeToString(expectedMAC))

	//fmt.Printf("%s,%s\n", messageMAC, encodedMAC)

	return hmac.Equal(messageMAC, []byte(encodedMAC))
}


// this function is used to process the pull request and download the files

func processPullReq(number, owner, repo string, headSHA string, repoURL string) {
	tmpFolder := "tmp"
	// cleaning up the tmpFolder to ensure no residues from previous scan
	cleanUp(tmpFolder)
	// Creating a Check Run for the Pull Request
	checkRunID := github.CreateGitCheckRun(owner, repo, headSHA)

	logger.CreateBreadcrumb("processPullReq", fmt.Sprintf("owner=%s, repo=%s", owner, repo))

	errClone := github.CloneGitRepository(repoURL, tmpFolder)
	if errClone != nil {
		fmt.Println("Error while cloning the repository")
		logger.Error(errClone)
	}
	// scan all the downloaded files
	err := scan(tmpFolder, number, owner, repo, checkRunID, headSHA, repoURL)
	if err != nil {
		logger.Error(err)
	}

	// delete the tmpFolder containing the code to be scanned
	defer func() {
		if err := os.RemoveAll(tmpFolder); err != nil {
			logger.Error(err)
		}
	}()
}

func scan(tmpFolder, number, owner, repo string, checkRunID string, headSHA string, repoURL string) (err error) {
	logger.CreateBreadcrumb("scan", fmt.Sprintf("owner=%s,repo=%s,pullReqNumber=%s", owner, repo, number))

	// scan
	finding, errorBit, err := scanner.ScanFolder(tmpFolder)
	var scanOutput string
	if err != nil {
		logger.Error(err)
		return err
	}

	if errorBit == 0{
		var warnings string = "CI-Brakeman Scan Result \\n"
		if len(finding.Warnings) == 0 {
			warnings = "No warnings"
			scanOutput = "Findings: \\n" + warnings
		} else {
			for _,w := range finding.Warnings {
				fPathURL := fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s#L%s", owner, repo, headSHA, w.File, strconv.Itoa(w.Line))
		
				warnings = warnings + "Warning Type: "+string(w.WarningType)+"\\n"+"File: "+string(fPathURL)+"\\n\\n"
			}
			
			scanOutput = "Findings: \\n" + warnings
			//Complete the Check Run in the pull request
			github.CompleteGitCheckRun(owner, repo, headSHA, checkRunID, scanOutput, "success")
		}

	}else {
		scanOutput = "ERROR: Some error occured while scanning the pull request. Please contact the administrator of the tool."
		github.CompleteGitCheckRun(owner, repo, headSHA, checkRunID, scanOutput, "failure")
	}
	

	// creating the PR comment body with the scanOutputString
	scanOutputJSON := `{"body": "` + string(scanOutput) + `"}`
	comment := string([]byte(scanOutputJSON))

	//post comment to Github Pull Request
	github.PostCommentToGit(owner, repo, number, comment)

	cleanUp(tmpFolder)
	return
}

// handler for pull request. If pushEvent function above is not needed, it can be deleted

func pullReqEvent(body []byte) (int, []byte) {
	//https://developer.github.com/webhooks/event-payloads/#pull_request

	number := gjson.GetBytes(body, "number").String()
	// get repository name and owner
	// could use full_name := heroku/reponame , but the api code expects owner and repo as strings
	// this should already be known, but using the webhook data to avoid
	// hardcoding these values
	repo := gjson.GetBytes(body, "pull_request.head.repo.name").String()
	owner := gjson.GetBytes(body, "pull_request.head.repo.owner.login").String()
	headSHA := gjson.GetBytes(body, "pull_request.head.sha").String()
	// who created the pull request
	puller := gjson.GetBytes(body, "pull_request.user.login").String()
	repoURL := gjson.GetBytes(body, "pull_request.head.repo.html_url").String()

	logger.CreateBreadcrumb("pullReqEvent", fmt.Sprintf("number=%s,repo=%s/%s,puller=%s", number, owner, repo, puller))


	processPullReq(number, owner, repo, headSHA, repoURL)
	return 200, nil
}

// function to clean up the files downloaded during the execution
func cleanUp(tmpFolder string) {
	fmt.Println("[cleanUp] Cleaning up the files")
	files, err := ioutil.ReadDir(tmpFolder)
	if err != nil {
		logger.Error(err)
	}

	for _, f := range files {
		err := os.RemoveAll(path.Join([]string{tmpFolder, f.Name()}...))
		if err != nil {
			logger.Error(err)
		} 
	}
	fmt.Println("Cleanup complete.")

}