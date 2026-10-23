package gui

import (
	"testing"

	"github.com/lulu58588/starling/internal/usage"
)

// ylorn 报告: with no cards in the menu bar the tooltip said only
// "starling". It says starling over today's total: every token, split into
// input, output and cache (NextCommit 报告), then the cost in
// Settings' currency ("+" when some calls had no price, none when none had
// one) and the requests.
func TestTrayIdleTip(t *testing.T) {
	for _, c := range []struct {
		t    usage.Totals
		want string
	}{
		{usage.Totals{Calls: 152, Input: 912_000, Output: 318_400, CacheRead: 4_900_000, CacheWrite: 200_000, Cost: 3.4, Unpriced: 2},
			"starling\nToday: 6.3M tokens (912K in · 318K out · 5.1M cached)\n≈$3.40+ · 152 requests"},
		{usage.Totals{Calls: 1, Input: 950, Output: 50},
			"starling\nToday: 1.0K tokens (950 in · 50 out)\n1 request"},
		{usage.Totals{}, "starling\nToday: no requests yet"},
	} {
		if got := trayIdleTip(c.t, "usd", 0); got != c.want {
			t.Errorf("trayIdleTip(%+v) =\n%q, want\n%q", c.t, got, c.want)
		}
	}
	// waybar reads the title as Pango markup: still safe after sniTip
	if got := sniTip(trayIdleTip(usage.Totals{Calls: 3, Input: 10, Cost: 2}, "usd", 0)); got != "starling\nToday: 10 tokens (10 in · 0 out)\n≈$2.00 · 3 requests" {
		t.Errorf("sniTip: %q", got)
	}
}
