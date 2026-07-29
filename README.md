# DaxPay Channel One (Go) — 通道适配实验副本

与 Java Boot 版 [`dax-pay-channel-one`](../dax-pay-channel-one/) 并列的 **Gin** 实现，端口同为 **20100**，**勿与 Boot / Quarkus 同时启动**。

已打通契约骨架，并自研实装 **支付宝 / 微信（直连+ISV）/ 银联商务(UMS) / 抖音** 四通道（无第三方支付 SDK）。

## 技术栈

- Go 1.26+
- Gin
- OpenTelemetry（进程内 Tracer + W3C `traceparent`，默认不导出 OTLP）
- 通道协议：自研 HTTP + 签名（支付宝 RSA2、微信 V3、UMS OPEN-BODY-SIG、抖音 DouyinPay-RSA）

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

### 传输加密密钥

`/channel/**` 路由强制 AES-256-GCM 双向加密，密钥须与主应用 `daxpay.channel.one.transport-encrypt.key` 完全一致（恰好 32 字节 UTF-8 字符）。两种配置方式：

- 配置文件 `configs/config.yaml` 的 `channel.transport-encrypt.key`（开发默认）
- 环境变量 `CHANNEL_TRANSPORT_KEY`（**优先于配置文件**，生产部署用此方式，勿提交真实密钥）

密钥为空或长度非 32 时启动失败。详见 `internal/transport/`、`internal/middleware/transport_encrypt.go`。

## 契约对齐（对标 Boot）

| 项 | 行为 |
|----|------|
| 响应 | HTTP **始终 200**，body `{code,msg,data}`，`code==0` 成功 |
| 错误码 | 0 / 10001–10008（`ChannelErrorCode`） |
| i18n | `Accept-Language` → embed JSON，key 如 `channel.error.*`（支持嵌套对象） |
| 追踪 | 入站 `traceparent`；响应头 `x-trace-id`；日志带 `traceId`/`spanId` |
| JSON | camelCase；`int64` 序列化为字符串；时间 UTC ISO |
| 传输加密 | `/channel/**` **AES-256-GCM 双向强制**；密文 `Base64(IV‖ct‖tag)`；头 `X-Dax-Payload-Encrypted`；`actuator/internal` 明文 |
| 通道 | 路径/字段与主应用 `*ChannelClient` 镜像；行为对齐 Boot 子应用 |

## 路由

| Method | Path | 说明 |
|--------|------|------|
| GET | `/actuator/health` | 健康检查 |
| GET | `/internal/probe` | 契约探测（locale + trace） |
| POST | `/channel/alipay/{pay,sync,close,refund,refund-sync,callback/parse-*,auth/app-token}` | 支付宝 |
| POST | `/channel/wechat/{pay,sync,close,refund,refund-sync,callback/parse-*}` | 微信直连 |
| POST | `/channel/wechat/isv/{pay,sync,close,refund,refund-sync}` | 微信服务商 |
| POST | `/channel/ums/{pay,sync,close,refund,refund-sync,callback/parse-*}` | 银联商务 |
| POST | `/channel/douyin/{pay,sync,close,refund,refund-sync,callback/parse-*}` | 抖音 |

## 验证示例

```bash
curl -s http://127.0.0.1:20100/actuator/health

curl -s -D - -H "Accept-Language: en-US" \
  -H "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" \
  http://127.0.0.1:20100/internal/probe
```

单元测试：

```bash
go test ./internal/alipay/... ./internal/wechat/... ./internal/ums/... ./internal/douyin/...
```

## 目录

```
cmd/server/          入口
configs/             配置
internal/
  alipay/            支付宝 OpenAPI + dto + service
  wechat/            微信 V3（直连/ISV）+ dto + service
  ums/               银联商务 openapi + dto + service
  douyin/            抖音 OpenAPI + dto + service
  channelerr/        统一 BizError
  httpclient/        共享 *http.Client（连接池复用）
  transport/         通道传输 AES-256-GCM 加解密器
  result/errcode/... 统一契约
  handler/server/    HTTP
  middleware/        Recovery/Otel/Locale/TransportEncrypt
resources/i18n/      十语 error.json 源文件（embed 副本在 internal/i18n/i18n）
```

## License

本项目基于 [GNU LGPL v3.0 或更高版本](./LICENSE) 协议开源，同时受[《用户授权使用协议》](./USER-AGREEMENT.txt)约束。在使用前请阅读上述协议，如果不同意请勿进行使用。
