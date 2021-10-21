package scanner

import (

	// "encoding/json"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	// "strings"
)

// ScanInfo is a construct to store scan_info from Findings
type ScanInfo struct {
	AppPath             string   `json:"app_path"`
	RailsVersion        string   `json:"rails_version"`
	SecurityWarnings    int      `json:"security_warnings"`
	StartTime           string   `json:"start_time"`
	EndTime             string   `json:"end_time"`
	Duration            float64  `json:"duration"`
	ChecksPerformed     []string `json:"checks_performed"`
	NumberOfControllers int      `json:"number_of_controllers"`
	NumberOfModels      int      `json:"number_of_models"`
	NumberOfTemplates   int      `json:"number_of_templates"`
	RubyVersion         string   `json:"ruby_version"`
	BrakemanVersion     string   `json:"brakeman_version"`
}

// Findings is a construct to store findings from Brakeman
type Findings struct {
	ScanInfo        ScanInfo `json:"scan_info,omitempty"`
	Warnings        []WarningInfo `json:"warnings,omitempty"`
	IgnoredWarnings []string `json:"ignored_warnings,omitempty"`
	Errors          []string `json:"erros,omitempty"`
	Obsolete        []string `json:"obsolete,omitempty"`
}

type WarningInfo struct {
	WarningType         string   		`json:"warning_type"`
	WarningCode         int 	 		`json:"warning_code"`
	FingerPrint    		string   		`json:"fingerprint"`
	CheckName           string   		`json:"check_name"`
	Message             string  	    `json:"message"`
	File            	string   	    `json:"file"`
	Line                int			    `json:"line"`
	Link				string   		`json:"link"`
	Code      			string   		`json:"code"`
	Location   			LocationInfo   	`json:"location"`
	UserInput         	string   		`json:"user_input"`
	Confidence     	    string   		`json:"confidence"`
}

type LocationInfo struct{
	Type         		string   	`json:"type"`
	Class         		string 	 	`json:"class"`
	Method    			string   	`json:"method"`

}

// tmp folder holds all the files that need to be scanned
var tmpFolder = "tmp"

// ScanFolder takes a path to a folder to scan, calls the grover binary to do the scan
// and returns a list of findings, and an error state.
func ScanFolder(tmpFolder string) (finding Findings, errorBit int, err error) {
	// this errorBit is set to 1 if there are errors in the execution of brakeman
	// in case of an error, the Github check is failed but the PR is not blocked
	errorBit = 0
	if _, err := os.Stat(tmpFolder + "/app"); !os.IsNotExist(err) {
		// since Brakeman scans only the app directory, we run the scan only if the app directory exists.

		cmd := exec.Command("./vendor/bundle/bin/brakeman", "-q", "--format", "json", "-p", tmpFolder, "--no-pager", "--no-exit-on-warn", "--no-exit-on-error")
		stdout, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		out := []byte(stdout)

		if err := json.Unmarshal(out, &finding); err != nil {
			fmt.Println(err)
		}
		// Uncomment the checks variable if you need 'Checks Performed' section in the output
		// append the checks variable to the warnings string
		// import the strings library 
		// checks := strings.Join(finding.ScanInfo.ChecksPerformed, ", ")

		

	} else {
		errorBit = 1
	}
	return finding, errorBit, nil
}