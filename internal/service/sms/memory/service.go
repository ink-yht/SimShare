package memory

import (
	"context"
	"fmt"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s Service) Send(ctx context.Context, tpl string, args []string, number ...string) error {
	fmt.Println("Send", tpl, args, number)
	return nil
}
