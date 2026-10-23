package provider

// 星鸣不读别家的登录态，此用例随之移除。
//
// The case asserted that a switch away from an account saves it with the
// credentials it has now — the ones Claude Code refreshed a moment ago
// (its refresh token rotated, the old one dead) — not with the ones
// starling last saw. What told the fresh ones was Claude Code's own
// .credentials.json, which starling no longer reads (it reads the sign-in
// kept in a saved account's own config directory), so the premise no
// longer holds.
