package slug

import (
	"regexp"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

var (
	reInvalid   = regexp.MustCompile(`[^a-z0-9-]`)
	reMultiDash = regexp.MustCompile(`-+`)
	pinyinArgs  = pinyin.NewArgs()
)

// Generate 从标题生成 SEO 友好的 slug。
// 中文 → 拼音首字母，英文/数字保留，其他字符 → '-'。
// 例如 "Day 0：环境极速就位" → "day-0-huan-jing-ji-su-jiu-wei"
func Generate(title string) string {
	if title == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(title) * 3)

	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r >= 'A' && r <= 'Z' {
			b.WriteRune(r + 32) // 转小写
		} else if r >= 0x4e00 && r <= 0x9fff {
			// 中文字符 → 拼音
			py := pinyin.SinglePinyin(r, pinyinArgs)
			if len(py) > 0 {
				b.WriteString(py[0])
			} else {
				b.WriteByte('-')
			}
		} else {
			// 其他字符 → '-'
			b.WriteByte('-')
		}
	}

	slug := b.String()
	slug = reInvalid.ReplaceAllString(slug, "-")
	slug = reMultiDash.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}