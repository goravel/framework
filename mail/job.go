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
	if len(args) != 8 {
		return fmt.Errorf("expected 8 arguments, got %d", len(args))
	}

	from, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("argument 0 should be of type string")
	}

	subject, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("argument 1 should be of type string")
	}

	body, ok := args[2].(string)
	if !ok {
		return fmt.Errorf("argument 2 should be of type string")
	}

	recipient, ok := args[3].(string)
	if !ok {
		return fmt.Errorf("argument 3 should be of type string")
	}

	cc, ok := args[4].([]string)
	if !ok {
		return fmt.Errorf("argument 4 should be of type []string")
	}

	bcc, ok := args[5].([]string)
	if !ok {
		return fmt.Errorf("argument 5 should be of type []string")
	}

	replyTo, ok := args[6].([]string)
	if !ok {
		return fmt.Errorf("argument 6 should be of type []string")
	}

	attachments, ok := args[7].([]string)
	if !ok {
		return fmt.Errorf("argument 7 should be of type []string")
	}

	return SendMail(r.config, from, subject, body, recipient, cc, bcc, replyTo, attachments)
}
