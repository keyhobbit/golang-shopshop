package utils

import (
	"fmt"
	"html/template"
	"math"
	"regexp"
	"strings"
	"time"
)

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatPrice": func(price float64) string {
			if price == 0 {
				return "Liên hệ"
			}
			p := int64(price)
			s := fmt.Sprintf("%d", p)
			n := len(s)
			if n <= 3 {
				return s + "₫"
			}
			var parts []string
			for n > 0 {
				end := n
				start := n - 3
				if start < 0 {
					start = 0
				}
				parts = append([]string{s[start:end]}, parts...)
				n = start
			}
			return strings.Join(parts, ".") + "₫"
		},
		"salePercent": func(original, sale float64) int {
			if original <= 0 || sale <= 0 || sale >= original {
				return 0
			}
			return int(math.Round(((original - sale) / original) * 100))
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02/01/2006")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("02/01/2006 15:04")
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul":     func(a, b float64) float64 { return a * b },
		"mulInt":  func(a float64, b int) float64 { return a * float64(b) },
		"statusBadge": func(status string) template.HTML {
			colors := map[string]string{
				"pending":   "bg-yellow-100 text-yellow-800",
				"confirmed": "bg-blue-100 text-blue-800",
				"shipping":  "bg-purple-100 text-purple-800",
				"delivered": "bg-green-100 text-green-800",
				"cancelled": "bg-red-100 text-red-800",
			}
			labels := map[string]string{
				"pending":   "Chờ xử lý",
				"confirmed": "Đã xác nhận",
				"shipping":  "Đang giao",
				"delivered": "Đã giao",
				"cancelled": "Đã hủy",
			}
			cls := colors[status]
			lbl := labels[status]
			if cls == "" {
				cls = "bg-gray-100 text-gray-800"
			}
			if lbl == "" {
				lbl = status
			}
			return template.HTML(fmt.Sprintf(`<span class="px-2 py-1 text-xs font-medium rounded-full %s">%s</span>`, cls, lbl))
		},
		"dict": func(values ...any) map[string]any {
			m := make(map[string]any)
			for i := 0; i < len(values)-1; i += 2 {
				key, ok := values[i].(string)
				if ok {
					m[key] = values[i+1]
				}
			}
			return m
		},
	}
}

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
