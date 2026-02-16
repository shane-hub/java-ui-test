# 翻鸡缺一门（Go 后端 + 原生 iOS 客户端）

你反馈得对：之前截图是网页，不是原生 iOS。
本次已补齐 **SwiftUI 原生 iOS 客户端**，并对接现有 Go 后端接口完成联调链路。

## 当前结构

- `cmd/server`：Go 后端启动入口
- `internal/...`：后端业务（登录、大厅、历史、双方胜率、结算）
- `ios/NativeFanjiApp/NativeFanjiApp`：原生 iOS (SwiftUI) 客户端

## 已实现功能

### 后端 API
- `POST /api/login`
- `POST /api/logout`
- `GET /api/hall`
- `GET /api/history?limit=20`
- `GET /api/stats/headtohead?opponent={userId}`
- `POST /api/settle`

### 原生 iOS 页面（SwiftUI）
- 登录/注册页（演示账号）
- 大厅页（房间列表 + 在线好友）
- 历史记录页（最近战绩）
- 咱俩胜率查询
- 结算调试按钮（触发一局示例结算并回刷历史）

## 运行与联调

### 1) 启动后端
```bash
go test ./...
go run ./cmd/server
```

### 2) 启动 iOS 客户端
1. 在 Xcode 中打开 `ios/NativeFanjiApp` 目录（创建 iOS App 工程后，把该目录下 Swift 文件拖入 Target）。
2. 确保模拟器可访问 `http://127.0.0.1:8080`（同机运行时可用；真机请改为局域网 IP）。
3. 运行后即可联调登录、大厅、历史、胜率、结算。

## 说明

- 当前仓库为了在受限环境可测，Redis/MySQL/Kafka 为内存适配实现；
  API 边界已稳定，后续可无缝替换真实基础设施驱动。
