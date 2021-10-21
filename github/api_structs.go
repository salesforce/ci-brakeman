/*
 * Copyright (c) 2021, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */

package github

import "time"

// TokenResponse represents a API token returned by the Github API
type TokenResponse struct {
	Token     string `json:"token,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// IssueCreate represents the struct to POST a new issue
type IssueCreate struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
	Labels    []string `json:"labels"`
}

// SignatureVerification represents GPG signature verification.
type SignatureVerification struct {
	Verified  *bool   `json:"verified,omitempty"`
	Reason    *string `json:"reason,omitempty"`
	Signature *string `json:"signature,omitempty"`
	Payload   *string `json:"payload,omitempty"`
}

// CommitAuthor represents the author or committer of a commit. The commit
// author may not correspond to a GitHub User.
type CommitAuthor struct {
	Date     *time.Time `json:"date,omitempty"`
	Name     *string    `json:"name,omitempty"`
	Email    *string    `json:"email,omitempty"`
	Username *string    `json:"username,omitempty"`
}

// CommitStats represents the stats on a commit
type CommitStats struct {
	Additions *int `json:"additions,omitempty"`
	Deletions *int `json:"deletions,omitempty"`
	Total     *int `json:"total,omitempty"`
}

// Commit represents a commit
type Commit struct {
	SHA          *string                `json:"sha,omitempty"`
	Author       *CommitAuthor          `json:"author,omitempty"`
	Committer    *CommitAuthor          `json:"committer,omitempty"`
	Message      *string                `json:"message,omitempty"`
	Tree         *Tree                  `json:"tree,omitempty"`
	Parents      []Commit               `json:"parents,omitempty"`
	Stats        *CommitStats           `json:"stats,omitempty"`
	HTMLURL      *string                `json:"html_url,omitempty"`
	URL          *string                `json:"url,omitempty"`
	Verification *SignatureVerification `json:"verification,omitempty"`
	NodeID       *string                `json:"node_id,omitempty"`
	CommentCount *int                   `json:"comment_count,omitempty"`
}

// RepoCommit represents a single commit for a repo
// used in API for owner/repo/commits/sha
type RepoCommit struct {
	URL         *string       `json:"url,omitempty"`
	SHA         *string       `json:"sha,omitempty"`
	NodeID      *string       `json:"node_id,omitempty"`
	HTMLURL     *string       `json:"html_url,omitempty"`
	CommentsURL *string       `json:"comments_url,omitempty"`
	Commit      Commit        `json:"commit,omitempty"`
	Author      *CommitAuthor `json:"author,omitempty"`
	Committer   *CommitAuthor `json:"committer,omitempty"`
	Parents     []Commit      `json:"parents,omitempty"`
	Stats       *CommitStats  `json:"stats,omitempty"`
	Files       []Files       `json:"files,omitempty"`
}

// Files  represents the files of a commit
type Files struct {
	Filename  *string `json:"filename,omitempty"`
	Additions int     `json:"additions,omitempty"`
	Deletions int     `json:"deletions,omitempty"`
	Changes   int     `json:"changes,omitempty"`
	Status    *string `json:"status,omitempty"`
	RawURL    *string `json:"raw_url,omitempty"`
	BlobURL   *string `json:"blob_url,omitempty"`
	Patch     *string `json:"patch,omitempty"`
}

// Content represents content in a git repo
type Content struct {
	Type        *string `json:"type,omitempty"`
	Encoding    *string `json:"encoding,omitempty"`
	Size        int     `json:"size,omitempty"`
	Name        *string `json:"name,omitempty"`
	Path        *string `json:"path,omitempty"`
	Content     *string `json:"content,omitempty"`
	SHA         *string `json:"sha,omitempty"`
	URL         *string `json:"url,omitempty"`
	GitURL      *string `json:"git_url,omitempty"`
	HTMLURL     *string `json:"html_url,omitempty"`
	DownloadURL *string `json:"download_url,omitempty"`
	Links       *Link   `json:"_links,omitempty"`
}

// Link represents the locations git content is located
type Link struct {
	Git  *string `json:"git,omitempty"`
	Self *string `json:"self,omitempty"`
	HTML *string `json:"html,omitempty"`
}

// Tree represents a GitHub tree.
type Tree struct {
	SHA       *string      `json:"sha,omitempty"`
	URL       *string      `json:"url,omitempty"`
	Entries   []*TreeEntry `json:"tree,omitempty"`
	Truncated *bool        `json:"truncated,omitempty"`
}

// TreeEntry represents a tree entry in git
type TreeEntry struct {
	SHA     *string `json:"sha,omitempty"`
	Path    *string `json:"path,omitempty"`
	Mode    *string `json:"mode,omitempty"`
	Type    *string `json:"type,omitempty"`
	Size    *int    `json:"size,omitempty"`
	Content *string `json:"content,omitempty"`
	URL     *string `json:"url,omitempty"`
}

// Response represents the error response the API may give
type Response struct {
	Message       string     `json:"message,omitempty"`
	Documentation string     `json:"documentation_url,omitempty"`
	Errors        []APIError `json:"errors,omitempty"`
}

// APIError represents more advanced errors returned when a 422 occurs
type APIError struct {
	Resource string `json:"resource,omitempty"`
	Field    string `json:"field,omitempty"`
	Code     string ` json:"code,omitempty"`
}

type Blob struct {
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	URL      string ` json:"url,omitempty"`
	SHA      string ` json:"sha,omitempty"`
	Size     int    ` json:"size,omitempty"`
}

// PullRequestFile struct
type PullRequestFile struct {
	Filename    *string `json:"filename,omitempty"`
	Additions   int     `json:"additions,omitempty"`
	Deletions   int     `json:"deletions,omitempty"`
	Changes     int     `json:"changes,omitempty"`
	Status      string  `json:"status,omitempty"`
	RawURL      string  `json:"raw_url,omitempty"`
	BlobURL     string  `json:"blob_url,omitempty"`
	ContentsURL string  `json:"contents_url,omitempty"`
	Patch       *string `json:"patch,omitempty"`
	SHA         *string `json:"sha,omitempty"`
}

// PullRequestFileResponses
type PullRequestFileResponse struct {
	Files []PullRequestFile
}
