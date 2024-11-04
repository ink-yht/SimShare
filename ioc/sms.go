package ioc

import (
	"SimShare/internal/service/sms"
	"SimShare/internal/service/sms/memory"
)

func InitSmsService() sms.Service {
	return memory.NewService()
}
