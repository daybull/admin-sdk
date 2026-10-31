package module

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
)

// SettingType, bir modül ayarının değer tipidir.
type SettingType string

const (
	SettingText   SettingType = "text"   // tek satır metin
	SettingNumber SettingType = "number" // sayı (float64); Min/Max verilebilir
	SettingBool   SettingType = "bool"   // açık / kapalı
	SettingSelect SettingType = "select" // Options içinden biri
)

// Setting, modülün müşteri bazında değiştirilebilen bir ayarıdır. Panel, ayar
// formunu bu tanımdan otomatik oluşturur. Değerler çekirdeğin veritabanında
// tutulur; kayıtlı değer yoksa Default kullanılır.
//
// Gizli bilgi (şifre, API anahtarı) ayar olarak tutulmaz: değerler ayarları
// görme yetkisi olan herkese görünür.
type Setting struct {
	Key         string      `json:"key" doc:"Ayar anahtarı (küçük harf, rakam, alt çizgi)" example:"body_limit"`
	Label       string      `json:"label" example:"Not metni en fazla (karakter)"`
	Description string      `json:"description,omitempty"`
	Type        SettingType `json:"type" enum:"text,number,bool,select"`
	Default     any         `json:"default" doc:"Kayıtlı değer yoksa kullanılan değer"`
	Options     []string    `json:"options,omitempty" doc:"select için seçenekler"`
	Min         *float64    `json:"min,omitempty" doc:"number için alt sınır"`
	Max         *float64    `json:"max,omitempty" doc:"number için üst sınır"`
}

// MaxTextSetting, metin ayarının en fazla uzunluğudur (karakter).
const MaxTextSetting = 2000

var settingKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)

// Validate, ayar tanımının kendi içinde tutarlı olduğunu kontrol eder.
func (s Setting) Validate() error {
	if !settingKeyPattern.MatchString(s.Key) {
		return fmt.Errorf("ayar anahtarı geçersiz: %q (küçük harf, rakam, alt çizgi; en fazla 48 karakter)", s.Key)
	}
	if strings.TrimSpace(s.Label) == "" {
		return fmt.Errorf("%s ayarının etiketi boş", s.Key)
	}
	switch s.Type {
	case SettingText, SettingNumber, SettingBool:
		if len(s.Options) > 0 {
			return fmt.Errorf("%s ayarı select değil; Options verilmez", s.Key)
		}
	case SettingSelect:
		if len(s.Options) == 0 {
			return fmt.Errorf("%s ayarı select; en az bir seçenek gerekli", s.Key)
		}
	default:
		return fmt.Errorf("%s ayarının tipi geçersiz: %q", s.Key, s.Type)
	}
	if (s.Min != nil || s.Max != nil) && s.Type != SettingNumber {
		return fmt.Errorf("%s ayarı number değil; Min/Max verilmez", s.Key)
	}
	if s.Min != nil && s.Max != nil && *s.Min > *s.Max {
		return fmt.Errorf("%s ayarında Min, Max'tan büyük", s.Key)
	}
	if _, err := s.Check(s.Default); err != nil {
		return fmt.Errorf("%s ayarının varsayılan değeri: %w", s.Key, err)
	}
	return nil
}

// Check, değeri ayarın tipine göre doğrular ve normalleştirir: number her zaman
// float64, text ve select string, bool bool döner.
func (s Setting) Check(v any) (any, error) {
	switch s.Type {
	case SettingText:
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("metin bekleniyor")
		}
		if len([]rune(str)) > MaxTextSetting {
			return nil, fmt.Errorf("en fazla %d karakter olabilir", MaxTextSetting)
		}
		return str, nil
	case SettingBool:
		b, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("açık/kapalı (true/false) bekleniyor")
		}
		return b, nil
	case SettingSelect:
		str, ok := v.(string)
		if !ok || !slices.Contains(s.Options, str) {
			return nil, fmt.Errorf("şunlardan biri olmalı: %s", strings.Join(s.Options, ", "))
		}
		return str, nil
	case SettingNumber:
		f, ok := toFloat(v)
		if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("sayı bekleniyor")
		}
		if s.Min != nil && f < *s.Min {
			return nil, fmt.Errorf("en az %v olabilir", *s.Min)
		}
		if s.Max != nil && f > *s.Max {
			return nil, fmt.Errorf("en fazla %v olabilir", *s.Max)
		}
		return f, nil
	}
	return nil, fmt.Errorf("bilinmeyen ayar tipi: %q", s.Type)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// Ptr, Min/Max gibi isteğe bağlı alanlar için kısayoldur: module.Ptr(1.0).
func Ptr[T any](v T) *T { return &v }

// Settings, modülün kendi ayarlarını okumasını sağlar (Deps.Settings).
type Settings interface {
	// Values, tüm ayarları döner; kayıtlı değeri olmayanlar varsayılandır.
	Values(ctx context.Context) (SettingValues, error)
}

// SettingValues, ayar anahtarı → değer. Değerler Setting.Check ile
// normalleştirilmiştir; tip uymazsa yardımcılar sıfır değer döner.
type SettingValues map[string]any

func (v SettingValues) String(key string) string { s, _ := v[key].(string); return s }
func (v SettingValues) Bool(key string) bool     { b, _ := v[key].(bool); return b }
func (v SettingValues) Number(key string) float64 {
	f, _ := toFloat(v[key])
	return f
}
