package scheduler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tsukaychan/mercury/pkg/ginx"
	"github.com/tsukaychan/mercury/pkg/gormx/connpool"
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/pkg/migrator"
	"github.com/tsukaychan/mercury/pkg/migrator/events"
	"github.com/tsukaychan/mercury/pkg/migrator/validator"
	"gorm.io/gorm"
	"sync"
	"time"
)

type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DualWritePool
	pattern    connpool.Pattern
	cancelFull func()
	cancelIncr func()
	producer   events.Producer
	l          logger.Logger
}

func NewScheduler[T migrator.Entity](
	src *gorm.DB,
	dst *gorm.DB,
	pool *connpool.DualWritePool,
	producer events.Producer,
	l logger.Logger,
) *Scheduler[T] {
	return &Scheduler[T]{
		src:        src,
		dst:        dst,
		pool:       pool,
		pattern:    connpool.PatternSrcOnly,
		cancelFull: func() {},
		cancelIncr: func() {},
		producer:   producer,
		l:          l,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.GET("/status", ginx.Wrap(s.Status))
	server.POST("/next", ginx.Wrap(s.Next))
	server.POST("/prev", ginx.Wrap(s.Prev))
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/start", ginx.WrapReq[StartIncrementRequest](s.StartIncrementValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrementValidation))
}

func (s *Scheduler[T]) Status(c *gin.Context) (ginx.Result, error) {
	return ginx.Result{Data: s.pattern.String()}, nil
}

func (s *Scheduler[T]) Next(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ptn := min(connpool.PatternDstOnly, s.pattern+1)
	s.pattern = ptn
	s.pool.WithPattern(ptn)

	return ginx.Result{Data: s.pattern.String()}, nil
}

func (s *Scheduler[T]) Prev(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ptn := max(connpool.PatternSrcOnly, s.pattern-1)

	s.pattern = ptn
	s.pool.WithPattern(ptn)

	return ginx.Result{Data: s.pattern.String()}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, s.producer, migrator.DirectionToTarget, s.l), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, s.producer, migrator.DirectionToBase, s.l), nil
	default:
		return nil, fmt.Errorf("pattern: %s", s.pattern.String())
	}
}

func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{Msg: "internal error"}, err
	}

	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		// cancel previous
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("exit full validation", logger.Error(err))
		}
	}()

	return ginx.Result{Msg: "success"}, nil
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cancelFull()

	return ginx.Result{Msg: "success"}, nil
}

type StartIncrementRequest struct {
	// Utime is the increment validation start time
	Utime int64 `json:"utime"`
	// Interval is the increment validation interval
	Interval int64 `json:"interval"`
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context, req StartIncrementRequest) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// cancel previous
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{Msg: "internal error"}, err
	}

	v.WithLoopInterval(time.Duration(req.Interval) * time.Millisecond).WithUtime(req.Utime)
	var ctx context.Context
	ctx, s.cancelIncr = context.WithCancel(context.Background())

	go func() {
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("exit increment validation", logger.Error(err))
		}
	}()

	return ginx.Result{Msg: "success"}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cancelIncr()

	return ginx.Result{Msg: "success"}, nil
}
