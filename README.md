# DaxPay Channel One (Go) — 通道适配实验副本

与 Java Boot 版 [`dax-pay-channel-one`](../dax-pay-channel-one/) 并列的 **Gin** 实现，端口同为 **20100**，**勿与 Boot / Quarkus 同时启动**。

首期已打通契约骨架与 **支付宝 OpenAPI 八接口实装**（自研 RSA2 + `gateway.do`，支持公钥/证书模式）。微信等其它通道未实现。

## 技术栈

- Go 1.26+
- Gin
- OpenTelemetry（进程内 Tracer + W3C `traceparent`，默认不导出 OTLP）
- 支付宝：自研 OpenAPI（无第三方 alipay SDK）

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
| i18n | `Accept-Language` → embed JSON，key 如 `channel.error.*` |
| 追踪 | 入站 `traceparent`；响应头 `x-trace-id`；日志带 `traceId`/`spanId` |
| JSON | camelCase；`int64` 序列化为字符串；时间 UTC ISO |
| 支付宝 | 路径/字段与主应用 `AlipayChannelClient` 镜像；OpenAPI 行为对齐 Boot 服务 |

## 路由

| Method | Path | 说明 |
|--------|------|------|
| GET | `/actuator/health` | 健康检查 |
| GET | `/internal/probe` | 契约探测（locale + trace） |
| POST | `/channel/alipay/pay` | 下单（WAP/APP/PC/QR/BARCODE/JSAPI） |
| POST | `/channel/alipay/sync` | 查单 |
| POST | `/channel/alipay/close` | 关单/撤销 |
| POST | `/channel/alipay/refund` | 退款 |
| POST | `/channel/alipay/refund-sync` | 退款查询 |
| POST | `/channel/alipay/callback/parse-pay` | 支付回调验签解析 |
| POST | `/channel/alipay/callback/parse-refund` | 退款回调验签解析 |
| POST | `/channel/alipay/auth/app-token` | 换 app_auth_token |

## 验证示例

```bash
curl -s http://127.0.0.1:20100/actuator/health

curl -s -D - -H "Accept-Language: en-US" \
  -H "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" \
  http://127.0.0.1:20100/internal/probe
```

单元测试：

```bash
go test ./internal/alipay/...
```

## 目录

```
cmd/server/          入口
configs/             配置
internal/
  alipay/            OpenAPI 客户端 + dto + service
  result/errcode/... 统一契约
  handler/server/    HTTP
resources/i18n/      十语 error.json 源文件（embed 副本在 internal/i18n/i18n）
```
