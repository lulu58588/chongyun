package gateway

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/lulu58588/starling/internal/codexsummary"
)

// A rollout written portable (sessions.PortCodexRollouts) holds, in place
// of starling's compaction item, the message the gateway sends for it: what
// reaches the model through starling is the same either way, for a starling
// model and for Codex's own, natively compacted or not.
func TestPortedCompactionReadsTheSame(t *testing.T) {
	enc := codexsummary.Encrypted("Progress: done.\nNext: tests.")
	msg, _ := codexsummary.Message(enc)
	body := func(item any) []byte {
		b, _ := json.Marshal(map[string]any{"model": "gpt-5.4", "input": []any{
			userMessage("fix the build"), item, userMessage("go on")}})
		return b
	}
	before := body(map[string]any{"type": "compaction", "id": "cmp_starling_1", "encrypted_content": enc})
	after := body(msg)
	for _, c := range []struct{ starlingModel, native bool }{{true, false}, {true, true}, {false, false}} {
		a, _ := codexInputAs(before, c.starlingModel, c.native)
		b, _ := codexInputAs(after, c.starlingModel, c.native)
		var x, y any
		json.Unmarshal(a, &x)
		json.Unmarshal(b, &y)
		if !reflect.DeepEqual(x, y) {
			t.Errorf("%+v: starling's item sent as\n%s\nthe ported message as\n%s", c, a, b)
		}
	}
}
