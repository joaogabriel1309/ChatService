package gateway

import (
	"context"

	"github.com/joaogabriel1309/ChatService/interno/domain/entity"
)

type ChatGateway interface {
	CriarChat(ctx context.Context, chat *entity.Chat) error
	ProcurarChatPorId(ctx context.Context, chatId string) (*entity.Chat, error)
	SalvarChat(ctx context.Context, chat *entity.Chat) error
}
