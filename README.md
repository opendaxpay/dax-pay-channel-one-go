# DaxPay Channel One (Go) — 通道适配实验副本

与 Java Boot 版 [`dax-pay-channel-one`](../dax-pay-channel-one/) 并列的 **Gin** 实现，端口同为 **20100**，**勿与 Boot / Quarkus 同时启动**。

首期仅打通契约骨架：**统一 `DaxResult`、Accept-Language i18n、W3C 链路追踪 + `x-trace-id`**；支付宝路由为占位，真实 SDK 对接留到二期。

## 技术栈

- Go 1.26+
- Gin
- OpenTelemetry（进程内 Tracer + W3C `traceparent`，默认不导出 OTLP）

## 运行

> 若 `go mod tidy` 访问 proxy.golang.org 超时，可先设：`$env:GOPROXY="https://goproxy.cn,direct"`

```bash
cd dax-pay-channel-one-go
go run ./cmd/server
```

或指定配置：

```bash
$env:DAXPAY_CONFIG="configs/config.yaml"
go run ./cmd/server
```

Health: http://127.0.0.1:20100/actuator/health

## 契约对齐（对标 Boot）

| 项 | 行为 |
|----|------|
| 响应 | HTTP **始终 200**，body `{code,msg,data}`，`code==0` 成功 |
| 错误码 | 0 / 10001–10008（`ChannelErrorCode`） |
| i18n | `Accept-Language` → `resources` embed JSON，key 如 `channel.error.*` |
| 追踪 | 入站 `traceparent`；响应头 `x-trace-id`；日志带 `traceId`/`spanId` |
| JSON | camelCase；`int64` 序列化为字符串；时间 UTC ISO |

## 路由（首期）

| Method | Path | 说明 |
|--------|------|------|
| GET | `/actuator/health` | 健康检查 |
| GET | `/internal/probe` | 契约探测（locale + trace） |
| POST | `/channel/alipay/{pay,sync,close,refund,refund-sync,callback/parse-pay,callback/parse-refund,auth/app-token}` | 占位 stub |

## 验证示例

```bash
curl -s http://127.0.0.1:20100/actuator/health

curl -s -D - -H "Accept-Language: en-US" \
  -H "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" \
  http://127.0.0.1:20100/internal/probe

curl -s -H "Accept-Language: zh-CN" -H "Content-Type: application/json" -d "{}" \
  http://127.0.0.1:20100/channel/alipay/pay
```

## 目录

```
cmd/server/          入口
configs/             配置
internal/            result / errcode / i18n / jsonx / middleware / handler / server
resources/i18n/      十语 error.json 源文件（embed 副本在 internal/i18n/i18n）
```
