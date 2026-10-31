package module

import "testing"

func TestSettingValidate(t *testing.T) {
	ok := []Setting{
		{Key: "title", Label: "Başlık", Type: SettingText, Default: ""},
		{Key: "limit", Label: "Sınır", Type: SettingNumber, Default: 10, Min: Ptr(1.0), Max: Ptr(100.0)},
		{Key: "on", Label: "Açık", Type: SettingBool, Default: true},
		{Key: "mode", Label: "Mod", Type: SettingSelect, Default: "a", Options: []string{"a", "b"}},
	}
	for _, s := range ok {
		if err := s.Validate(); err != nil {
			t.Errorf("%s: beklenmeyen hata: %v", s.Key, err)
		}
	}
	bad := []Setting{
		{Key: "Bad", Label: "x", Type: SettingText, Default: ""},
		{Key: "x", Label: " ", Type: SettingText, Default: ""},
		{Key: "x", Label: "x", Type: "color", Default: ""},
		{Key: "x", Label: "x", Type: SettingSelect, Default: "a"},
		{Key: "x", Label: "x", Type: SettingSelect, Default: "c", Options: []string{"a"}},
		{Key: "x", Label: "x", Type: SettingText, Default: "", Options: []string{"a"}},
		{Key: "x", Label: "x", Type: SettingText, Default: "", Min: Ptr(1.0)},
		{Key: "x", Label: "x", Type: SettingNumber, Default: 5, Min: Ptr(9.0), Max: Ptr(1.0)},
		{Key: "x", Label: "x", Type: SettingNumber, Default: 0, Min: Ptr(1.0)},
		{Key: "x", Label: "x", Type: SettingBool, Default: "true"},
	}
	for i, s := range bad {
		if err := s.Validate(); err == nil {
			t.Errorf("%d: hata bekleniyordu: %+v", i, s)
		}
	}
}

func TestSettingCheckNormalizes(t *testing.T) {
	s := Setting{Key: "n", Label: "n", Type: SettingNumber, Default: 1}
	v, err := s.Check(int64(7))
	if err != nil || v != 7.0 {
		t.Fatalf("got %v, %v", v, err)
	}
	vals := SettingValues{"n": v, "s": "x", "b": true}
	if vals.Number("n") != 7 || vals.String("s") != "x" || !vals.Bool("b") || vals.String("yok") != "" {
		t.Fatalf("yardımcılar hatalı: %v", vals)
	}
}
