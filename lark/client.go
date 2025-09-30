package lark

import (
	"os"

	lark "github.com/larksuite/oapi-sdk-go/v3"
)

// Lark App: Broadcast
var (
	appId     = os.Getenv("LARK_APP_ID")
	appSecret = os.Getenv("LARK_APP_SECRET")
)

var larkClient *lark.Client

func init() {
	larkClient = lark.NewClient(appId, appSecret)
}

func GetLarkClient() *lark.Client {
	return larkClient
}
