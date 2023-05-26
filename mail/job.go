package mail

import (
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

//Signature The name and signature of the job.
func (r *SendMailJob) Signature() string {
	return "goravel_send_mail_job"
}

//Handle Execute the job.
func (r *SendMailJob) Handle(args ...any) error {
	return SendMail(r.config, args[0].(string), args[1].(string), args[2].(string), args[3].(string), args[4].([]string), args[5].([]string), args[6].([]string), args[7].([]string))
}
