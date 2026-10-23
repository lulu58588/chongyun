package gui

import "testing"

// The text of a page, in each language, and where a language has no entry:
// to English, never to nothing.
func TestWebTextIn(t *testing.T) {
	const en = "Sign in"
	const zh = "登录"
	if got := webTextIn("zh", en); got != zh {
		t.Errorf("webTextIn(zh) = %q, want %q", got, zh)
	}
	if got := webTextIn("zh-TW", en); got != "登入" {
		t.Errorf("webTextIn(zh-TW) = %q, want 登入", got)
	}
	if got := webTextIn("ja", en); got != "サインイン" {
		t.Errorf("webTextIn(ja) = %q, want サインイン", got)
	}
	if got := webTextIn("de", en); got != "Anmelden" {
		t.Errorf("webTextIn(de) = %q, want Anmelden", got)
	}
	// A language starling does not speak stays English.
	if got := webTextIn("fr", en); got != en {
		t.Errorf("webTextIn(fr) = %q, want %q", got, en)
	}
	// A language with no entry for this text stays English.
	if got := webTextIn("zh", "a sentence nobody translated"); got != "a sentence nobody translated" {
		t.Errorf("webTextIn(zh, untranslated) = %q, want the English", got)
	}
	// Traditional Chinese falls back to Simplified before English.
	const long = "this link's key is not this starling web's: use the link it printed when it started"
	if got := webTextIn("zh-TW", long); got == long {
		t.Error("webTextIn(zh-TW) fell back to English where Simplified was there")
	}
}
