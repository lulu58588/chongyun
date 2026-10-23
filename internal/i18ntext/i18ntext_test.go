package i18ntext

import "testing"

// The language of a request: the ones starling speaks, of the tags a
// browser sends.
func TestAccepted(t *testing.T) {
	cases := []struct{ hdr, want string }{
		{"zh-CN,zh;q=0.9", "zh"},
		{"zh-TW", "zh-TW"},
		{"zh-HK", "zh-TW"},
		{"zh-Hant-TW", "zh-TW"},
		{"zh-Hans-CN", "zh"},
		{"en-US,en;q=0.9", ""},
		{"ja", "ja"},
		{"ja-JP", "ja"},
		{"de-DE", "de"},
		{"", ""},
	}
	for _, c := range cases {
		got := Accepted(c.hdr)
		if got != c.want {
			t.Errorf("Accepted(%q) = %q, want %q", c.hdr, got, c.want)
		}
	}
}

// The fallback order: the language's own table, then Simplified Chinese for
// Traditional Chinese, then English.
func TestPick(t *testing.T) {
	zh := map[string]string{"A": "甲"}
	zhTW := map[string]string{"B": "乙"}
	ja := map[string]string{"C": "さ"}
	de := map[string]string{"D": "D"}
	if got := Pick("zh", "A", zh, zhTW, ja, de); got != "甲" {
		t.Errorf("Pick(zh) = %q, want 甲", got)
	}
	if got := Pick("zh-TW", "B", zh, zhTW, ja, de); got != "乙" {
		t.Errorf("Pick(zh-TW) = %q, want 乙", got)
	}
	if got := Pick("ja", "C", zh, zhTW, ja, de); got != "さ" {
		t.Errorf("Pick(ja) = %q, want さ", got)
	}
	if got := Pick("de", "D", zh, zhTW, ja, de); got != "D" {
		t.Errorf("Pick(de) = %q, want D", got)
	}
	// Traditional Chinese with no entry of its own falls back to Simplified.
	if got := Pick("zh-TW", "A", zh, zhTW, ja, de); got != "甲" {
		t.Errorf("Pick(zh-TW, only in Simplified) = %q, want 甲", got)
	}
	// A language starling does not speak stays English.
	if got := Pick("fr", "A", zh, zhTW, ja, de); got != "A" {
		t.Errorf("Pick(fr) = %q, want A", got)
	}
	// A sentence nobody translated stays English.
	if got := Pick("zh", "E", zh, zhTW, ja, de); got != "E" {
		t.Errorf("Pick(zh, untranslated) = %q, want E", got)
	}
}

// A sentence with something in it: the translation has to keep the
// placeholder, or the English is used instead of a broken %!s.
func TestSprintf(t *testing.T) {
	zh := map[string]string{"%s is added.": "%s 已添加。"}
	zhTW := map[string]string{}
	ja := map[string]string{}
	de := map[string]string{}
	if got := Sprintf("zh", "%s is added.", []any{"x"}, zh, zhTW, ja, de); got != "x 已添加。" {
		t.Errorf("Sprintf(zh) = %q, want x 已添加。", got)
	}
	// A translation without the placeholder falls back to the English.
	if got := Sprintf("zh", "%s is added.", []any{"x"}, map[string]string{"%s is added.": "已添加。"}, zhTW, ja, de); got != "x is added." {
		t.Errorf("Sprintf(zh, translation without the placeholder) = %q, want the English", got)
	}
}
