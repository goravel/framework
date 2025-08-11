package mail

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

type SendMailJobTestSuite struct {
	suite.Suite
	job        *SendMailJob
	mockConfig *mocksconfig.Config
}

func TestSendMailJobTestSuite(t *testing.T) {
	suite.Run(t, new(SendMailJobTestSuite))
}

func (r *SendMailJobTestSuite) SetupTest() {
	r.mockConfig = mocksconfig.NewConfig(r.T())
	r.job = NewSendMailJob(r.mockConfig)
	r.NotNil(r.job)
	r.Equal(r.mockConfig, r.job.config)
}

func (r *SendMailJobTestSuite) TestSignature() {
	r.Equal("goravel_send_mail_job", r.job.Signature())
}

func (r *SendMailJobTestSuite) TestHandle_WrongArgumentCount() {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "too few arguments",
			args: []any{"subject", "html"},
		},
		{
			name: "too many arguments",
			args: []any{
				"subject", "html", "text", "from", "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"}, "extra",
			},
		},
		{
			name: "no arguments",
			args: []any{},
		},
	}

	for _, test := range tests {
		r.Run(test.name, func() {
			err := r.job.Handle(test.args...)
			r.Contains(err.Error(), "expected 10 arguments")
		})
	}
}

func (r *SendMailJobTestSuite) TestHandle_WrongArgumentTypes() {
	tests := []struct {
		name     string
		args     []any
		errorMsg string
	}{
		{
			name: "subject not string",
			args: []any{
				123, "html", "text", "from", "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be string",
		},
		{
			name: "html not string",
			args: []any{
				"subject", 123, "text", "from", "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be string",
		},
		{
			name: "text not string",
			args: []any{
				"subject", "html", 123, "from", "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be string",
		},
		{
			name: "fromAddress not string",
			args: []any{
				"subject", "html", "text", 123, "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be string",
		},
		{
			name: "fromName not string",
			args: []any{
				"subject", "html", "text", "from", 123,
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be string",
		},
		{
			name: "to not []string",
			args: []any{
				"subject", "html", "text", "from", "name",
				"not-a-slice", []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be []string",
		},
		{
			name: "cc not []string",
			args: []any{
				"subject", "html", "text", "from", "name",
				[]string{"to"}, "not-a-slice", []string{"bcc"},
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be []string",
		},
		{
			name: "bcc not []string",
			args: []any{
				"subject", "html", "text", "from", "name",
				[]string{"to"}, []string{"cc"}, "not-a-slice",
				[]string{"attachments"}, []string{"headers"},
			},
			errorMsg: "should be []string",
		},
		{
			name: "attachments not []string",
			args: []any{
				"subject", "html", "text", "from", "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				"not-a-slice", []string{"headers"},
			},
			errorMsg: "should be []string",
		},
		{
			name: "headers not []string",
			args: []any{
				"subject", "html", "text", "from", "name",
				[]string{"to"}, []string{"cc"}, []string{"bcc"},
				[]string{"attachments"}, "not-a-slice",
			},
			errorMsg: "should be []string",
		},
	}

	for _, test := range tests {
		r.Run(test.name, func() {
			err := r.job.Handle(test.args...)
			r.Contains(err.Error(), test.errorMsg)
		})
	}
}
