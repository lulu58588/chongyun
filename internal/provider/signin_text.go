package provider

// The words of the pages a vendor's browser sign-in ends on. They are the
// pages starling serves itself (plain HTML, no script, so no front-end
// translation), in the language the user set or the browser asks for.

import "github.com/lulu58588/starling/internal/i18ntext"

// signInText is a translator for the sign-in pages, in one language.
func signInText(lang string) func(string) string {
	return func(en string) string {
		return i18ntext.Pick(lang, en, signInTextZh, signInTextZhTW, signInTextJa, signInTextDe)
	}
}

// signInTextf is signInText for a sentence with something in it to fill in:
// the translation has to keep the placeholders, or the English is used.
func signInTextf(lang string) func(string, ...any) string {
	return func(en string, args ...any) string {
		return i18ntext.Sprintf(lang, en, args, signInTextZh, signInTextZhTW, signInTextJa, signInTextDe)
	}
}

// The text of the sign-in pages, in Chinese. The key is the English as
// written in the code, so a page that says something new falls back to
// English rather than to nothing.
var signInTextZh = map[string]string{
	"This sign-in is over":                             "这次登录已经结束",
	"Start it again from starling.":                    "请从 starling 重新发起登录。",
	"Sign-in didn't finish":                            "登录没有完成",
	"This link isn't from starling's sign-in":          "这个链接不是来自 starling 的登录",
	"This sign-in is already finishing":                "这次登录正在完成中",
	"starling shows the account when it's done.":       "完成后 starling 会显示这个账号。",
	"You're signed in":                                 "登录成功",
	"%s is added to starling. You can close this tab.": "%s 已添加到 starling，可以关闭这个标签页了。",
}

// The same, in Traditional Chinese.
var signInTextZhTW = map[string]string{
	"This sign-in is over":                             "這次登入已經結束",
	"Start it again from starling.":                    "請從 starling 重新發起登入。",
	"Sign-in didn't finish":                            "登入沒有完成",
	"This link isn't from starling's sign-in":          "這個連結不是來自 starling 的登入",
	"This sign-in is already finishing":                "這次登入正在完成中",
	"starling shows the account when it's done.":       "完成後 starling 會顯示這個帳號。",
	"You're signed in":                                 "登入成功",
	"%s is added to starling. You can close this tab.": "%s 已加入 starling，可以關閉這個分頁了。",
}

// The same, in Japanese.
var signInTextJa = map[string]string{
	"This sign-in is over":                             "このサインインは終了しました",
	"Start it again from starling.":                    "starling からもう一度サインインしてください。",
	"Sign-in didn't finish":                            "サインインが完了しませんでした",
	"This link isn't from starling's sign-in":          "このリンクは starling のサインインのものではありません",
	"This sign-in is already finishing":                "このサインインはすでに完了処理中です",
	"starling shows the account when it's done.":       "完了すると starling にアカウントが表示されます。",
	"You're signed in":                                 "サインインしました",
	"%s is added to starling. You can close this tab.": "%s が starling に追加されました。このタブは閉じて構いません。",
}

// The same, in German.
var signInTextDe = map[string]string{
	"This sign-in is over":                             "Diese Anmeldung ist vorbei",
	"Start it again from starling.":                    "Bitte starte sie erneut aus starling.",
	"Sign-in didn't finish":                            "Anmeldung nicht abgeschlossen",
	"This link isn't from starling's sign-in":          "Dieser Link stammt nicht aus der Anmeldung von starling",
	"This sign-in is already finishing":                "Diese Anmeldung wird gerade abgeschlossen",
	"starling shows the account when it's done.":       "starling zeigt das Konto an, wenn es fertig ist.",
	"You're signed in":                                 "Du bist angemeldet",
	"%s is added to starling. You can close this tab.": "%s wurde zu starling hinzugefügt. Du kannst diesen Tab schließen.",
}
