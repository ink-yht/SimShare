package retryable

import (
	"SimShare/internal/service/sms"
	"context"
)

// Service 小心并发问题
type Service struct {
	svc sms.Service
	// 重试
	retryCnt int
}

func (s Service) Send(ctx context.Context, tpl string, args []string, number ...string) error {
	err := s.svc.Send(ctx, tpl, args, number...)
	for err != nil && s.retryCnt < 10 {
		err = s.svc.Send(ctx, tpl, args, number...)
		s.retryCnt++
	}
	return err
}
