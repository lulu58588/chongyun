package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/lulu58588/starling/internal/agentenv"
)

func azureHome(t *testing.T) {
	t.Helper()
	isolate(t)
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("APPDATA", filepath.Join(h, "AppData", "Roaming"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	t.Setenv("PATH", h)
	for _, v := range agentenv.Vars {
		t.Setenv(v, "")
	}
}

// Whatever the portal or another app gave — the resource's endpoint, its
// /openai, the v1 base, a classic deployment's URL, a newer resource's
// host, or only its name — starling asks the resource's v1 API.
func TestAzureBase(t *testing.T) {
	for in, want := range map[string]string{
		"https://contoso.openai.azure.com":              "https://contoso.openai.azure.com/openai/v1",
		"https://contoso.openai.azure.com/":             "https://contoso.openai.azure.com/openai/v1",
		"https://contoso.openai.azure.com/openai":       "https://contoso.openai.azure.com/openai/v1",
		" https://Contoso.openai.azure.com/openai/v1/ ": "https://contoso.openai.azure.com/openai/v1",
		"https://contoso.openai.azure.com/openai/deployments/gpt-4o/chat/completions?api-version=2024-10-21": "https://contoso.openai.azure.com/openai/v1",
		"contoso.openai.azure.com":                       "https://contoso.openai.azure.com/openai/v1",
		"https://eastus2-x.cognitiveservices.azure.com/": "https://eastus2-x.cognitiveservices.azure.com/openai/v1",
		"https://proj.services.ai.azure.com/openai/v1":   "https://proj.services.ai.azure.com/openai/v1",
		"contoso": "https://contoso.openai.azure.com/openai/v1",
	} {
		if got, ok := AzureBase(in); !ok || got != want {
			t.Errorf("%q: %q %v, want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "https://api.openai.com/v1", "https://openai.azure.com", "https://apim.example.com/openai"} {
		if got, ok := AzureBase(in); ok {
			t.Errorf("%q taken for a resource: %q", in, got)
		}
	}
	if !(Provider{Chat: "https://x.openai.azure.com/openai/v1"}).IsAzure() || (Provider{Chat: "https://api.openai.com/v1"}).IsAzure() {
		t.Error("IsAzure by host")
	}
}

// The preset is filled in with the resource's endpoint alone: saved, both
// chat completions and Responses are on its v1 API. Without one it isn't
// saved. The key goes in api-key alone, never as a Bearer.
func TestAzurePresetSaved(t *testing.T) {
	azureHome(t)
	p, err := FromPreset(AzurePreset)
	if err != nil {
		t.Fatal(err)
	}
	pr := Preset(AzurePreset)
	if pr.Endpoint == "" || pr.EndpointHint == "" || p.Chat != "" || p.Responses != "" {
		t.Fatalf("preset: %+v %+v", pr, p)
	}
	p.Key = "k"
	if err := Save(p); err == nil || !strings.Contains(err.Error(), "endpoint") {
		t.Fatalf("saved without an endpoint: %v", err)
	}
	p.Chat, p.Responses = "https://contoso.openai.azure.com/", "https://contoso.openai.azure.com/"
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	q, err := Find(AzurePreset)
	if err != nil {
		t.Fatal(err)
	}
	if q.Chat != "https://contoso.openai.azure.com/openai/v1" || q.Responses != q.Chat || q.Anthropic != "" || !q.IsAzure() {
		t.Fatalf("saved: %+v", q)
	}
	for _, proto := range []Protocol{Chat, Responses} {
		if h := AuthHeaders(*q, proto); h["api-key"] != "k" || len(h) != 1 {
			t.Errorf("%s: %v", proto, h)
		}
	}
	// a deployment named for a GPT model on Responses, like OpenAI's own;
	// its test asks max_completion_tokens
	if q.Native("gpt-5-codex") != Responses || q.Native("my-deploy") != Chat {
		t.Errorf("native: %s %s", q.Native("gpt-5-codex"), q.Native("my-deploy"))
	}
	url, body := tinyBody(*q, Chat, "my-deploy")
	if url != "https://contoso.openai.azure.com/openai/v1/chat/completions" || !strings.Contains(body, `"max_completion_tokens":16`) {
		t.Errorf("test: %s %s", url, body)
	}
	if url, _ := tinyBody(*q, Responses, "gpt-5-codex"); url != "https://contoso.openai.azure.com/openai/v1/responses" {
		t.Errorf("test: %s", url)
	}
	// nothing offered from the catalog before the resource's list: a
	// model not deployed there would be turned away
	if ms := q.Available(); len(ms) != 0 {
		t.Errorf("available before a list: %d", len(ms))
	}
}

// The resource's deployments are its model ids: asked on the data plane
// with the key in api-key, no Authorization, and the v1 API's model list
// where that is refused.
func TestAzureModels(t *testing.T) {
	azureHome(t)
	var mu sync.Mutex
	var asked []string
	deployments := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		asked = append(asked, r.URL.RequestURI())
		mu.Unlock()
		if r.Header.Get("api-key") != "k" || r.Header.Get("Authorization") != "" || r.Header.Get("x-api-key") != "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.URL.Path == "/openai/deployments" && r.URL.Query().Get("api-version") == "2022-12-01" && deployments:
			w.Write([]byte(`{"data":[{"id":"team-gpt","model":"gpt-5","object":"deployment","status":"succeeded"},{"id":"gpt-5-codex","model":"gpt-5-codex","object":"deployment","status":"succeeded"}],"object":"list"}`))
		case r.URL.Path == "/openai/v1/models":
			w.Write([]byte(`{"object":"list","data":[{"id":"gpt-4.1","object":"model"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code":"404","message":"Resource not found"}}`))
		}
	}))
	defer srv.Close()

	p, _ := FromPreset(AzurePreset)
	p.Key = "k"
	p.Chat, p.Responses = srv.URL+"/openai/v1", srv.URL+"/openai/v1"
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
	q, _ := Find(AzurePreset)
	ms, err := q.Fetch(context.Background())
	if err != nil || len(ms) != 2 || ms[0].ID != "team-gpt" || ms[1].ID != "gpt-5-codex" {
		t.Fatalf("deployments: %v %+v", err, ms)
	}
	if asked[0] != "/openai/deployments?api-version=2022-12-01" {
		t.Errorf("asked %v", asked)
	}
	if got := q.Available(); len(got) != 2 {
		t.Errorf("available %+v", got)
	}

	deployments = false
	asked = nil
	ms, err = q.Fetch(context.Background())
	if err != nil || len(ms) != 1 || ms[0].ID != "gpt-4.1" {
		t.Fatalf("v1 models: %v %+v", err, ms)
	}
	if len(asked) != 2 || asked[1] != "/openai/v1/models" {
		t.Errorf("asked %v", asked)
	}

	q.Key = "wrong"
	if _, err := q.Fetch(context.Background()); err == nil || !strings.Contains(err.Error(), "deployments' names") {
		t.Errorf("refused: %v", err)
	}
}

