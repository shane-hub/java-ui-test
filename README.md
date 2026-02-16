# 翻鸡缺一门（V8.0）规则引擎 MVP

该仓库当前实现了一个 **Swift/SwiftUI 方向** 的最小可运行版本，重点落地：

- 牌型优先级与“就低不就高”（清一色特判最高优先）
- 自摸、一炮多响、抢杠胡多响（3 倍全包）
- 杠后点炮杠分清零
- 鸡分对冲 + 未听牌全赔
- 流局时杠分清零且不结算鸡分

## 目录

- `Sources/MahjongRulesEngine`: 规则领域模型与结算引擎
- `Sources/MahjongDemoApp`: 命令行演示入口
- `Tests/MahjongRulesEngineTests`: 规则单元测试

## 运行

```bash
swift test
swift run MahjongDemoApp
```

## iOS/SwiftUI 接入说明

`SettlementViewModel.swift` 中包含了 `SettlementSandboxView`（`#if canImport(SwiftUI)` 包裹），
可直接在 iOS App target 中引用并作为规则沙盒页面。

## 下一步建议

1. 将 `RoundInput` 扩展为完整牌局流水（摸打吃碰杠胡事件）。
2. 增加“九条触发金鸡”的牌面判定输入。
3. 接入 Sign in with Apple / 微信登录 / Universal Link。
4. 添加局内语音素材与战绩回放 JSON 存储 API。
