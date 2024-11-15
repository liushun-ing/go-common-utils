package taskservice

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

var Servers = sync.Map{}

func NewServer(serverName string, handler MessageHandler, scheduledTasks []ScheduledTask, opts ...Option) (server *Server, err error) {
	if serverName == "" {
		serverName = RandomString(16)
	}
	_, ok := Servers.Load(serverName)
	if ok {
		return nil, fmt.Errorf("server %s already exists", serverName)
	}

	option := NewDefaultOptions()

	for _, opt := range opts {
		if err = opt(option); err != nil {
			return nil, err
		}
	}

	server = &Server{
		Name:                      serverName,
		Opts:                      option,
		Queue:                     make(chan *Message, option.capacity),
		MsgHandler:                handler,
		MsgHandlerGoroutineNum:    runtime.NumCPU() * 2,
		MsgHandlerExit:            []chan int{},
		ScheduledTaskGoroutineNum: len(scheduledTasks),
		ScheduledTasks:            scheduledTasks,
		ScheduledTaskExit:         []chan int{},
	}

	if server.Opts.mode == LastMsg {
		server.MsgHandlerGoroutineNum = 1
	}

	if handler == nil {
		server.MsgHandlerGoroutineNum = 0
	}

	server.State = Stopped
	Servers.Store(serverName, server)
	return server, nil
}

func StopServer(serverName string) error {
	load, ok := Servers.Load(serverName)
	if !ok {
		return fmt.Errorf("server %s not found", serverName)
	}
	s := load.(*Server)
	s.Stop()
	return nil
}

func SendMessage(ctx context.Context, serverName string, msg interface{}) (err error) {
	load, ok := Servers.Load(serverName)
	if !ok {
		return fmt.Errorf("server %s not found", serverName)
	}
	s := load.(*Server)
	err = s.SendMessage(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}

func SendMessageWithResult(ctx context.Context, serverName string, msg interface{}) (resp interface{}, err error) {
	load, ok := Servers.Load(serverName)
	if !ok {
		return nil, fmt.Errorf("server %s not found", serverName)
	}
	s := load.(*Server)
	return s.SendMessageWithResult(ctx, msg)
}

func ExitTaskService() {
	Servers.Range(func(key, value interface{}) bool {
		server := value.(*Server)
		server.Stop()
		return true
	})
}

func NumOfTaskService() int {
	count := 0
	Servers.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
