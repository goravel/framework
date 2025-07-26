package mail

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
)

type SendMailJob struct {
	config config.Config
}

func NewSendMailJob(config config.Config) *SendMailJob {
	return &SendMailJob{
		config: config,
	}
}

// Signature The name and signature of the job.
func (r *SendMailJob) Signature() string {
	return "goravel_send_mail_job"
}

// Handle Execute the job.
func (r *SendMailJob) Handle(args ...any) error {
	if len(args) != 9 {
		return fmt.Errorf("expected 9 arguments, got %d", len(args))
	}

	from, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("FROM should be of type string")
	}

	subject, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("SUBJECT should be of type string")
	}

	html, ok := args[2].(string)
	if !ok {
		return fmt.Errorf("BODY should be of type string")
	}

	recipient, ok := args[3].(string)
	if !ok {
		return fmt.Errorf("RECIPIENT should be of type string")
	}

	cc, ok := args[4].([]string)
	if !ok {
		return fmt.Errorf("CC should be of type []string")
	}

	bcc, ok := args[5].([]string)
	if !ok {
		return fmt.Errorf("BCC should be of type []string")
	}

	replyTo, ok := args[6].([]string)
	if !ok {
		return fmt.Errorf("ReplyTo should be of type []string")
	}

	attachments, ok := args[7].([]string)
	if !ok {
		return fmt.Errorf("ATTACHMENTS should be of type []string")
	}

	headers, ok := args[8].([]string)
	if !ok {
		return fmt.Errorf("HEADERS should be of type []string")
	}

	text, ok := args[9].(string)
	if !ok {
		return fmt.Errorf("BODY should be of type string")
	}

	return SendMail(r.config, from, subject, text, html, recipient, cc, bcc, replyTo, attachments, convertSliceHeadersToMap(headers))
}
