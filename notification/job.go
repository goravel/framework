package notification

import (
	"fmt"
	"github.com/goravel/framework/contracts/config"
)

type SendNotificationJob struct {
	config config.Config
}

func NewSendNotificationJob(config config.Config) *SendNotificationJob {
	return &SendNotificationJob{
		config: config,
	}
}

// Signature The name and signature of the job.
func (r *SendNotificationJob) Signature() string {
	return "goravel_send_notification_job"
}

// Handle Execute the job.
func (r *SendNotificationJob) Handle(args ...any) error {
	fmt.Println(args[0], args[1])
	
	return nil
}
