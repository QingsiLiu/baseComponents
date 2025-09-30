package lark

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/QingsiLiu/baseComponents/lark"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func NewSendCard(header *larkcard.MessageCardHeader, elements ...larkcard.MessageCardElement) (string, error) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(false).
		Build()
	var aElementPool []larkcard.MessageCardElement
	aElementPool = append(aElementPool, elements...)
	// 卡片消息体
	return larkcard.NewMessageCard().
		Config(config).
		Header(header).
		Elements(aElementPool).
		String()
}

func NewSendCardWithoutHeader(elements ...larkcard.MessageCardElement) (string, error) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(false).
		Build()
	var aElementPool []larkcard.MessageCardElement
	aElementPool = append(aElementPool, elements...)
	// 卡片消息体
	cardContent, err := larkcard.NewMessageCard().
		Config(config).
		Elements(aElementPool).
		String()
	return cardContent, err
}

// 用于生成分割线
func WithSplitLine() larkcard.MessageCardElement {
	return larkcard.NewMessageCardHr().
		Build()
}

// 用于生成消息头
func WithHeader(title string, color string) *larkcard.
	MessageCardHeader {
	if title == "" {
		title = "🤖️机器人提醒"
	}
	return larkcard.NewMessageCardHeader().
		Template(color).
		Title(larkcard.NewMessageCardPlainText().
			Content(title).
			Build()).
		Build()
}

// 用于生成纯文本脚注
func WithNote(note string) larkcard.MessageCardElement {
	return larkcard.NewMessageCardNote().
		Elements([]larkcard.MessageCardNoteElement{larkcard.NewMessageCardPlainText().
			Content(note).
			Build()}).
		Build()
}

// 用于生成markdown消息体
func WithMainMd(msgs ...string) larkcard.MessageCardElement {
	fields := []*larkcard.MessageCardField{}
	for _, msg := range msgs {
		msg, i := ProcessMessage(msg)
		msg = ProcessNewLine(msg)
		msg = CleanTextBlock(msg)
		if i != nil {
			return nil
		}

		cardField := larkcard.NewMessageCardField().
			Text(larkcard.NewMessageCardLarkMd().
				Content(msg).
				Build()).
			IsShort(true).
			Build()

		fields = append(fields, cardField)
	}

	return larkcard.NewMessageCardDiv().
		Fields(fields).
		Build()
}

// 用于生成纯文本消息体
func WithMainText(msg string) larkcard.MessageCardElement {
	msg, i := ProcessMessage(msg)
	msg = CleanTextBlock(msg)
	if i != nil {
		return nil
	}
	return larkcard.NewMessageCardDiv().
		Fields([]*larkcard.MessageCardField{larkcard.NewMessageCardField().
			Text(larkcard.NewMessageCardPlainText().
				Content(msg).
				Build()).
			IsShort(false).
			Build()}).
		Build()
}

// 用于生成带有额外按钮的消息体
func WithMdAndExtraBtn(msg string, btn *larkcard.MessageCardEmbedButton) larkcard.MessageCardElement {
	msg, i := ProcessMessage(msg)
	msg = ProcessNewLine(msg)
	if i != nil {
		return nil
	}
	return larkcard.NewMessageCardDiv().
		Fields(
			[]*larkcard.MessageCardField{
				larkcard.NewMessageCardField().
					Text(larkcard.NewMessageCardLarkMd().
						Content(msg).
						Build()).
					IsShort(true).
					Build()}).
		Extra(btn).
		Build()
}

func NewBtn(content string, url string, value map[string]interface{}, typename larkcard.MessageCardButtonType) *larkcard.
	MessageCardEmbedButton {
	return larkcard.NewMessageCardEmbedButton().
		Type(typename).
		Url(url).
		Value(value).
		Text(larkcard.NewMessageCardPlainText().
			Content(content).
			Build())
}

func WithOneBtn(btn *larkcard.MessageCardEmbedButton) larkcard.
	MessageCardElement {
	return larkcard.NewMessageCardAction().
		Actions([]larkcard.MessageCardActionElement{btn}).
		Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
		Build()
}

func ProcessMessage(msg interface{}) (string, error) {
	msg = strings.TrimSpace(msg.(string))
	msgB, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	msgStr := string(msgB)

	if len(msgStr) >= 2 {
		msgStr = msgStr[1 : len(msgStr)-1]
	}
	return msgStr, nil
}

func ProcessNewLine(msg string) string {
	return strings.Replace(msg, "\\n", `
`, -1)
}

func ProcessQuote(msg string) string {
	return strings.Replace(msg, "\\\"", "\"", -1)
}

// 将字符中 \u003c 替换为 <  等等
func ProcessUnicode(msg string) string {
	regex := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	return regex.ReplaceAllStringFunc(msg, func(s string) string {
		r, _ := regexp.Compile(`\\u`)
		s = r.ReplaceAllString(s, "")
		i, _ := strconv.ParseInt(s, 16, 32)
		return string(rune(i))
	})
}

func CleanTextBlock(msg string) string {
	msg = ProcessNewLine(msg)
	msg = ProcessUnicode(msg)
	msg = ProcessQuote(msg)
	return msg
}

func sendMsg(ctx context.Context, req *larkim.CreateMessageReq) error {
	cli := lark.GetLarkClient()
	resp, err := cli.Im.Message.Create(ctx, req)
	slog.Info(fmt.Sprintf("send msg resp: %+v", resp))

	return err
}
