# 翻鸡缺一门（Go 全栈 MVP）

本版本按你的要求，提供**前后端一体**开发结果，重点覆盖：

- 登录（演示账号体系）
- 大厅（房间列表 + 在线好友）
- 历史记录（局记录）
- 咱俩胜率（Head-to-Head）
- 回合结算（V8.0 核心规则）

## 技术栈

- 后端：Go（标准库 HTTP）
- 实时：`/ws/rooms/{roomId}/connect` 路由（当前以 SSE 形态承载实时推送）
- 存储边界：Redis / MySQL / Kafka（当前为内存适配，便于离线联调测试）
- 前端：内嵌 HTML + JS（服务端 `GET /`）

## 功能清单

### 1) 登录
- `POST /api/login`
- 首次用户名自动注册，重复登录校验密码。
- 登录成功后通过 `session_token` cookie 维持会话。

### 2) 大厅
- `GET /api/hall`
- 返回房间列表和在线好友列表。

### 3) 历史记录
- `GET /api/history?limit=20`
- 返回当前登录用户参与的最近对局记录。

### 4) 咱俩胜率
- `GET /api/stats/headtohead?opponent={userId}`
- 返回我方胜局、对方胜局、平局和胜率。

### 5) 结算与记录入库
- `POST /api/settle`
- 结算后会：
  - 更新房间状态（Redis边界）
  - 保存回放（MySQL边界）
  - 记录对局统计（用于历史和胜率）
  - 触发实时广播与消息发布（WebSocket/Kafka边界）

## 启动

```bash
go test ./...
go run ./cmd/server
```

启动后访问：
- 前端页面：`http://127.0.0.1:8080/`
- 健康检查：`http://127.0.0.1:8080/healthz`

## 联调测试（已内置）

新增 `internal/app/server_integration_test.go`，覆盖：

1. 两个用户登录
2. 查询大厅
3. 发起一局结算
4. 查询历史记录
5. 查询双方胜率

命令：

```bash
go test ./...
```

## 说明

由于当前容器网络限制，第三方依赖下载受限，因此 Redis/MySQL/Kafka 连接器使用内存实现，但接口边界已固定，后续替换真实驱动时不影响上层 API 与联调流程。
