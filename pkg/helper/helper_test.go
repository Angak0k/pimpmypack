package helper

import (
	"log"
	"os"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/config"
)

func TestMain(m *testing.M) {
	// init env
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file or environement variable : %v", err)
	}

	ret := m.Run()
	os.Exit(ret)

}

// should implement a smtp mock server to test this function
// func TestSendEmail(t *testing.T) {
//	mailRcpt := "pimpmypack@alki.earth"
//	mailSubject := "PimpMyPack - Test"
//	mailBody := "This is a test email from PimpMyPack."
//
//	t.Run("Test sending email", func(t *testing.T) {
//		fmt.Println("Running TestSendEmail")
//		err := SendEmail(mailRcpt, mailSubject, mailBody)
//		if err != nil {
//			t.Error(err)
//		}
//	})
//}
