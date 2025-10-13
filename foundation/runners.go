package foundation

import (
	"github.com/goravel/framework/contracts/grpc"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
)

type RouteRunner struct {
	route route.Route
}

func NewRouteRunner(route route.Route) *RouteRunner {
	return &RouteRunner{
		route: route,
	}
}

func (r *RouteRunner) Name() string {
	return "Route"
}

func (r *RouteRunner) Run() error {
	return r.route.Run()
}

func (r *RouteRunner) Shutdown() error {
	return r.route.Shutdown()
}

type GrpcRunner struct {
	grpc grpc.Grpc
}

func NewGrpcRunner(grpc grpc.Grpc) *GrpcRunner {
	return &GrpcRunner{
		grpc: grpc,
	}
}

func (r *GrpcRunner) Name() string {
	return "Grpc"
}

func (r *GrpcRunner) Run() error {
	return r.grpc.Run()
}

func (r *GrpcRunner) Shutdown() error {
	return r.grpc.Shutdown()
}

type QueueRunner struct {
	worker queue.Worker
}

func NewQueueRunner(queue queue.Queue) *QueueRunner {
	return &QueueRunner{
		worker: queue.Worker(),
	}
}

func (r *QueueRunner) Name() string {
	return "Queue"
}

func (r *QueueRunner) Run() error {
	return r.worker.Run()
}

func (r *QueueRunner) Shutdown() error {
	return r.worker.Shutdown()
}

type ScheduleRunner struct {
	schedule schedule.Schedule
}

func NewScheduleRunner(schedule schedule.Schedule) *ScheduleRunner {
	return &ScheduleRunner{
		schedule: schedule,
	}
}

func (r *ScheduleRunner) Name() string {
	return "Schedule"
}

func (r *ScheduleRunner) Run() error {
	r.schedule.Run()

	return nil
}

func (r *ScheduleRunner) Shutdown() error {
	return r.schedule.Shutdown()
}
