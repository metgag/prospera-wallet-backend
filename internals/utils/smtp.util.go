package utils

import (
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

func SendResetPasswordEmail(to string, resetURL string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "Prospera <titus.rangga.wicaksono@gmail.com>")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Reset Password")
	m.SetBody("text/html", "<p>Click link berikut untuk reset password:</p><a href='"+resetURL+"'>Reset Password</a>")

	d := gomail.NewDialer("smtp.gmail.com", 587, "titus.rangga.wicaksono@gmail.com", "your-app-password")

	return d.DialAndSend(m)
}

func GenerateRandomToken() string {
	return uuid.New().String()
}
