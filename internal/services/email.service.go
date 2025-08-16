package services

import (
	"fmt"
	"net/smtp"
	"os"
	"SalaryAdvance/internal/domain"
)

type EmailServiceImpl struct{}

func NewEmailService() domain.EmailService {
	return &EmailServiceImpl{}
}

func (s *EmailServiceImpl) SendInvite(toEmail, link string) error {
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	message := []byte(fmt.Sprintf("Subject: Invitation to Join\n\nClick the link to register: %s", link))

	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
}