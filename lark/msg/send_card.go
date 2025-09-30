package lark

import (
	"context"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// SendCard 发送卡片
func SendCard(ctx context.Context, newCard, chatId string) error {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeInteractive).
			ReceiveId(chatId).
			Content(newCard).
			Build()).
		Build()
	return sendMsg(ctx, req)
}
