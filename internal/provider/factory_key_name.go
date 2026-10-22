package provider

// Factory's API keys, as they are told apart.

// factoryKeyPrefix is droid's Uh: what a Factory API key starts with.
const factoryKeyPrefix = "fk-"

// factoryKeyName is a Factory API key as a name: its first seven
// characters and its last four, the middle one an ellipsis.
func factoryKeyName(k string) string {
	if len(k) <= 10 {
		return k
	}
	return k[:7] + "…" + k[len(k)-4:]
}
