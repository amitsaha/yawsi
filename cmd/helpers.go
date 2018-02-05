package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws/session"
)

func createSession() *session.Session {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		log.Fatal("Must specify AWS_PROFILE")
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))
	return sess
}

func modifyUserData(userData string) (*string, error) {
	vi := "vim"
	tmpDir := os.TempDir()
	tmpFile, tmpFileErr := ioutil.TempFile(tmpDir, "tempFilePrefix")
	if tmpFileErr != nil {
		fmt.Printf("Error %s while creating tempFile", tmpFileErr)
	}

	err := ioutil.WriteFile(tmpFile.Name(), []byte(userData), 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	path, err := exec.LookPath(vi)
	if err != nil {
		fmt.Printf("Error %s while looking up for %s!!", path, vi)
		return nil, err
	}

	cmd := exec.Command(path, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Waiting for command to finish.\n")
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	// Read the contents back
	editedFileContents, err := ioutil.ReadFile(tmpFile.Name())
	editedFileContentsStr := string(editedFileContents)
	return &editedFileContentsStr, err
}
