# 翻鸡缺一门后端（Go）

按你的要求，后端主栈已切换为：

- **Go (Golang)**
- **WebSocket**（实时房间推送）
- **Redis**（房间状态缓存）
- **MySQL**（回放/战绩落库）
- **Kafka**（牌局事件流）

## 当前实现（MVP）

### 1) V8.0 规则结算引擎
实现于 `internal/settlement/engine.go`，覆盖：

- 牌型“就低不就高”，`清一色`最高优先特判
- 自摸、一炮多响、抢杠胡多响（3 倍）
- 杠后点炮杠分作废
- 鸡分：听牌者对冲 + 未听牌全赔
- 流局：杠分归零、不翻鸡、不结鸡分

### 2) 服务接口
- `POST /api/settle`：输入回合数据，返回结算结果。
- `GET /ws/rooms/{roomId}/connect`：房间 WebSocket 连接。
- `GET /healthz`：健康检查。

### 3) 存储与消息
- Redis：保存房间结算状态快照
- MySQL：保存 `round_replay` 回放 JSON
- Kafka：发布房间结算事件（topic 可配置）

## 启动

```bash
go mod tidy
go test ./...
HTTP_ADDR=:8080 \
MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/fanji?parseTime=true' \
REDIS_ADDR='127.0.0.1:6379' \
KAFKA_BROKERS='127.0.0.1:9092' \
KAFKA_TOPIC='mahjong.room.events' \
go run ./cmd/server
```

## MySQL 建议建表

```sql
CREATE TABLE IF NOT EXISTS round_replay (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  room_id VARCHAR(64) NOT NULL,
  payload_json JSON NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 下一步（直接可排期）

1. 接入用户体系（Apple / 微信）与鉴权中间件。
2. 增加完整牌局事件模型（摸打吃碰杠胡）并沉淀回放轨迹。
3. 按房间分区策略设计 Kafka key 与消费者组。
4. 将“九条触发金鸡”从外部输入改为牌面自动判定。
5. 增加风控规则：同 IP、地理距离审计。
