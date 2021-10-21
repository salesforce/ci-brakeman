package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/ci-brakeman/github"
	"github.com/ci-brakeman/logger"
)

func DownloadRaw(tmpFolder, filename, url string) error {
	return downloadRaw(tmpFolder, filename, url)
}

// downloadRaw downloads the raw content of file from GitHub
// this uses the URL retrieved using the GetContents API call
// meaning it still goes through https://api.github.com/ and the Auth
// token is thus still valid
func downloadRaw(tmpFolder, filename, url string) error {
	fmt.Println(tmpFolder)
	logger.CreateBreadcrumb("downloadRaw", fmt.Sprintf("filename=%s,url=%s", filename, url))

	// if in sub-dir, recreate sub-dir structure
	tDir := path.Dir(filename)
	if tDir != "." && tDir != "/" {
		tpFolder := filepath.Join(tmpFolder, tDir)
		e := os.MkdirAll(tpFolder, 0755)
		if e != nil {
			logger.Error(e)
			return e
		}
	}
	fp := filepath.Join(tmpFolder, filepath.Clean(filename))
	tmpfile, err := os.Create(fp)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer tmpfile.Close()

	resp, err := http.Get(url)
	if err != nil {
		defer func() {
			e := os.Remove(tmpfile.Name())
			if e != nil {
				logger.Error(e)
			}
		}()
		logger.Error(err)
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		defer func() {
			e := os.Remove(tmpfile.Name())
			if e != nil {
				logger.Error(e)
			}
		}()
		logger.Error(err)
		return err
	}

	return nil
}

// downloadRawLarge functions the same as downloadRaw in that it downloads
// a file from GitHub via the API. The GetContents API is limited to files
// smaller than 1Mb, meaning we need to use additional API calls to get the
// raw file via the API (need to use the API as the Auth token is scoped to the API)
// this uses the GitHub Tree API to retrieve the URL to the raw blob
func downloadRawLarge(tmpFolder, owner, repo, filename, sha string) error {
	fmt.Println(tmpFolder)
	logger.CreateBreadcrumb("downloadRawLarge", fmt.Sprintf("filename=%s", filename))

	// if in sub-dir, recreate sub-dir structure
	tDir := path.Dir(filename)
	if tDir != "." && tDir != "/" {
		tpFolder := filepath.Join(tmpFolder, tDir)
		e := os.MkdirAll(tpFolder, 0755)
		if e != nil {
			logger.Error(e)
			return e
		}
	}
	// cleanup filename, since the filename might be somedir/somedir/somedir/file.txt
	fp := filepath.Join(tmpFolder, filepath.Clean(filename))

	tmpfile, err := os.Create(fp)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer tmpfile.Close()

	blob, resp, err := github.GetFileFromTree(owner, repo, filename, sha)

	if err != nil {
		if resp != nil {
			logger.CreateBreadcrumb("downloadRawLargeFail", resp.Message)
		}
		defer func() {
			e := os.Remove(tmpfile.Name())
			if e != nil {
				logger.Error(e)
			}
		}()
		logger.Error(err)
		return err
	}

	// data is ALWAYS base64 encoded, we will always need to decode it
	data, e := base64.StdEncoding.DecodeString(blob.Content)
	if e != nil {
		logger.Error(e)
	}

	if _, e := tmpfile.Write(data); e != nil {
		defer func() {
			e := os.Remove(tmpfile.Name())
			if e != nil {
				logger.Error(e)
			}
		}()
		logger.Error(e)
		return e
	}

	return nil
}
