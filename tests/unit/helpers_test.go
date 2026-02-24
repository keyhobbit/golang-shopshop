package unit

import (
	"testing"

	"shoop-golang/pkg/utils"
)

func TestFormatPrice(t *testing.T) {
	funcs := utils.TemplateFuncs()
	formatPrice := funcs["formatPrice"].(func(float64) string)

	tests := []struct {
		name  string
		price float64
		want  string
	}{
		{"zero", 0, "Liên hệ"},
		{"1000", 1000, "1.000₫"},
		{"1990000", 1990000, "1.990.000₫"},
		{"500", 500, "500₫"},
		{"100", 100, "100₫"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrice(tt.price)
			if got != tt.want {
				t.Errorf("formatPrice(%v) = %q, want %q", tt.price, got, tt.want)
			}
		})
	}
}

func TestSalePercent(t *testing.T) {
	funcs := utils.TemplateFuncs()
	salePercent := funcs["salePercent"].(func(float64, float64) int)

	tests := []struct {
		name     string
		original float64
		sale     float64
		want     int
	}{
		{"100k_80k", 100000, 80000, 20},
		{"100k_0", 100000, 0, 0},
		{"0_80k", 0, 80000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := salePercent(tt.original, tt.sale)
			if got != tt.want {
				t.Errorf("salePercent(%v, %v) = %d, want %d", tt.original, tt.sale, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	funcs := utils.TemplateFuncs()
	truncate := funcs["truncate"].(func(string, int) string)

	tests := []struct {
		name   string
		s      string
		length int
		want   string
	}{
		{"hello_world_5", "Hello World", 5, "Hello..."},
		{"hi_10", "Hi", 10, "Hi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.length)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.length, got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"hello_world", "Hello World", "hello-world"},
		{"trimmed_vietnamese", "  Tượng Phong Thủy  ", "t-ng-phong-th-y"},
		{"multiple_dashes", "test---multiple", "test-multiple"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.Slugify(tt.s)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := utils.TemplateFuncs()
	expected := []string{
		"formatPrice", "salePercent", "truncate", "safeHTML",
		"formatDate", "formatDateTime", "seq", "add", "sub",
		"mul", "mulInt", "statusBadge", "dict",
	}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			if _, ok := funcs[name]; !ok {
				t.Errorf("TemplateFuncs() missing expected function %q", name)
			}
		})
	}
}
