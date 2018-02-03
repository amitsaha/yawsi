package cmd

import (
	"log"
	"os"

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
