package taskservice

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

type GoState int

const (
	Stopped GoState = iota
	Started
)

type MessageHandler func(ctx context.Context, msg interface{}, num int) (resp interface{}, err error)
type ScheduledHandler func(num int)

type ScheduledTask struct {
	Task ScheduledHandler
	Time time.Duration
}

type Message struct {
	ch      chan interface{}
	message interface{}
	err     error
	ctx     context.Context
	sync    bool
}

const LastMsg = "LastMsg"

type Options struct {
	mode     string
	capacity int
}

type Option func(*Options) error

func WithOptionLastMsg() Option {
	return func(opt *Options) error {
		opt.mode = LastMsg
		return nil
	}
}

func NewDefaultOptions() *Options {
	return &Options{
		capacity: 4096,
	}
}

type Server struct {
	Name                      string
	Opts                      *Options
	Queue                     chan *Message
	State                     GoState
	MsgHandlerGoroutineNum    int
	MsgHandler                MessageHandler
	MsgHandlerExit            []chan int
	ScheduledTaskGoroutineNum int
	ScheduledTasks            []ScheduledTask
	ScheduledTaskExit         []chan int
}

func (server *Server) Go() {
	var wg sync.WaitGroup
	for i := 0; i < server.MsgHandlerGoroutineNum; i++ {
		number := i
		wg.Add(1)
		go func() {
			wg.Done()
			for {
				select {
				case msg := <-server.Queue:
					if server.Opts.mode == LastMsg {
						for {
							if len(server.Queue) > 1 {
								<-server.Queue
							} else if len(server.Queue) == 1 {
								msg = <-server.Queue
								break
							} else {
								break
							}
						}
					}
					if server.MsgHandler != nil {
						resp, err := server.MsgHandler(msg.ctx, msg.message, number)
						if msg.sync {
							msg.err = err
							msg.ch <- resp
						}
					}
				case <-server.MsgHandlerExit[number]:
					return
				}
			}
		}()
	}
	for j := 0; j < server.ScheduledTaskGoroutineNum; j++ {
		number := j
		wg.Add(1)
		go func() {
			wg.Done()
			server.ScheduledTasks[number].Task(number)
			for {
				select {
				case <-server.ScheduledTaskExit[number]:
					return
				case <-time.After(server.ScheduledTasks[number].Time):
					server.ScheduledTasks[number].Task(number)

				}
			}
		}()
	}
	wg.Wait()
	server.State = Started
	return
}

func (server *Server) Stop() {
	if server.State == Started {
		server.State = Stopped
		for i := 0; i < server.MsgHandlerGoroutineNum; i++ {
			server.MsgHandlerExit[i] <- 1
		}
		for i := 0; i < server.ScheduledTaskGoroutineNum; i++ {
			server.ScheduledTaskExit[i] <- 1
		}
		close(server.Queue)
		Servers.Delete(server.Name)
		fmt.Println(server.Name, " Server stopped")
	}
}

func (server *Server) SendMessage(ctx context.Context, msg interface{}) (err error) {
	if server.State == Stopped {
		return fmt.Errorf("server is stopped")
	}
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("server is stopped")
			return
		}
	}()
	m := &Message{
		ch:      nil,
		message: msg,
		err:     nil,
		ctx:     ctx,
		sync:    false,
	}
	server.Queue <- m
	return nil
}

func (server *Server) SendMessageWithResult(ctx context.Context, msg interface{}) (resp interface{}, err error) {
	if server.State == Stopped {
		return nil, fmt.Errorf("server is stopped")
	}
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("server is stopped")
			return
		}
	}()
	m := &Message{
		ch:      make(chan interface{}),
		message: msg,
		err:     nil,
		ctx:     ctx,
		sync:    true,
	}
	server.Queue <- m
	select {
	case resp = <-m.ch:
		close(m.ch)
		return resp, m.err
	case <-ctx.Done():
		go func(m *Message) {
			<-m.ch
			close(m.ch)
		}(m)
	}
	return nil, ctx.Err()
}

func RandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		b = []byte(time.Now().String())
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
