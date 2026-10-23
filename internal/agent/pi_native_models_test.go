package agent

import (
	"path/filepath"
	"slices"
	"testing"
)

type pickRow struct{ value, note, group string }

func pickRows(opts []Option, prefix string) []pickRow {
	var out []pickRow
	for _, o := range opts {
		if len(o.Value) > len(prefix) && o.Value[:len(prefix)] == prefix {
			out = append(out, pickRow{o.Value, o.Note, o.Group})
		}
	}
	return out
}

// 星鸣不读别家的登录态，此用例随之移除。

// omp shares the gap for the provider of its current model: its own
// catalog (@oh-my-pi/pi-catalog) lists openai-codex's models, without the
// ones that aren't chat models.
func TestOmpOffersItsOwnOpenAICodexModels(t *testing.T) {
	home := syncHome(t)
	modules := filepath.Join(home, ".bun", "install", "global", "node_modules")
	writeFile(t, filepath.Join(modules, "@oh-my-pi", "pi-catalog", "src", "models.json"), `{
	  "openai": {"gpt-6-astra": {"id":"gpt-6-astra","name":"GPT-6 Astra"}},
	  "openai-codex": {
	    "gpt-6-astra": {"id":"gpt-6-astra","name":"GPT-6 Astra"},
	    "gpt-5.5": {"id":"gpt-5.5","name":"GPT-5.5"},
	    "gpt-image-2": {"id":"gpt-image-2","name":"gpt-image-2","kind":"image"}
	  }
	}`)
	nodeModulesOf = func(bin string, pkgs []string) []string {
		if bin == "omp" && slices.Contains(pkgs, "@oh-my-pi/pi-coding-agent") {
			return []string{filepath.Join(modules, "@oh-my-pi/pi-coding-agent", "node_modules"), modules}
		}
		return nil
	}
	t.Cleanup(func() { nodeModulesOf = func(string, []string) []string { return nil } })
	got := pickRows(ompOwnOptions(filepath.Join(home, "none.yml"), "openai-codex/gpt-6-astra"), "openai-codex/")
	want := []pickRow{{"openai-codex/gpt-6-astra", "GPT-6 Astra", "OpenAI Codex"}, {"openai-codex/gpt-5.5", "GPT-5.5", "OpenAI Codex"}}
	if !slices.Equal(got, want) {
		t.Fatalf("openai-codex rows: %v", got)
	}
}
