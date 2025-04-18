package queue

import (
	"time"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
)

type Task struct {
	Job
	Uuid  string `json:"uuid"`
	Chain []Job  `json:"chain"`
}

type Job struct {
	Signature string      `json:"signature"`
	Args      []queue.Arg `json:"args"`
	Delay     *time.Time  `json:"delay"`
}

func TaskToJson(task queue.Task, json foundation.Json) ([]byte, error) {
	chain := make([]Job, len(task.Chain))
	for i, taskData := range task.Chain {
		for j, arg := range taskData.Args {
			// To avoid converting []uint8 to base64
			if arg.Type == "[]uint8" {
				taskData.Args[j].Value = cast.ToIntSlice(arg.Value)
			}
		}

		job := Job{
			Signature: taskData.Job.Signature(),
			Args:      taskData.Args,
		}

		if !taskData.Delay.IsZero() {
			job.Delay = &taskData.Delay
		}

		chain[i] = job
	}

	var args []queue.Arg
	for _, arg := range task.Args {
		if arg.Type == "[]uint8" {
			arg.Value = cast.ToIntSlice(arg.Value)
		}
		args = append(args, arg)
	}

	job := Job{
		Signature: task.Job.Signature(),
		Args:      args,
	}

	if !task.Delay.IsZero() {
		job.Delay = &task.Delay
	}

	t := Task{
		Uuid:  task.UUID,
		Job:   job,
		Chain: chain,
	}

	payload, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func JsonToTask(payload string, jobRepository queue.JobRepository, json foundation.Json) (queue.Task, error) {
	var task Task
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return queue.Task{}, err
	}

	chain := make([]queue.Jobs, len(task.Chain))
	for i, item := range task.Chain {
		job, err := jobRepository.Get(item.Signature)
		if err != nil {
			return queue.Task{}, err
		}

		jobs := queue.Jobs{
			Job:  job,
			Args: item.Args,
		}

		if item.Delay != nil && !item.Delay.IsZero() {
			jobs.Delay = *item.Delay
		}

		chain[i] = jobs
	}

	job, err := jobRepository.Get(task.Job.Signature)
	if err != nil {
		return queue.Task{}, err
	}

	jobs := queue.Jobs{
		Job:  job,
		Args: task.Args,
	}

	if task.Delay != nil && !task.Delay.IsZero() {
		jobs.Delay = *task.Delay
	}

	return queue.Task{
		UUID:  task.Uuid,
		Jobs:  jobs,
		Chain: chain,
	}, nil
}

func filterArgsType(args []queue.Arg) []any {
	realArgs := make([]any, 0, len(args))
	for _, arg := range args {
		switch arg.Type {
		case "bool":
			realArgs = append(realArgs, cast.ToBool(arg.Value))
		case "int":
			realArgs = append(realArgs, cast.ToInt(arg.Value))
		case "int8":
			realArgs = append(realArgs, cast.ToInt8(arg.Value))
		case "int16":
			realArgs = append(realArgs, cast.ToInt16(arg.Value))
		case "int32":
			realArgs = append(realArgs, cast.ToInt32(arg.Value))
		case "int64":
			realArgs = append(realArgs, cast.ToInt64(arg.Value))
		case "uint":
			realArgs = append(realArgs, cast.ToUint(arg.Value))
		case "uint8":
			realArgs = append(realArgs, cast.ToUint8(arg.Value))
		case "uint16":
			realArgs = append(realArgs, cast.ToUint16(arg.Value))
		case "uint32":
			realArgs = append(realArgs, cast.ToUint32(arg.Value))
		case "uint64":
			realArgs = append(realArgs, cast.ToUint64(arg.Value))
		case "float32":
			realArgs = append(realArgs, cast.ToFloat32(arg.Value))
		case "float64":
			realArgs = append(realArgs, cast.ToFloat64(arg.Value))
		case "string":
			realArgs = append(realArgs, cast.ToString(arg.Value))
		case "[]bool":
			realArgs = append(realArgs, cast.ToBoolSlice(arg.Value))
		case "[]int":
			realArgs = append(realArgs, cast.ToIntSlice(arg.Value))
		case "[]int8":
			var int8Slice []int8
			for _, v := range cast.ToSlice(arg.Value) {
				int8Slice = append(int8Slice, cast.ToInt8(v))
			}
			realArgs = append(realArgs, int8Slice)
		case "[]int16":
			var int16Slice []int16
			for _, v := range cast.ToSlice(arg.Value) {
				int16Slice = append(int16Slice, cast.ToInt16(v))
			}
			realArgs = append(realArgs, int16Slice)
		case "[]int32":
			var int32Slice []int32
			for _, v := range cast.ToSlice(arg.Value) {
				int32Slice = append(int32Slice, cast.ToInt32(v))
			}
			realArgs = append(realArgs, int32Slice)
		case "[]int64":
			var int64Slice []int64
			for _, v := range cast.ToSlice(arg.Value) {
				int64Slice = append(int64Slice, cast.ToInt64(v))
			}
			realArgs = append(realArgs, int64Slice)
		case "[]uint":
			var uintSlice []uint
			for _, v := range cast.ToSlice(arg.Value) {
				uintSlice = append(uintSlice, cast.ToUint(v))
			}
			realArgs = append(realArgs, uintSlice)
		case "[]uint8":
			var uint8Slice []uint8
			for _, v := range cast.ToSlice(arg.Value) {
				uint8Slice = append(uint8Slice, cast.ToUint8(v))
			}
			realArgs = append(realArgs, uint8Slice)
		case "[]uint16":
			var uint16Slice []uint16
			for _, v := range cast.ToSlice(arg.Value) {
				uint16Slice = append(uint16Slice, cast.ToUint16(v))
			}
			realArgs = append(realArgs, uint16Slice)
		case "[]uint32":
			var uint32Slice []uint32
			for _, v := range cast.ToSlice(arg.Value) {
				uint32Slice = append(uint32Slice, cast.ToUint32(v))
			}
			realArgs = append(realArgs, uint32Slice)
		case "[]uint64":
			var uint64Slice []uint64
			for _, v := range cast.ToSlice(arg.Value) {
				uint64Slice = append(uint64Slice, cast.ToUint64(v))
			}
			realArgs = append(realArgs, uint64Slice)
		case "[]float32":
			var float32Slice []float32
			for _, v := range cast.ToSlice(arg.Value) {
				float32Slice = append(float32Slice, cast.ToFloat32(v))
			}
			realArgs = append(realArgs, float32Slice)
		case "[]float64":
			var float64Slice []float64
			for _, v := range cast.ToSlice(arg.Value) {
				float64Slice = append(float64Slice, cast.ToFloat64(v))
			}
			realArgs = append(realArgs, float64Slice)
		case "[]string":
			realArgs = append(realArgs, cast.ToStringSlice(arg.Value))
		}
	}
	return realArgs
}
