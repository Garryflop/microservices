package provider

import "context"

type EmailSender interface {
	SendNotification(ctx context.Context, to, subject, body string) error
}
