package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	"github.com/thk-im/thk-im-base-server/event"
	"github.com/thk-im/thk-im-base-server/websocket"
	msgDto "github.com/thk-im/thk-im-msgapi-server/pkg/dto"
	"github.com/thk-im/thk-im-msgonlinepush-server/pkg/app"
	"github.com/thk-im/thk-im-msgonlinepush-server/pkg/errorx"
	"strconv"
	"time"
)

func RegisterMsgOnlinePushHandlers(ctx *app.Context) {
	server := ctx.WebsocketServer()
	server.SetUidGetter(func(claim baseDto.ThkClaims) (int64, error) {
		if res, err := ctx.LoginApi().LoginByToken(claim); err != nil {
			ctx.Logger().WithFields(logrus.Fields(claim)).Errorf("GetUidByToken: %v, err: %v", claim, err)
			return 0, err
		} else {
			if res.User == nil {
				ctx.Logger().WithFields(logrus.Fields(claim)).Errorf("GetUidByToken: %v, err: %v", claim, res)
				return 0, errors.New("user info is nil")
			}
			return res.Id, nil
		}
	})
	server.SetOnClientConnected(func(client websocket.Client) {
		ctx.Logger().Infof("OnClientConnected: %v", client.Info())
		{
			// 下发服务器时间
			now := time.Now().UnixMilli()
			serverTimeBody, err := event.BuildSignalBody(event.SignalSyncTime, strconv.Itoa(int(now)))
			if err != nil {
				ctx.Logger().WithFields(logrus.Fields(client.Claims())).Errorf("OnClientConnected: %s", err.Error())
			}
			err = client.WriteMessage(serverTimeBody)
			if err != nil {
				ctx.Logger().WithFields(logrus.Fields(client.Claims())).Errorf("OnClientConnected: %s", err.Error())
			}
		}
		{
			// 下发连接id
			connIdBody, err := event.BuildSignalBody(event.SignalConnId, fmt.Sprintf("%d", client.Info().Id))
			if err != nil {
				ctx.Logger().Errorf("OnClientConnected: %s", err.Error())
			}
			err = client.WriteMessage(connIdBody)
			if err != nil {
				ctx.Logger().WithFields(logrus.Fields(client.Claims())).Errorf("OnClientConnected: %s", err.Error())
			}
		}
		// 发送用户上线事件
		{
			if userOnlineEvent, err := event.BuildUserOnlineEvent(ctx.NodeId(), true,
				client.Info().UId, client.Info().Id, client.Info().FirstOnLineTime, client.Claims().GetClientPlatform()); err != nil {
				ctx.Logger().WithFields(logrus.Fields(client.Claims())).Error("UserOnlineEvent Build err:", err)
			} else {
				if err = ctx.ServerEventPublisher().Pub(fmt.Sprintf("uid-%d", client.Info().UId), userOnlineEvent); err != nil {
					ctx.Logger().WithFields(logrus.Fields(client.Claims())).Error("UserOnlineEvent Pub err:", err)
				}
			}
		}
		// rpc通知api服务用户上线
		{
			sendUserOnlineStatus(ctx, client, true, true)
		}
	})

	server.SetOnClientClosed(func(client websocket.Client) {
		ctx.Logger().WithFields(logrus.Fields(client.Claims())).Infof("OnClientClosed: %v", client.Info())
		sendUserOnlineStatus(ctx, client, false, false)
	})

	server.SetOnClientMsgReceived(func(msg string, client websocket.Client) {
		signal := &event.SignalBody{}
		if err := json.Unmarshal([]byte(msg), signal); err != nil {
			ctx.Logger().WithFields(logrus.Fields(client.Claims())).Errorf("OnClientMsgReceived err: %s, msg: %s", err.Error(), msg)
		} else {
			err = onWsClientMsgReceived(ctx, client, signal.Type, signal.Body)
		}
	})

	errInit := server.Init()
	if errInit != nil {
		panic(errInit)
	}

	ctx.MsgPusherSubscriber().Sub(func(m map[string]interface{}) error {
		return onMqPushMsgReceived(m, server, ctx)
	})

	ctx.ServerEventSubscriber().Sub(func(m map[string]interface{}) error {
		return onMqServerEventReceived(m, server, ctx)
	})
}

