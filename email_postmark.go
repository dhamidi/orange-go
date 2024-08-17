package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type PostmarkEmailSender struct {
	Logger *log.Logger
	APIKey string
	Client *http.Client
}

func NewPostmarkEmailSender(logger *log.Logger, apiKey string) *PostmarkEmailSender {
	return &PostmarkEmailSender{
		Logger: logger,
		APIKey: apiKey,
		Client: http.DefaultClient,
	}
}

type PostmarkSendWithTemplate struct {
	TemplateAlias string
	TemplateModel map[string]any
	From          string
	To            string
	Metadata      map[string]any
	MessageStream string
}

type PostmarkSendWithTemplateResponse struct {
	To          string
	SubmittedAt time.Time
	MessageID   string
	ErrorCode   int
	Message     string
}

func (res *PostmarkSendWithTemplateResponse) IsSuccess() bool {
	return res.ErrorCode == 0
}

func (res *PostmarkSendWithTemplateResponse) Error() string {
	return res.Message
}

func (res *PostmarkSendWithTemplateResponse) ExternalMessageID() string {
	return res.MessageID
}

func (p *PostmarkEmailSender) toRequest(email *Email) (*http.Request, error) {
	body := bytes.NewBufferString("")
	message := &PostmarkSendWithTemplate{
		TemplateAlias: email.TemplateName,
		TemplateModel: email.TemplateData,
		From:          "Dario <dario@decode.ee>",
		To:            email.Recipient,
		Metadata:      map[string]any{"internal_id": email.InternalID},
		MessageStream: "outbound",
	}
	if err := json.NewEncoder(body).Encode(message); err != nil {
		return nil, fmt.Errorf("Failed to encode email: %w", err)
	}
	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email/withTemplate", body)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct request for email: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", p.APIKey)
	return req, nil
}

// SendEmail implements EmailSender.
func (p *PostmarkEmailSender) SendEmail(email *Email) (EmailReceipt, error) {
	request, err := p.toRequest(email)
	if err != nil {
		return nil, err
	}
	response, err := p.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request to Postmark: %w", err)
	}
	responseBody := new(bytes.Buffer)
	if _, err := io.Copy(responseBody, response.Body); err != nil {
		return nil, fmt.Errorf("Failed to read response body from Postmark: %w", err)
	}

	message := &PostmarkSendWithTemplateResponse{}
	if err := json.Unmarshal(responseBody.Bytes(), message); err != nil {
		return nil, fmt.Errorf("Failed to decode response from Postmark: %w (response: %s)", err, responseBody.String())
	}

	if message.IsSuccess() {
		return message, nil
	}

	return nil, message
}

var _ EmailSender = &PostmarkEmailSender{}
