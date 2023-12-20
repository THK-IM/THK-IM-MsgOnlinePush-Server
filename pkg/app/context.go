package app

import (
	"github.com/thk-im/thk-im-base-server/conf"
	"github.com/thk-im/thk-im-base-server/server"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
	"github.com/thk-im/thk-im-msgonlinepush-server/pkg/loader"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
)

type Context struct {
	*server.Context
}

func (c *Context) UserApi() userSdk.UserApi {
	if c.Context.SdkMap["user_api"] == nil {
		return nil
	}
	return c.Context.SdkMap["user_api"].(userSdk.UserApi)
}

func (c *Context) MsgApi() msgSdk.MsgApi {
	if c.Context.SdkMap["msg_api"] == nil {
		return nil
	}
	return c.Context.SdkMap["msg_api"].(msgSdk.MsgApi)
}

func (c *Context) Init(config *conf.Config) {
	c.Context = &server.Context{}
	c.Context.Init(config)
	c.Context.SdkMap = loader.LoadSdks(c.Config().Sdks, c.Logger())
}
