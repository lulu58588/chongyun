// Package i18ntext picks the words of the pages starling serves itself, in
// the language the user set or the browser asks for.
//
// The pages the app serves are translated by the front end (assets/i18n.js);
// the ones written by Go (the sign-in pages, the MCP authorization page, the
// page starling web shows when it turns a browser away) are plain HTML with
// no script, so their words are here, in a table of its own per package.
package i18ntext

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lulu58588/starling/internal/settings"
)

// Lang is the language a request asks for: what the user set in Settings
// (which says "system" for the browser's own), and the browser's
// Accept-Language when it does.
func Lang(r *http.Request) string {
	if r == nil {
		return ""
	}
	s := settings.Load()
	if s.Lang != "" && s.Lang != "system" {
		return s.Lang
	}
	return Accepted(r.Header.Get("Accept-Language"))
}

// Accepted is the language of an Accept-Language header, of the ones starling
// speaks: zh, zh-TW, ja, de, else "" (English).
func Accepted(h string) string {
	for _, part := range strings.Split(h, ",") {
		tag := strings.TrimSpace(part)
		if i := strings.IndexByte(tag, ';'); i >= 0 {
			tag = strings.TrimSpace(tag[:i])
		}
		switch strings.ToLower(tag) {
		case "zh", "zh-cn", "zh-hans", "zh-sg", "zh-hans-cn":
			return "zh"
		case "zh-tw", "zh-hk", "zh-mo", "zh-hant", "zh-hant-tw", "zh-hant-hk":
			return "zh-TW"
		case "ja", "ja-jp":
			return "ja"
		case "de", "de-de", "de-at", "de-ch":
			return "de"
		}
	}
	return ""
}

// Sprintf is fmt.Sprintf on the translation of en, falling back to the
// English when the translation lost a placeholder the English has (a
// sentence with a %s in it, translated without one).
func Sprintf(lang, en string, args []any, zh, zhTW, ja, de map[string]string) string {
	out := fmt.Sprintf(Pick(lang, en, zh, zhTW, ja, de), args...)
	if strings.Contains(out, "%!") {
		return fmt.Sprintf(en, args...)
	}
	return out
}

// Pick is the translation of en in lang, out of the tables a package holds:
// the language's own table, then Simplified Chinese for Traditional Chinese,
// then English. A sentence nobody translated shows English, never nothing.
func Pick(lang, en string, zh, zhTW, ja, de map[string]string) string {
	switch lang {
	case "zh":
		if t, ok := zh[en]; ok {
			return t
		}
	case "zh-TW":
		if t, ok := zhTW[en]; ok {
			return t
		}
		if t, ok := zh[en]; ok {
			return t
		}
	case "ja":
		if t, ok := ja[en]; ok {
			return t
		}
	case "de":
		if t, ok := de[en]; ok {
			return t
		}
	}
	return en
}
