package service

import (
	"github.com/RichardKnop/machinery/v1/backends/iface"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
)

type MachineryServer interface {
	SendTask(signature *tasks.Signature) (*result.AsyncResult, error)
	GetBackend() iface.Backend
}

type MachineryBackend interface {
	GetState(taskUUID string) (*tasks.TaskState, error)
}
