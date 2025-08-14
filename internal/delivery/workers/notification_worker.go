// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package workers

import (
	"context"
	"fmt"
	"net/smtp"

	"streaming-platform/pkg/logger"
)

type NotificationWorker struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	logger       logger.Logger
}

func NewNotificationWorker(smtpHost, smtpPort, smtpUsername, smtpPassword string) *NotificationWorker {
	return &NotificationWorker{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		logger:       logger.NewLogger(),
	}
}

func (nw *NotificationWorker) ProcessJob(ctx context.Context, job map[string]interface{}) error {
	notificationType, ok := job["notification_type"].(string)
	if !ok {
		return fmt.Errorf("invalid notification_type")
	}

	switch notificationType {
	case "email":
		return nw.sendEmail(job)
	case "push":
		return nw.sendPushNotification(job)
	default:
		return fmt.Errorf("unknown notification type: %s", notificationType)
	}
}

func (nw *NotificationWorker) sendEmail(job map[string]interface{}) error {
	to, ok := job["to"].(string)
	if !ok {
		return fmt.Errorf("invalid email recipient")
	}

	subject, ok := job["subject"].(string)
	if !ok {
		return fmt.Errorf("invalid email subject")
	}

	body, ok := job["body"].(string)
	if !ok {
		return fmt.Errorf("invalid email body")
	}

	// Configurar autenticación SMTP
	auth := smtp.PlainAuth("", nw.smtpUsername, nw.smtpPassword, nw.smtpHost)

	// Construir mensaje
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// Enviar email
	err := smtp.SendMail(nw.smtpHost+":"+nw.smtpPort, auth, nw.smtpUsername, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	nw.logger.Info("Email sent successfully to: %s", to)
	return nil
}

func (nw *NotificationWorker) sendPushNotification(job map[string]interface{}) error {
	// Implementar envío de push notifications
	// Esto dependería del servicio que uses (Firebase, etc.)
	nw.logger.Info("Push notification would be sent here")
	return nil
}
