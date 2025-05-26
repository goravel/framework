package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/convert"
)

func TestTaskToJson(t *testing.T) {
	mockJson := mocksfoundation.NewJson(t)
	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	tests := []struct {
		name          string
		task          contractsqueue.Task
		setup         func()
		expectedJson  string
		expectedError error
	}{
		{
			name: "successful conversion with simple task",
			task: contractsqueue.Task{
				UUID: "test-uuid",
				ChainJob: contractsqueue.ChainJob{
					Job: &TestJobOne{},
					Args: []contractsqueue.Arg{
						{Type: "string", Value: "test"},
					},
				},
			},
			setup: func() {
				expectedTask := Task{
					UUID: "test-uuid",
					Job: Job{
						Signature: "test_job_one",
						Args: []contractsqueue.Arg{
							{Type: "string", Value: "test"},
						},
					},
					Chain: []Job{},
				}
				mockJson.EXPECT().MarshalString(expectedTask).Return("{\"test\":true}", nil).Once()
			},
			expectedJson:  "{\"test\":true}",
			expectedError: nil,
		},
		{
			name: "successful conversion with task chain",
			task: contractsqueue.Task{
				UUID: "test-uuid",
				ChainJob: contractsqueue.ChainJob{
					Job: &TestJobOne{},
				},
				Chain: []contractsqueue.ChainJob{
					{
						Job: &TestJobTwo{},
						Args: []contractsqueue.Arg{
							{Type: "int", Value: 42},
						},
					},
				},
			},
			setup: func() {
				expectedTask := Task{
					UUID: "test-uuid",
					Job: Job{
						Signature: "test_job_one",
					},
					Chain: []Job{
						{
							Signature: "test_job_two",
							Args: []contractsqueue.Arg{
								{Type: "int", Value: 42},
							},
						},
					},
				}
				mockJson.EXPECT().MarshalString(expectedTask).Return("{\"test\":true}", nil).Once()
			},
			expectedJson:  "{\"test\":true}",
			expectedError: nil,
		},
		{
			name: "successful conversion with delay",
			task: contractsqueue.Task{
				UUID: "test-uuid",
				ChainJob: contractsqueue.ChainJob{
					Job:   &TestJobOne{},
					Delay: carbon.Now().StdTime(),
				},
			},
			setup: func() {
				expectedTask := Task{
					UUID: "test-uuid",
					Job: Job{
						Signature: "test_job_one",
						Delay:     convert.Pointer(carbon.Now().StdTime()),
					},
					Chain: []Job{},
				}
				mockJson.EXPECT().MarshalString(expectedTask).Return("{\"test\":true}", nil).Once()
			},
			expectedJson:  "{\"test\":true}",
			expectedError: nil,
		},
		{
			name: "marshal error",
			task: contractsqueue.Task{
				UUID: "test-uuid",
				ChainJob: contractsqueue.ChainJob{
					Job: &TestJobOne{},
				},
			},
			setup: func() {
				expectedTask := Task{
					UUID: "test-uuid",
					Job: Job{
						Signature: "test_job_one",
					},
					Chain: []Job{},
				}
				mockJson.EXPECT().MarshalString(expectedTask).Return("", assert.AnError).Once()
			},
			expectedJson:  "",
			expectedError: assert.AnError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			json, err := TaskToJson(test.task, mockJson)

			assert.Equal(t, test.expectedJson, json)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestJsonToTask(t *testing.T) {
	mockJson := mocksfoundation.NewJson(t)
	mockJobStorer := mocksqueue.NewJobStorer(t)

	tests := []struct {
		name          string
		payload       string
		setup         func()
		expectedTask  contractsqueue.Task
		expectedError error
	}{
		{
			name:    "successful conversion with simple task",
			payload: "{\"uuid\":\"test-uuid\",\"signature\":\"test_job_one\",\"args\":[]}",
			setup: func() {
				var task Task
				mockJson.EXPECT().Unmarshal([]byte("{\"uuid\":\"test-uuid\",\"signature\":\"test_job_one\",\"args\":[]}"), &task).
					Run(func(_ []byte, taskPtr any) {
						taskPtr.(*Task).UUID = "test-uuid"
						taskPtr.(*Task).Job.Signature = "test_job_one"
					}).Return(nil).Once()
				mockJobStorer.EXPECT().Get("test_job_one").Return(&TestJobOne{}, nil).Once()
			},
			expectedTask: contractsqueue.Task{
				UUID: "test-uuid",
				ChainJob: contractsqueue.ChainJob{
					Job: &TestJobOne{},
				},
				Chain: []contractsqueue.ChainJob{},
			},
			expectedError: nil,
		},
		{
			name:    "successful conversion with task chain",
			payload: "{\"uuid\":\"test-uuid\",\"signature\":\"test_job_one\",\"args\":[],\"chain\":[{\"signature\":\"test_job_two\",\"args\":[{\"type\":\"int\",\"value\":42}]}]}",
			setup: func() {
				var task Task
				mockJson.EXPECT().Unmarshal([]byte("{\"uuid\":\"test-uuid\",\"signature\":\"test_job_one\",\"args\":[],\"chain\":[{\"signature\":\"test_job_two\",\"args\":[{\"type\":\"int\",\"value\":42}]}]}"), &task).
					Run(func(_ []byte, taskPtr any) {
						taskPtr.(*Task).UUID = "test-uuid"
						taskPtr.(*Task).Job.Signature = "test_job_one"
						taskPtr.(*Task).Chain = []Job{
							{
								Signature: "test_job_two",
								Args: []contractsqueue.Arg{
									{Type: "int", Value: 42},
								},
							},
						}
					}).Return(nil).Once()
				mockJobStorer.EXPECT().Get("test_job_one").Return(&TestJobOne{}, nil).Once()
				mockJobStorer.EXPECT().Get("test_job_two").Return(&TestJobTwo{}, nil).Once()
			},
			expectedTask: contractsqueue.Task{
				UUID: "test-uuid",
				ChainJob: contractsqueue.ChainJob{
					Job: &TestJobOne{},
				},
				Chain: []contractsqueue.ChainJob{
					{
						Job: &TestJobTwo{},
						Args: []contractsqueue.Arg{
							{Type: "int", Value: 42},
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:    "unmarshal error",
			payload: "invalid json",
			setup: func() {
				var task Task
				mockJson.EXPECT().Unmarshal([]byte("invalid json"), &task).Return(assert.AnError).Once()
			},
			expectedTask:  contractsqueue.Task{},
			expectedError: assert.AnError,
		},
		{
			name:    "job storer error",
			payload: "{\"uuid\":\"test-uuid\",\"signature\":\"test_job_one\",\"args\":[]}",
			setup: func() {
				var task Task
				mockJson.EXPECT().Unmarshal([]byte("{\"uuid\":\"test-uuid\",\"signature\":\"test_job_one\",\"args\":[]}"), &task).
					Run(func(_ []byte, taskPtr any) {
						taskPtr.(*Task).UUID = "test-uuid"
						taskPtr.(*Task).Job.Signature = "test_job_one"
					}).Return(nil).Once()
				mockJobStorer.EXPECT().Get("test_job_one").Return(nil, assert.AnError).Once()
			},
			expectedTask:  contractsqueue.Task{},
			expectedError: assert.AnError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			task, err := JsonToTask(test.payload, mockJobStorer, mockJson)

			assert.Equal(t, test.expectedTask, task)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestConvertArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []contractsqueue.Arg
		expected []any
	}{
		{
			name: "convert basic types",
			args: []contractsqueue.Arg{
				{Type: "bool", Value: true},
				{Type: "int", Value: 42},
				{Type: "string", Value: "test"},
				{Type: "float64", Value: 3.14},
			},
			expected: []any{true, 42, "test", 3.14},
		},
		{
			name: "convert slice types",
			args: []contractsqueue.Arg{
				{Type: "[]bool", Value: []bool{true, false}},
				{Type: "[]int", Value: []int{1, 2, 3}},
				{Type: "[]string", Value: []string{"a", "b", "c"}},
				{Type: "[]float64", Value: []float64{1.1, 2.2, 3.3}},
			},
			expected: []any{
				[]bool{true, false},
				[]int{1, 2, 3},
				[]string{"a", "b", "c"},
				[]float64{1.1, 2.2, 3.3},
			},
		},
		{
			name: "convert uint types",
			args: []contractsqueue.Arg{
				{Type: "uint", Value: uint(42)},
				{Type: "uint8", Value: uint8(42)},
				{Type: "uint16", Value: uint16(42)},
				{Type: "uint32", Value: uint32(42)},
				{Type: "uint64", Value: uint64(42)},
			},
			expected: []any{
				uint(42),
				uint8(42),
				uint16(42),
				uint32(42),
				uint64(42),
			},
		},
		{
			name: "convert int types",
			args: []contractsqueue.Arg{
				{Type: "int8", Value: int8(42)},
				{Type: "int16", Value: int16(42)},
				{Type: "int32", Value: int32(42)},
				{Type: "int64", Value: int64(42)},
			},
			expected: []any{
				int8(42),
				int16(42),
				int32(42),
				int64(42),
			},
		},
		{
			name: "convert float types",
			args: []contractsqueue.Arg{
				{Type: "float32", Value: float32(3.14)},
				{Type: "float64", Value: 3.14},
			},
			expected: []any{
				float32(3.14),
				3.14,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ConvertArgs(test.args)
			assert.Equal(t, test.expected, result)
		})
	}
}