func onMqPushMsgReceived(m map[string]interface{}, server websocket.Server, ctx *app.Context) error {
	ctx.Logger().Info("onMqPushMsgReceived", m)
	eventType, okType := m[event.PushEventTypeKey].(string)
	uIdsStr, okId := m[event.PushEventReceiversKey].(string)
	body, okBody := m[event.PushEventBodyKey].(string)
	if !okType || !okId || !okBody {
		return errorx.ErrMessageFormat
	}
	iType, eType := strconv.Atoi(eventType)
	if eType != nil {
		return errorx.ErrMessageFormat
	}
	uIds := make([]int64, 0)
	if err := json.Unmarshal([]byte(uIdsStr), &uIds); err != nil {
		return errorx.ErrMessageFormat
	}
	if content, err := event.BuildSignalBody(iType, body); err != nil {
		return err
	} else {
		return server.SendMessageToUsers(uIds, content)
	}
}

func onWsClientMsgReceived(ctx *app.Context, client websocket.Client, ty int, body *string) error {
	if ty == event.SignalHeatBeat {
		return onWsHeatBeatMsgReceived(ctx, client, body)
	}
	return nil
}

func onWsHeatBeatMsgReceived(ctx *app.Context, client websocket.Client, body *string) error {
	// 心跳
	heatBody, err := event.BuildSignalBody(event.SignalHeatBeat, "pong")
	if err != nil {
		return err
	}
	ctx.Logger().WithFields(logrus.Fields(client.Claims())).Info(client.Info())
	sendUserOnlineStatus(ctx, client, true, false)
	return client.WriteMessage(heatBody)
}

func sendUserOnlineStatus(ctx *app.Context, client websocket.Client, online, isLogin bool) {
	if ctx.MsgApi() == nil {
		return
	}
	now := time.Now().UnixMilli()
	client.SetLastOnlineTime(now)
	req := msgDto.PostUserOnlineReq{
		NodeId:      ctx.NodeId(),
		ConnId:      client.Info().Id,
		Online:      online,
		IsLogin:     isLogin,
		UId:         client.Info().UId,
		Platform:    client.Claims().GetClientPlatform(),
		TimestampMs: time.Now().UnixMilli(),
	}
	if err := ctx.MsgApi().PostUserOnlineStatus(&req, client.Claims()); err != nil {
		ctx.Logger().WithFields(logrus.Fields(client.Claims())).Errorf("sendUserOnlineStatus, err: %v", err)
	}
}

func onMqServerEventReceived(m map[string]interface{}, server websocket.Server, appCtx *app.Context) error {
	tp, okType := m[event.ServerEventTypeKey].(string)
	body, okBody := m[event.ServerEventBodyKey].(string)
	if !okType || !okBody {
		return errorx.ErrMessageFormat
	}
	if tp == event.ServerEventUserOnline {
		onlineBody := event.ParserOnlineBody(body)
		if onlineBody != nil {
			processUserOnlineEvent(onlineBody, server, appCtx)
		} else {
			appCtx.Logger().Error("ServerEventUserOnline, onlineBody is nil, body is: ", body)
		}
	}
	return nil
}

func processUserOnlineEvent(body *event.OnlineBody, server websocket.Server, appCtx *app.Context) {
	clients := server.GetUserClient(body.UserId)
	for _, client := range clients {
		if client.Info().Id != body.ConnId {
			if appCtx.Config().WebSocket.MultiPlatform == 0 {
				// 不允许跨平台
				if client.Info().FirstOnLineTime < body.OnLineTime {
					if err := server.RemoveClient(body.UserId, "connect at other device", client); err != nil {
						appCtx.Logger().Error("OnUserOnLineEvent RemoveClient err: ", err)
					}
				}
			} else if appCtx.Config().WebSocket.MultiPlatform == 1 {
				// 一个平台只能登录一台设备
				if client.Claims().GetClientPlatform() == body.Platform {
					if client.Info().FirstOnLineTime < body.OnLineTime {
						if err := server.RemoveClient(body.UserId, "connect at other device", client); err != nil {
							appCtx.Logger().Error("OnUserOnLineEvent RemoveClient err: ", err)
						}
					}
				}
			} else {
				// 随意跨平台登录
			}
		}
	}
}
