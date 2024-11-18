package emailHandler

import (
	"io"
	"os"
	"path/filepath"
	"prodata/database/account"
	"prodata/logs"
	"strings"
)

type SimpleSender struct {
	From    string
	To      string
	Subject string
	Message []byte
}

func BuilderHTML(s *SimpleSender, html string) []byte {
	return []byte("From: " + s.From + "\r\n" +
		"To: " + s.To + "\r\n" +
		"Subject: " + s.Subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" + html)
}

func LoadHTMLFiles(name string) string {
	abs, err := filepath.Abs("./emailHandler/htmls/" + name + ".html")
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return ""
	}

	file, err := os.Open(abs)
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return ""
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)

	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return ""
	}

	return string(bytes)
}

func SetNoreply() *EmailHandler {
	return &EmailHandler{
		SmtpServer:   os.Getenv("HOST_MAIL"),
		SmtpPort:     os.Getenv("PORT_MAIL"),
		SmtpAddress:  os.Getenv("SEND_MAIL"),
		SmtpUser:     os.Getenv("SEND_MAIL"),
		SmtpPassword: os.Getenv("PASSWORD_MAIL"),
	}
}

func SendMagicLinkVerification(email, device string) {
	noreply := SetNoreply()

	magicEmail := LoadHTMLFiles("email_verification")

	magicLink := account.MagicLinkGenerator(device, email)

	parts := strings.Split(magicEmail, "%TOKEN%")

	sender := SimpleSender{
		From:    noreply.SmtpAddress,
		To:      email,
		Subject: "Verifique seu email",
	}

	sender.Message = BuilderHTML(&sender, parts[0]+magicLink+parts[1])

	SendEmail(&sender, noreply)
}
