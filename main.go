package main

import (
	"github.com/thk-im/thk-im-base-server/conf"
	"github.com/thk-im/thk-im-msgonlinepush-server/pkg/app"
	"github.com/thk-im/thk-im-msgonlinepush-server/pkg/handler"
)

func main() {
	configPath := "etc/msg_online_push_server.yaml"
	config := conf.LoadConfig(configPath)

	appCtx := &app.Context{}
	appCtx.Init(config)
	handler.RegisterMsgOnlinePushHandlers(appCtx)

	appCtx.StartServe()
}
