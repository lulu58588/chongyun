package gateway

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/lulu58588/starling/internal/catalog"
	"github.com/lulu58588/starling/internal/provider"
)

// A Remote starling searches the web for the models it serves as it would
// for its own agents: by the model's provider, when that searches by
// itself, else with its own searcher. Its list says which of the two for
// each model (web_search: "native" or "starling"), and a web search offered
// to one of them goes to it as it was offered rather than being run here
// with this starling's searcher (Player 报告: remoteStarling/glm-5.3's
// searches went to another model). Only a model it searches by itself,
// natively, is one this starling may search with for its other models: a
// search starling runs for another model is marked (SearchingHeader), and
// the other starling runs no searcher of its own for it, so two starlings
// that each name the other never ask each other in turn.

// SearchingHeader marks a request a starling made to search for another
// model, to a remote starling, which then searches only by the model itself.
const SearchingHeader = "X-Starling-Searching"

// Web search as a Remote starling's list says it.
const (
	searchNative   = "native"
	searchStarling = "starling"
)

// remoteSearch is how a Remote starling says it searches the web for model,
// "" for another provider or a model it doesn't.
func remoteSearch(p provider.Provider, model string) string {
	if !p.IsRemoteStarling() {
		return ""
	}
	if i := slices.IndexFunc(p.Available(), func(m catalog.Model) bool { return m.ID == model }); i >= 0 {
		return p.Available()[i].WebSearch
	}
	return ""
}

// remoteSearchesAny is whether a Remote starling searches natively for any
// of its models, which makes it one this starling may search with.
func remoteSearchesAny(p provider.Provider) bool {
	return p.IsRemoteStarling() && slices.ContainsFunc(p.Available(), func(m catalog.Model) bool { return m.WebSearch == searchNative })
}

// searchableModel is a model of p that starling may search with: a Gemini
// on a Google sign-in, one a Remote starling searches natively for, one of
// searchModels or a plugin marks that searches by itself (partSearcher),
// any other
// provider's.
func searchableModel(p provider.Provider, model string) bool {
	switch {
	case googleAccount(p):
		return searchModel(model)
	case p.IsRemoteStarling():
		return remoteSearch(p, model) == searchNative
	case partSearcher(p):
		return searchesAlone(p, model)
	}
	return true
}

// webSearchOf is what this starling's list tells another starling of how it
// searches for e (remoteSearch): starling is whether it has a searcher.
func webSearchOf(e provider.Entry, starling bool) string {
	p := e.Provider
	if e.Group == "" {
		if p.IsRemoteStarling() && remoteSearch(p, e.Model) == searchNative || codeAssistSearches(p, provider.CodeAssist, e.Model) {
			return searchNative
		}
		for _, proto := range p.Speaks() {
			if searchesItself(p, proto) || modelSearches(p, proto, e.Model) {
				return searchNative
			}
		}
	}
	if starling || searchesOwn(p) {
		return searchStarling
	}
	return ""
}

// markSearching marks a request to a remote starling made for a search.
func markSearching(ctx context.Context, req *http.Request) {
	if searching(ctx) {
		req.Header.Set(SearchingHeader, "1")
	}
}

// searchingFrom is whether a request is a search another starling made.
func searchingFrom(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), "starling/") && r.Header.Get(SearchingHeader) != ""
}
