package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
)

type QueueFailedCommandTestSuite struct {
	suite.Suite
	mockFailer *mocksqueue.Failer
	mockQueue  *mocksqueue.Queue
	command    *QueueFailedCommand
}

func TestQueueFailedCommandTestSuite(t *testing.T) {
	suite.Run(t, new(QueueFailedCommandTestSuite))
}

func (s *QueueFailedCommandTestSuite) SetupTest() {
	s.mockFailer = mocksqueue.NewFailer(s.T())
	s.mockQueue = mocksqueue.NewQueue(s.T())

	s.command = NewQueueFailedCommand(s.mockQueue)
}

func (s *QueueFailedCommandTestSuite) TestHandle() {
	var mockCtx *mocksconsole.Context

	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "success",
			setup: func() {
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()

				mockFailedJob := mocksqueue.NewFailedJob(s.T())
				s.mockFailer.EXPECT().All().Return([]contractsqueue.FailedJob{
					mockFailedJob,
				}, nil).Once()

				mockCtx.EXPECT().Line("").Once()
				mockFailedJob.EXPECT().UUID().Return("test-uuid").Once()
				mockFailedJob.EXPECT().Connection().Return("test-connection").Once()
				mockFailedJob.EXPECT().Queue().Return("test-queue").Once()
				carbon.SetTestNow(carbon.Now())
				defer carbon.ClearTestNow()

				mockCtx.EXPECT().TwoColumnDetail("\x1b[90m"+carbon.Now().ToDateTimeString()+"\x1b[0m test-uuid", "test-connection@test-queue").Once()
			},
		},
		{
			name: "failed to get failed jobs",
			setup: func() {
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()
				s.mockFailer.EXPECT().All().Return(nil, assert.AnError).Once()
				mockCtx.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "no failed jobs found",
			setup: func() {
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()
				s.mockFailer.EXPECT().All().Return(nil, nil).Once()
				mockCtx.EXPECT().Info(errors.QueueNoFailedJobsFound.Error()).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			mockCtx = mocksconsole.NewContext(s.T())

			tt.setup()

			err := s.command.Handle(mockCtx)

			s.NoError(err)
		})
	}
}

func (s *QueueFailedCommandTestSuite) TestPrintSuccess() {
	mockCtx := mocksconsole.NewContext(s.T())
	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	mockCtx.EXPECT().TwoColumnDetail("\x1b[90m"+carbon.Now().ToDateTimeString()+"\x1b[0m test-uuid", "test-connection@test-queue").Once()

	s.command.printJob(mockCtx, "test-uuid", "test-connection", "test-queue")
}
