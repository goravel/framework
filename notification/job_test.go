package notification

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

type SendNotificationJobTestSuite struct {
	suite.Suite
	job        *SendNotificationJob
	mockConfig *mocksconfig.Config
}

func TestSendNotificationJobTestSuite(t *testing.T) {
	suite.Run(t, new(SendNotificationJobTestSuite))
}

func (r *SendNotificationJobTestSuite) SetupTest() {
	r.mockConfig = mocksconfig.NewConfig(r.T())
	r.job = NewSendNotificationJob(r.mockConfig, nil, nil)
	r.NotNil(r.job)
	r.Equal(r.mockConfig, r.job.config)
}

func (r *SendNotificationJobTestSuite) TestSignature() {
	r.Equal("goravel_send_notification_job", r.job.Signature())
}

func (r *SendNotificationJobTestSuite) TestHandle_WrongArgumentCount() {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "too few arguments",
			args: []any{"subject", "html"},
		},
	}

	for _, test := range tests {
		r.Run(test.name, func() {
			err := r.job.Handle(test.args...)
			r.Contains(err.Error(), "expected 3 arguments")
		})
	}
}

func (r *SendNotificationJobTestSuite) TestHandle_WrongArgumentTypes() {

}
