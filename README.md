# THK-IM-MsgOnlinePush-Server

## 启动服务

- online push服务

```
go run main.go --config-path etc/msg_online_push_server.yaml

```

## 构建镜像

```
docker build -t thk-im/msg-online-push-server:v1  -f ./deploy/.Dockerfile .
```

