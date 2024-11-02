package repository

import (
	"SimShare/internal/repository/cache"
	"context"
)

var (
	ErrCodeSetTooMany         = cache.ErrCodeSetTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(c *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: c,
	}
}
func (repo *CodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) error {
	_, err := repo.cache.Verify(ctx, biz, phone, inputCode)
	return err
}
