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
	if len(args) != 10 {
		return fmt.Errorf("expected 10 arguments, got %d arguments", len(args))
	}

	subject, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("SUBJECT should be string")
	}

	html, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("HTML body should be string")
	}

	text, ok := args[2].(string)
	if !ok {
		return fmt.Errorf("TEXT body should be string")
	}

	fromAddress, ok := args[3].(string)
	if !ok {
		return fmt.Errorf("FromAddress should be string")
	}

	fromName, ok := args[4].(string)
	if !ok {
		return fmt.Errorf("FromName should be string")
	}

	to, ok := args[5].([]string)
	if !ok {
		return fmt.Errorf("TO should be []string")
	}

	cc, ok := args[6].([]string)
	if !ok {
		return fmt.Errorf("CC should be []string")
	}

	bcc, ok := args[7].([]string)
	if !ok {
		return fmt.Errorf("BCC should be []string")
	}

	attachments, ok := args[8].([]string)
	if !ok {
		return fmt.Errorf("ATTACHMENTS should be []string")
	}

	headerSlice, ok := args[9].([]string)
	if !ok {
		return fmt.Errorf("HEADERS should be []string")
	}

	params := Params{
		Subject:     subject,
		HTML:        html,
		Text:        text,
		FromAddress: fromAddress,
		FromName:    fromName,
		To:          to,
		CC:          cc,
		BCC:         bcc,
		Attachments: attachments,
		Headers:     convertSliceHeadersToMap(headerSlice),
	}

	return SendMail(r.config, params)
}
