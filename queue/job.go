package queue

import (
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/samber/do/v2"
)

var injector = do.New()

// Register registers jobs to the registry.
// Register 将作业注册到注册表。
func Register(jobs []contractsqueue.Job) error {
	for _, job := range jobs {
		do.ProvideNamedValue(injector, job.Signature(), job)
	}

	return nil
}

// Call calls a registered job using its signature.
// Call 使用其签名调用已注册的作业。
func Call(signature string, args []contractsqueue.Arg) error {
	job, err := do.InvokeNamed[contractsqueue.Job](injector, signature)
	if err != nil {
		return err
	}

	return job.Handle(args)
}

// Get gets a registered job using its signature.
// Get 使用其签名获取已注册的作业。
func Get(signature string) (contractsqueue.Job, error) {
	return do.InvokeNamed[contractsqueue.Job](injector, signature)
}
