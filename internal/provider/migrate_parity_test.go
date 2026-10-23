package provider

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/lulu58588/starling/internal/catalog"
	"github.com/lulu58588/starling/internal/plugin"
)

// A built-in moved onto its plugin offers the reasoning levels it did for
// a model its vendor lists none for (Command Code's, Grok's, WorkBuddy's):
// its maker's, as the built-in borrowed them; a plugin of its own gives
// what it says.
func TestMovedKeepsEfforts(t *testing.T) {
	movedPlugin(t, "grok", map[string]map[string]any{})
	os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o755)
	os.WriteFile(catalog.CachePath(), []byte(`{"xai":{"id":"xai","models":{"grok-4.7":{"id":"grok-4.7",
	  "reasoning_options":[{"type":"effort","values":["low","medium","high","xhigh"]}]}}}}`), 0o644)
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	id := "grok-4.7"
	if len(borrowedEfforts(id)) == 0 {
		t.Fatal("the catalog written gives grok-4.7 no efforts")
	}
	pp := plugin.Provider{ID: "grok", Spec: "opencode-grok-auth", Models: []plugin.Model{{ID: id, Reasoning: true}}}
	if got := pluginCatalog(pp)[0].Efforts; !slices.Equal(got, borrowedEfforts(id)) {
		t.Fatalf("moved: %s's efforts %v, want %v", id, got, borrowedEfforts(id))
	}
	pp.ID = "someplugin"
	if got := pluginCatalog(pp)[0].Efforts; len(got) != 0 {
		t.Fatalf("a plugin's own provider borrowed %v", got)
	}
}

// Cline's plugin borrows its models' makers' levels, as ClinePass by API
// key does: signed in through the plugin, ClinePass showed none.
func TestClinePluginBorrowsEfforts(t *testing.T) {
	claudeHome(t)
	os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o755)
	os.WriteFile(catalog.CachePath(), []byte(`{"xai":{"id":"xai","models":{"grok-4.7":{"id":"grok-4.7",
	  "reasoning_options":[{"type":"effort","values":["low","medium","high","xhigh"]}]}}}}`), 0o644)
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	id := "cline-pass/grok-4.7"
	want := effortsOf(catalog.Model{ID: id})
	if len(want) == 0 {
		t.Fatal("the catalog written gives grok-4.7 no efforts")
	}
	pp := plugin.Provider{ID: "cline", Spec: "@magpie-community/opencode-cline-auth", Models: []plugin.Model{{ID: id}}}
	if got := pluginCatalog(pp)[0].Efforts; !slices.Equal(got, want) {
		t.Fatalf("cline plugin: %s's efforts %v, want the API key's %v", id, got, want)
	}
}

// Moved, the agent's own sign-in is the plugin's: the agent signing in to
// another account meanwhile writes no row of the built-in's, which came
// back beside the one set aside as a second own account.
func TestMovedOwnRowNotWritten(t *testing.T) {
	movedPlugin(t, "grok", map[string]map[string]any{})
	sideLogins("grok", "new@x", func(savedLogin) bool { return true })
	for _, l := range readLogins() {
		if l.Agent == "grok" {
			t.Fatalf("a grok row written while moved: %+v", l)
		}
	}
}
