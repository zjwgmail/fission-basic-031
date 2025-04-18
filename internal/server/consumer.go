package server

import (
	"fission-basic/contrib/task"
	"fission-basic/internal/conf"
	"fission-basic/internal/service"
)

func NewConsumerServer(d *conf.Data, t *service.TaskService) *task.Server {
	s := task.NewServer()

	for _, t := range t.ListTasks() {
		_ = s.AddTask(t)
	}

	return s
}
