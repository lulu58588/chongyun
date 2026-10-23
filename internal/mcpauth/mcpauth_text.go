package mcpauth

// The words of the page an MCP server's authorization ends on. It is a page
// starling serves itself (plain HTML, no script, so no front-end
// translation), in the language the user set or the browser asks for.

import "github.com/lulu58588/starling/internal/i18ntext"

// mcpAuthText is a translator for the authorization page, in one language.
func mcpAuthText(lang string) func(string) string {
	return func(en string) string {
		return i18ntext.Pick(lang, en, mcpAuthTextZh, mcpAuthTextZhTW, mcpAuthTextJa, mcpAuthTextDe)
	}
}

// mcpAuthTextf is mcpAuthText for a sentence with something in it to fill in:
// the translation has to keep the placeholders, or the English is used.
func mcpAuthTextf(lang string) func(string, ...any) string {
	return func(en string, args ...any) string {
		return i18ntext.Sprintf(lang, en, args, mcpAuthTextZh, mcpAuthTextZhTW, mcpAuthTextJa, mcpAuthTextDe)
	}
}

// The text of the authorization page, in Chinese. The key is the English as
// written in the code, so a page that says something new falls back to
// English rather than to nothing.
var mcpAuthTextZh = map[string]string{
	"This sign-in link is out of date. Start the sign-in again from Starling.":                                    "这个登录链接已过期。请从 Starling 重新发起登录。",
	"Starling has this sign-in already. You can close this tab.":                                                  "Starling 已经完成了这次登录。可以关闭这个标签页。",
	"The sign-in was refused: %s":                                                                                 "登录被拒绝：%s",
	"Starling couldn't finish the sign-in: %s":                                                                    "Starling 没能完成登录：%s",
	"Starling couldn't keep the sign-in: %s":                                                                      "Starling 没能保存这次登录：%s",
	"Starling is signed in to %s. Every agent given it in the Library uses this sign-in. You can close this tab.": "Starling 已登录 %s。Library 里用到它的每个 Agent 都会用这次登录。可以关闭这个标签页。",
}

// The same, in Traditional Chinese.
var mcpAuthTextZhTW = map[string]string{
	"This sign-in link is out of date. Start the sign-in again from Starling.":                                    "這個登入連結已過期。請從 Starling 重新發起登入。",
	"Starling has this sign-in already. You can close this tab.":                                                  "Starling 已經完成了這次登入。可以關閉這個分頁。",
	"The sign-in was refused: %s":                                                                                 "登入被拒絕：%s",
	"Starling couldn't finish the sign-in: %s":                                                                    "Starling 沒能完成登入：%s",
	"Starling couldn't keep the sign-in: %s":                                                                      "Starling 沒能保存這次登入：%s",
	"Starling is signed in to %s. Every agent given it in the Library uses this sign-in. You can close this tab.": "Starling 已登入 %s。Library 裡用到它的每個 Agent 都會用這次登入。可以關閉這個分頁。",
}

// The same, in Japanese.
var mcpAuthTextJa = map[string]string{
	"This sign-in link is out of date. Start the sign-in again from Starling.":                                    "このサインインリンクは期限切れです。Starling からもう一度サインインしてください。",
	"Starling has this sign-in already. You can close this tab.":                                                  "このサインインはすでに完了しています。このタブは閉じて構いません。",
	"The sign-in was refused: %s":                                                                                 "サインインは拒否されました: %s",
	"Starling couldn't finish the sign-in: %s":                                                                    "Starling はサインインを完了できませんでした: %s",
	"Starling couldn't keep the sign-in: %s":                                                                      "Starling はサインインを保存できませんでした: %s",
	"Starling is signed in to %s. Every agent given it in the Library uses this sign-in. You can close this tab.": "Starling は %s にサインインしました。Library でそれを受け取ったすべてのエージェントがこのサインインを使います。このタブは閉じて構いません。",
}

// The same, in German.
var mcpAuthTextDe = map[string]string{
	"This sign-in link is out of date. Start the sign-in again from Starling.":                                    "Dieser Anmeldelink ist abgelaufen. Bitte starte die Anmeldung erneut aus Starling.",
	"Starling has this sign-in already. You can close this tab.":                                                  "Starling hat diese Anmeldung bereits abgeschlossen. Du kannst diesen Tab schließen.",
	"The sign-in was refused: %s":                                                                                 "Die Anmeldung wurde abgelehnt: %s",
	"Starling couldn't finish the sign-in: %s":                                                                    "Starling konnte die Anmeldung nicht abschließen: %s",
	"Starling couldn't keep the sign-in: %s":                                                                      "Starling konnte die Anmeldung nicht speichern: %s",
	"Starling is signed in to %s. Every agent given it in the Library uses this sign-in. You can close this tab.": "Starling ist bei %s angemeldet. Jeder Agent, der es in der Library bekommt, nutzt diese Anmeldung. Du kannst diesen Tab schließen.",
}
