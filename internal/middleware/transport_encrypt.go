// 通道传输加密中间件：/channel/** 强制 AES-256-GCM 双向透明加密
//
// 入站：检查 X-Dax-Payload-Encrypted 头 → 解密 body → 明文替换 body，下游 handler 透明。
// 出站：用 captureWriter 捕获 handler 响应 → 加密 body → 设加密头 → 回放到底层 writer。
// 失败：缺加密头/解密失败 → 400 明文 JSON（不加密，对齐 Java writeBadRequest）。
// panic：内层 recover 兜底（/channel/** 不依赖外层 Recovery），写 10006 后仍加密回放。
//
// 仅对 /channel/** 生效；/actuator/**、/internal/** 放行明文（对齐 Java shouldNotFilter）。
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
	"daxpay.open/dax-pay-channel-one-go/internal/transport"
)

// HeaderPayloadEncrypted：传输加密标识头（与 Java WebHeaderCode.X_DAX_PAYLOAD_ENCRYPTED 一致）
const HeaderPayloadEncrypted = "X-Dax-Payload-Encrypted"

// TransportEncrypt：/channel/** 强制 AES-256-GCM 双向透明加密中间件
//
// 中间件顺序要求：置于 Recovery → Otel → TraceIDHeader → Locale 之后（locale/trace 先就绪）。
func TransportEncrypt(encryptor *transport.Encryptor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅对 /channel/** 强制加解密；actuator/internal 放行明文
		if !strings.HasPrefix(c.Request.URL.Path, "/channel") {
			c.Next()
			return
		}
		ctx := c.Request.Context()

		// ===== 入站：解密请求体 =====
		raw, err := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		if err != nil {
			slog.WarnContext(ctx, "通道传输加密：读取请求体失败", "path", c.Request.URL.Path, "err", err)
			writePlainBadRequest(c, i18n.T(ctx, "channel.error.transportEncrypt.decryptFailed"))
			return
		}
		if len(raw) > 0 {
			// 请求未携带传输加密头
			flag := c.GetHeader(HeaderPayloadEncrypted)
			if !strings.EqualFold(flag, "true") {
				slog.WarnContext(ctx, "通道传输加密：请求未携带加密头", "path", c.Request.URL.Path)
				writePlainBadRequest(c, i18n.T(ctx, "channel.error.transportEncrypt.requestHeaderMissing"))
				return
			}
			plain, err := encryptor.Decrypt(string(raw))
			if err != nil {
				// err.Error() 为对应 i18n messageKey（keyInvalid/ciphertextInvalid/decryptFailed）
				slog.WarnContext(ctx, "通道传输解密失败", "path", c.Request.URL.Path)
				writePlainBadRequest(c, i18n.T(ctx, err.Error()))
				return
			}
			// 明文替换 body，Content-Type 改回 application/json，下游 handler 透明
			plainBytes := []byte(plain)
			c.Request.Body = io.NopCloser(bytes.NewReader(plainBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.ContentLength = int64(len(plainBytes))
		}

		// ===== 出站：captureWriter 替换 c.Writer，捕获 handler 响应 =====
		original := c.Writer
		capture := &captureWriter{ResponseWriter: original, body: &bytes.Buffer{}}
		c.Writer = capture

		// c.Next() + panic 兜底：确保 /channel/** 即使 panic 也走加密回放
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "panic recovered in channel handler", "err", r)
					// 重置为 200 + 10006 系统错误（对标 Recovery 中间件）
					capture.body.Reset()
					capture.ResponseWriter.WriteHeader(http.StatusOK)
					errResp := result.Fail(errcode.SystemError.Code, errcode.SystemError.Message(ctx))
					b, _ := json.Marshal(errResp)
					_, _ = capture.body.Write(b)
				}
			}()
			c.Next()
		}()

		// ===== 加密回放到底层 writer =====
		body := capture.body.Bytes()
		if len(body) == 0 {
			// 无 body（极端情况），原样回放 header/status
			original.WriteHeaderNow()
			return
		}
		ciphertext, err := encryptor.Encrypt(string(body))
		if err != nil {
			// 加密失败极少见（密钥/算法异常），兜底 500 明文
			slog.ErrorContext(ctx, "通道传输加密失败", "err", err)
			original.Header().Del(HeaderPayloadEncrypted)
			original.Header().Set("Content-Type", "application/json")
			original.WriteHeader(http.StatusInternalServerError)
			_, _ = original.Write([]byte(`{"code":500,"msg":"channel transport encrypt failed"}`))
			return
		}
		encryptedBytes := []byte(ciphertext)
		// handler 设的 header（含 x-trace-id 等）已在 original.Header()（capture 透传），只覆盖加密相关
		original.Header().Set(HeaderPayloadEncrypted, "true")
		original.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		original.Header().Set("Content-Length", strconv.Itoa(len(encryptedBytes)))
		original.WriteHeaderNow()
		_, _ = original.Write(encryptedBytes)
	}
}

// captureWriter：捕获 handler 写出的 body，header/status 透传给底层惰性记录
//
// 嵌入 gin.ResponseWriter，重写 Write/WriteString（只写 buffer）、WriteHeaderNow（no-op）。
// handler 调 c.JSON() 时：WriteHeader(200) 透传底层记录 status，Write(body) 进 buffer，
// WriteHeaderNow() 被拦截不真正发送 → c.Next() 后由中间件统一加密回放。
type captureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 只写 buffer，不触发底层 WriteHeaderNow
func (w *captureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// WriteString 只写 buffer
func (w *captureWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// WriteHeaderNow 拦截，不真正发送 header（延迟到回放）
func (w *captureWriter) WriteHeaderNow() {}

// writePlainBadRequest：写 400 明文 JSON（未加密），用于入站解密失败场景
func writePlainBadRequest(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"code": http.StatusBadRequest,
		"msg":  msg,
	})
}
