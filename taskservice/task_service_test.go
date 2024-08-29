package taskservice

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSvr(t *testing.T) {

	t.Run("server-no-name", func(t *testing.T) {
		s, _ := NewServer("", func(ctx context.Context, msg interface{}, num int) (resp interface{}, err error) {
			return msg, nil
		}, []ScheduledTask{{
			Task: func(num int) {
				return
			},
			Time: 0,
		}})

		s.Go()
		s.SendMessage(context.Background(), "1")
		resp, err := s.SendMessageWithResult(context.Background(), "2")
		assert.Nil(t, err)
		assert.Equal(t, resp, "2")
		s.Stop()
	})

	t.Run("server-name", func(t *testing.T) {
		s, _ := NewServer("server", func(ctx context.Context, msg interface{}, num int) (resp interface{}, err error) {
			return msg, nil
		}, []ScheduledTask{{
			Task: func(num int) {
				return
			},
			Time: 0,
		}})

		s.Go()
		SendMessage(context.Background(), "server", "1")
		resp, err := SendMessageWithResult(context.Background(), "server", "2")
		assert.Nil(t, err)
		assert.Equal(t, resp, "2")
		StopServer("server")
	})

	t.Run("lastMsg-mode", func(t *testing.T) {
		var total int
		s, _ := NewServer("lastMsg-mode-server", func(ctx context.Context, msg interface{}, num int) (resp interface{}, err error) {
			total++
			return msg, nil
		}, nil, WithOptionLastMsg())

		s.Go()
		_, err := s.SendMessageWithResult(context.Background(), "1")
		assert.Nil(t, err)

		for i := 0; i < 10; i++ {
			s.SendMessage(context.Background(), i)
		}
		time.Sleep(10 * time.Millisecond)
		assert.LessOrEqual(t, total, 10)
		s.Stop()
	})

	t.Run("timeout", func(t *testing.T) {
		s, _ := NewServer("timeout", func(ctx context.Context, msg interface{}, num int) (resp interface{}, err error) {
			time.Sleep(20 * time.Millisecond)
			return msg, nil
		}, []ScheduledTask{{
			Task: func(num int) {
				return
			},
			Time: 0,
		}})

		s.Go()
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, err := s.SendMessageWithResult(ctx, "1")
		assert.NotNil(t, err)
		s.Stop()
	})
}
