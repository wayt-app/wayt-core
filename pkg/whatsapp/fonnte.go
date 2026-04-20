package whatsapp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Sender sends a WhatsApp message to a phone number.
type Sender interface {
	Send(phone, message string) error
}

// fonnteSender sends messages via Fonnte API (https://fonnte.com).
type fonnteSender struct {
	token      string
	httpClient *http.Client
}

// NewFonnteSender creates a Sender backed by Fonnte.
// token is the device token from Fonnte dashboard.
func NewFonnteSender(token string) Sender {
	return &fonnteSender{token: token, httpClient: &http.Client{}}
}

func (f *fonnteSender) Send(phone, message string) error {
	// Normalize phone: strip leading +, keep digits
	phone = strings.TrimPrefix(phone, "+")
	// Convert 08xxx → 628xxx
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	}

	form := url.Values{}
	form.Set("target", phone)
	form.Set("message", message)

	req, err := http.NewRequest("POST", "https://api.fonnte.com/send", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", f.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fonnte error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// noopSender discards all messages silently.
type noopSender struct{}

func NewNoopSender() Sender { return &noopSender{} }

func (n *noopSender) Send(_, _ string) error { return nil }
