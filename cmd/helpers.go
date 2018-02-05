package cmd

import (
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
	// TODO: support this better:
	// https://bbengfort.github.io/snippets/2018/01/06/cli-editor-app.html
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	tmpDir := os.TempDir()
	tmpFile, tmpFileErr := ioutil.TempFile(tmpDir, "yawsiTmp")
	if tmpFileErr != nil {
		return nil, tmpFileErr
	}

	err := ioutil.WriteFile(tmpFile.Name(), []byte(userData), 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	path, err := exec.LookPath(editor)
	if err != nil {
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
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	// Read the contents back
	editedFileContents, err := ioutil.ReadFile(tmpFile.Name())
	editedFileContentsStr := string(editedFileContents)
	return &editedFileContentsStr, err
}
