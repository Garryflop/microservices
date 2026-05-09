package provider

import "context"

// EmailSender defines the interface for sending notifications.
// Implementations can be swapped via the PROVIDER_MODE env variable.
type EmailSender interface {
	SendNotification(ctx context.Context, to, subject, body string) error
}
