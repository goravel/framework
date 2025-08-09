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
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument of type mail.Params, got %d arguments", len(args))
	}

	params, ok := args[0].(Params)
	if !ok {
		return fmt.Errorf("argument should be of type mail.Params")
	}

	return SendMail(r.config, params)
}
