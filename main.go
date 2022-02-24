package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	tele "gopkg.in/telebot.v3"
)

func main() {
	// Load Config from Environment Variable
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	b := InitTelegramService()

	b.Handle(tele.OnText, func(c tele.Context) error {
		// All the text messages that weren't
		// captured by existing handlers.

		var (
			text = c.Text()
		)
		client, session := InitSSHClientSession()
		result := RunCommand(session, text)
		defer client.Close()
		defer session.Close()
		return c.Send(result)
	})
	b.Start()

}

func InitSSHClientSession() (*ssh.Client, *ssh.Session) {
	SSH_ADDRESS := os.Getenv("SSH_ADDRESS")
	SSH_USERNAME := os.Getenv("MIKROTIK_USER")
	SSH_PASSWORD := os.Getenv("MIKROTIK_PASSWORD")

	sshConfig := &ssh.ClientConfig{
		User:            SSH_USERNAME,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(SSH_PASSWORD),
		},
	}

	// Open connection to server
	client, err := ssh.Dial("tcp", SSH_ADDRESS, sshConfig)
	if client != nil {
		log.Println("sucessfull login")
	}
	if err != nil {
		log.Fatal("Failed to dial. " + err.Error())
	}

	// start ssh session
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session. " + err.Error())
	}
	return client, session
}

func RunCommand(session *ssh.Session, command string) (result string) {
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal("Error getting stdin pipe. " + err.Error())
	}

	// Run Command
	err = session.Start(command)
	if err != nil {
		log.Fatal("Error starting bash. " + err.Error())
	}
	commands := []string{
		"quit",
	}
	for _, cmd := range commands {
		if _, err = fmt.Fprintln(stdin, cmd); err != nil {
			log.Fatal(err)
		}
	}
	err = session.Wait()
	if err != nil {
		log.Fatal(err)
	}

	outputString := stdout.String()
	outputSErr := stderr.String()

	if outputSErr != "" {
		log.Fatal("Failed to dial. " + outputSErr)
	}
	return outputString
}

func InitTelegramService() *tele.Bot {
	pref := tele.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Run Telegram Service")
	return b
}
