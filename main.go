package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ci-brakeman/github"
	"github.com/ci-brakeman/handlers"
	"github.com/ci-brakeman/logger"
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/joho/godotenv"
)

var githubToken, githubInstallationID string
var gitHubAppID, gitHubKeyData string

func main() {

	initEnviron()

	// get initial auth token
	setupAuth()

	// Create a tmp folder if it does not exist
	createTmpFolder()

	// auth JWT is only valid for 10 minutes, so get a new one every 7
	ticker := time.NewTicker(7 * time.Minute)
	go func() {
		for range ticker.C {
			// get a new Auth token
			setupAuth()

		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	http.Handle("/", handlers.AuthCheck(http.HandlerFunc(handlers.Catcher)))
	http.Handle("/hook", handlers.AuthCheck(http.HandlerFunc(handlers.Catcher)))

	http.ListenAndServe(":"+port, nil)

}

func initEnviron() (err error) {
	// for local dev, use a .env file
	godotenv.Load(".env")

	environ := os.Getenv("ENVIRON")
	logger.Setup(environ)

	githubInstallationID = os.Getenv("GITHUB_INSTALLID")
	gitHubAppID = os.Getenv("GITHUB_APPID")
	gitHubKeyData = os.Getenv("GITHUB_PRIVATE_KEY")

	return nil
}

func setupAuth() error {

	var err error
	if github.JWTToken, err = generateJWT(); err != nil {
		logger.Error(err)
		return err
	}

	// try at least 3 times to get a token
	for i := 0; i < 3; i++ {
		if err = github.GetAccessToken(githubInstallationID); err != nil {
			logger.Error(fmt.Errorf("Auth attempt: %d Err: %s", i, err))
		} else {
			break
		}
	}
	return err
}

func generateJWT() (jwtToken string, err error) {

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"iss": gitHubAppID,
		"exp": time.Now().Add(time.Minute * 10).Unix(),
	})

	key, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(gitHubKeyData))
	jwtToken, err = token.SignedString(key)
	if err != nil {
		return
	}
	return
}

// if tmp folder does not exist, create it. This is mostly used when the application is first started.
func createTmpFolder() {
	_, err := os.Stat("tmp")

	if os.IsNotExist(err) {
		errDir := os.MkdirAll("tmp", 0755)
		if errDir != nil {
			logger.Error(err)
		} else {
			fmt.Println("[createTmpFolder] tmp folder created")
		}
	} else {
		fmt.Println("[createTmpFolder] Skipping tmp folder creation. Folder already exists.")
	}
	return
}
