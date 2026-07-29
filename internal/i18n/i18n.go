package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 默认回退 locale（对标 JsonMessageSource defaultLocale = zh-CN）
const DefaultLocale = "zh-CN"

type ctxKey struct{}

// 请求级 locale 写入 context
func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, ctxKey{}, locale)
}

// 从 context 读取已解析 locale；缺省为 zh-CN
func LocaleFrom(ctx context.Context) string {
	if ctx == nil {
		return DefaultLocale
	}
	if v, ok := ctx.Value(ctxKey{}).(string); ok && v != "" {
		return v
	}
	return DefaultLocale
}

// MessageSource：基于 JSON 文件的消息源（对标 Boot JsonMessageSource）
//
// 扫描规则：i18n/{locale}/**/*.json，文件路径映射为 key 前缀
// 例如 i18n/zh-CN/channel/error.json → key 前缀 channel.error
type MessageSource struct {
	// locale -> (key -> message)
	messages map[string]map[string]string
	mu       sync.RWMutex
}

// 从 embed/OS FS 的 i18n 根目录加载全部 locale
func Load(fsys fs.FS, root string) (*MessageSource, error) {
	ms := &MessageSource{messages: make(map[string]map[string]string)}
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".json") {
			return nil
		}
		// root/{locale}/.../file.json
		rel, err := pathRel(root, p)
		if err != nil {
			return err
		}
		parts := strings.Split(rel, "/")
		if len(parts) < 2 {
			return nil
		}
		locale := parts[0]
		// 路径段（去掉 locale 与 .json）→ key 前缀
		fileParts := parts[1:]
		last := fileParts[len(fileParts)-1]
		fileParts[len(fileParts)-1] = strings.TrimSuffix(last, ".json")
		prefix := strings.Join(fileParts, ".")

		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		var nested map[string]any
		if err := json.Unmarshal(data, &nested); err != nil {
			return fmt.Errorf("parse %s: %w", p, err)
		}
		bucket := ms.messages[locale]
		if bucket == nil {
			bucket = make(map[string]string)
			ms.messages[locale] = bucket
		}
		for k, v := range flattenMessages(nested, "") {
			bucket[prefix+"."+k] = v
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ms, nil
}

func pathRel(root, p string) (string, error) {
	root = strings.TrimSuffix(root, "/")
	if root == "." || root == "" {
		return p, nil
	}
	if !strings.HasPrefix(p, root+"/") && p != root {
		return "", fmt.Errorf("path %s not under %s", p, root)
	}
	if p == root {
		return "", nil
	}
	return strings.TrimPrefix(p, root+"/"), nil
}

// flattenMessages：递归扁平化嵌套 JSON 对象为点分 key→string 映射
//
// 支持 Java 端 channel/error.json 里的嵌套结构，例如：
//
//	{"transportEncrypt": {"keyInvalid": "..."}}  →  {"transportEncrypt.keyInvalid": "..."}
//
// 字符串值原样保留；数字/布尔等非容器值用 fmt.Sprint 兜底；JSON 数组当前不出现于 i18n 文件，遇则按元素下标展开。
func flattenMessages(m map[string]any, prefix string) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]any:
			for kk, vv := range flattenMessages(val, key) {
				out[kk] = vv
			}
		default:
			out[key] = fmt.Sprint(val)
		}
	}
	return out
}

// 按 locale 取消息；缺 key 回退 zh-CN；仍缺则返回 key 本身
func (ms *MessageSource) Get(locale, key string, args ...any) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	locale = ResolveResourceLocale(locale)
	msg := ms.lookup(locale, key)
	if msg == "" && locale != DefaultLocale {
		msg = ms.lookup(DefaultLocale, key)
	}
	if msg == "" {
		return key
	}
	return formatMessage(msg, args...)
}

func (ms *MessageSource) lookup(locale, key string) string {
	if bucket, ok := ms.messages[locale]; ok {
		return bucket[key]
	}
	return ""
}

// Java MessageFormat 风格：将 {0} {1} ... 替换为 args
func formatMessage(pattern string, args ...any) string {
	if len(args) == 0 {
		return pattern
	}
	out := pattern
	for i, a := range args {
		ph := "{" + strconv.Itoa(i) + "}"
		out = strings.ReplaceAll(out, ph, fmt.Sprint(a))
	}
	return out
}

// Accept-Language / 客户端变体 → 资源目录 locale（对标 JsonMessageSource.LOCALE_ALIASES）
var localeAliases = map[string]string{
	"zh-hans":    "zh-CN",
	"zh-hans-cn": "zh-CN",
	"zh-cn":      "zh-CN",
	"zh":         "zh-CN",
	"zh-hant":    "zh-TW",
	"zh-hant-tw": "zh-TW",
	"zh-tw":      "zh-TW",
	"zh-hant-hk": "zh-HK",
	"zh-hk":      "zh-HK",
	"zh-hant-mo": "zh-HK",
	"zh-mo":      "zh-HK",
	"en":         "en-US",
	"en-us":      "en-US",
	"en-gb":      "en-US",
	"ja":         "ja-JP",
	"ja-jp":      "ja-JP",
	"ko":         "ko-KR",
	"ko-kr":      "ko-KR",
	"id":         "id-ID",
	"id-id":      "id-ID",
	"vi":         "vi-VN",
	"vi-vn":      "vi-VN",
	"th":         "th-TH",
	"th-th":      "th-TH",
	"ms":         "ms-MY",
	"ms-my":      "ms-MY",
}

// 将任意 locale 标签规范到资源目录名
func ResolveResourceLocale(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return DefaultLocale
	}
	// 取首段（去掉 ;q=）
	if i := strings.IndexByte(tag, ';'); i >= 0 {
		tag = strings.TrimSpace(tag[:i])
	}
	lower := strings.ToLower(tag)
	if mapped, ok := localeAliases[lower]; ok {
		return mapped
	}
	// 已是标准目录名（大小写敏感目录：zh-CN）
	for _, known := range []string{
		"zh-CN", "en-US", "zh-TW", "zh-HK", "ja-JP", "ko-KR",
		"id-ID", "vi-VN", "th-TH", "ms-MY",
	} {
		if strings.EqualFold(tag, known) {
			return known
		}
	}
	// 仅语言码回退
	if i := strings.IndexByte(lower, '-'); i > 0 {
		if mapped, ok := localeAliases[lower[:i]]; ok {
			return mapped
		}
	}
	return DefaultLocale
}

// 解析 Accept-Language，取 q 值最高项并映射到资源 locale
func ParseAcceptLanguage(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return DefaultLocale
	}
	type item struct {
		tag string
		q   float64
	}
	var items []item
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag := part
		q := 1.0
		if i := strings.IndexByte(part, ';'); i >= 0 {
			tag = strings.TrimSpace(part[:i])
			rest := strings.TrimSpace(part[i+1:])
			if strings.HasPrefix(strings.ToLower(rest), "q=") {
				if v, err := strconv.ParseFloat(strings.TrimSpace(rest[2:]), 64); err == nil {
					q = v
				}
			}
		}
		items = append(items, item{tag: tag, q: q})
	}
	if len(items) == 0 {
		return DefaultLocale
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].q > items[j].q
	})
	return ResolveResourceLocale(items[0].tag)
}

// 全局 MessageSource（由 main 初始化）
var global *MessageSource

func SetGlobal(ms *MessageSource) {
	global = ms
}

// T：按 context locale 取本地化消息（国际化处：业务错误/成功文案）
func T(ctx context.Context, key string, args ...any) string {
	if global == nil {
		return key
	}
	return global.Get(LocaleFrom(ctx), key, args...)
}
