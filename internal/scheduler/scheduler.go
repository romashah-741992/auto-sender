package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/romashah-741992/auto-sender/internal/messages"
)

type Scheduler struct {
	mu       sync.Mutex
	running  bool
	cancelFn context.CancelFunc

	service  *messages.Service
	interval time.Duration
}

func New(service *messages.Service, interval time.Duration) *Scheduler {
	return &Scheduler{
		service:  service,
		interval: interval,
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		log.Println("[scheduler] already running")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	s.running = true

	go s.run(ctx)
	log.Println("[scheduler] started")
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		log.Println("[scheduler] not running")
		return
	}

	s.cancelFn()
	s.running = false
	log.Println("[scheduler] stopped")
}

func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) run(ctx context.Context) {
	s.tickOnce(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tickOnce(ctx)
		}
	}
}

func (s *Scheduler) tickOnce(ctx context.Context) {
	log.Println("[scheduler] tick: processing up to 2 pending messages")
	if err := s.service.SendPendingMessages(ctx, 2); err != nil {
		log.Println("[scheduler] error:", err)
	}
}
