package email

import (
	"encoding/json"
	"log"
	"os"

	mail "github.com/xhit/go-simple-mail/v2"
)

// SendPlainEmail sends a plain text email with the provided data
func SendPlainEmail(subject string, data interface{}) error {
	server := mail.NewSMTPClient()
	server.Host = "mail.privateemail.com"
	server.Port = 587
	server.Username = "brian@datasnack.ai"
	server.Password = os.Getenv("EMAIL_PASSWORD")
	server.Encryption = mail.EncryptionSTARTTLS

	smtpClient, err := server.Connect()
	if err != nil {
		log.Printf("SMTP connection error: %v", err)
		return err
	}

	// Create email
	email := mail.NewMSG()
	email.SetFrom("DataSnack <brian@datasnack.ai>")
	email.AddTo("analytics@datasnack.ai")
	email.SetSubject(subject)

	// Convert data to JSON for email body
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return err
	}

	email.SetBody(mail.TextPlain, string(content))

	// Send email
	err = email.Send(smtpClient)
	if err != nil {
		log.Printf("Email send error: %v", err)
		return err
	}

	log.Println("Email sent successfully")
	return nil
}

// SendContactEmail sends a contact form email
func SendContactEmail(contactData interface{}) error {
	return SendPlainEmail("Contact Us Form Submission", contactData)
}
