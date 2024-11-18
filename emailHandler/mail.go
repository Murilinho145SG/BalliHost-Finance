package emailHandler

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"prodata/logs"
)

type EmailHandler struct {
	SmtpServer   string
	SmtpPort     string
	SmtpAddress  string
	SmtpUser     string
	SmtpPassword string
}

func SendEmail(sender *SimpleSender, handler *EmailHandler) {
	logger := logs.NewSistemLogger()

	sender.From = handler.SmtpAddress
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         handler.SmtpServer,
	}
	server := fmt.Sprintf("%s:%s", handler.SmtpServer, handler.SmtpPort)
	dial, err := tls.Dial("tcp", server, tlsConfig)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	defer dial.Close()

	client, err := smtp.NewClient(dial, handler.SmtpServer)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	defer client.Quit()

	auth := smtp.PlainAuth("", handler.SmtpAddress, handler.SmtpPassword, handler.SmtpServer)

	if err := client.Auth(auth); err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	if err := client.Mail(handler.SmtpAddress); err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	if err := client.Rcpt(sender.To); err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	writer, err := client.Data()

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	_, err = writer.Write(sender.Message)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	err = writer.Close()

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	logger.LogAndSendSystemMessage("Email enviado")
}
