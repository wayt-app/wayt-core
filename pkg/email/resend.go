package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Sender interface {
	Send(to, subject, htmlBody string) error
}

type resendSender struct {
	apiKey    string
	fromEmail string
}

func NewResendSender(apiKey, fromEmail string) Sender {
	return &resendSender{apiKey: apiKey, fromEmail: fromEmail}
}

type noopSender struct{}

func NewNoopSender() Sender { return &noopSender{} }

func (n *noopSender) Send(to, subject, htmlBody string) error { return nil }

func (r *resendSender) Send(to, subject, htmlBody string) error {
	if r.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY belum diset")
	}
	payload := map[string]any{
		"from":    r.fromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gagal mengirim email: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}
