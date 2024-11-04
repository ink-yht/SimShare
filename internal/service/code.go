package service

import (
	"SimShare/internal/repository"
	"SimShare/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

const codeTelId = "1877556"

var ErrCodeSetTooMany = repository.ErrCodeSetTooMany
var ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{repo: repo, smsSvc: smsSvc}
}

// biz 区别业务场景

func (svc *codeService) Send(ctx context.Context, biz string, phone string) error {
	// 生成验证码
	code := svc.generateCode()
	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		// 有问题
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTelId, []string{code}, phone)
	if err != nil {
		// 这意味着，Redis 有这个验证码，但没发成功，用户根本收不到
		// err 可能是超时的 err，不知道发出去没
		// 要重试的的话，初始化的时候，传入一个积极就会重试的 smsSvc
		return err
	}
	return nil
}

func (svc *codeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *codeService) generateCode() string {
	// 六位数，num 在 0，999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的，加上前导 0 ， 000001
	return fmt.Sprintf("%06d", num)
}
