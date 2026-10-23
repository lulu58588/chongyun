package provider

import (
	"maps"
	"strings"

	"github.com/lulu58588/starling/internal/catalog"
)

// A remote starling is another computer's starling gateway, shared on its
// network (Settings → Share on local network), as a provider: the
// computers' agents stay wired by their own starling, while the providers,
// routing groups and usage are the one starling's they all share. Its
// address serves every API starling does, so the provider speaks them all
// — Chat Completions and Responses at <address>/v1, Anthropic's Messages
// at the address — and a request goes on in the API the agent spoke. Its
// model list tells, of each model, the APIs its own provider serves it on
// (native_endpoints), and a request for one goes on one of those: turned
// into another API once, here, when it must be, never on both computers.

// Its list also has the image models the other starling draws with, which
// it gives only to a starling that asks for them (DrawersHeader), each
// marked "kind": "image": an older starling would have taken one whose id
// doesn't say it draws for a model to chat with. They go on to the other
// starling's images API, which asks their vendor as it would its own.
// Each model's label is the other starling's name for it, with its provider
// there after it ("Claude Sonnet 5 · RelayA"), as its own agents see it:
// two of its providers' models of one name are told apart here too.
// Provider in model names Off here names each by its name there alone
// (display_name, catalog.Model's Plain), whatever that starling's own
// setting put after it (ARNO 报告).

// Its video models come the same way (VideomakersHeader), marked "kind":
// "video", and go on to its videos API (#545).

// Its list also tells what each model costs there (PricesHeader): the
// price the other starling counts its usage at — what its user set for the
// model by hand, else its provider's list price, at that provider's price
// rate — so its calls cost the same here (Sorghum 报告). It is this
// provider's list price for the model: a price set here for it still comes
// first, as one set for any provider's list price does. A starling tells it
// only to its own computer and to one sharing it with a gateway key, and
// one from before it tells nothing, which leaves the models priced as
// they were: by models.dev.

// Its decision models are requested with DecidersHeader, marked
// "kind": "decision", and asked at its System One API, including models
// whose names aren't Jev's.

// DrawersHeader asks a starling's model list for its image models as well.
const DrawersHeader = "X-Starling-Drawers"

// VideomakersHeader asks a starling's model list for its video models as
// well: a starling that doesn't send it would take grok-imagine-video, whose
// id says "imagine", for an image model.
const VideomakersHeader = "X-Starling-Videomakers"

// PricesHeader asks a starling's model list for what each model costs
// there ("starling_price"): a starling that doesn't send it is told nothing
// more than before.
const PricesHeader = "X-Starling-Prices"

// DecidersHeader asks a starling's model list for its decision models too.
const DecidersHeader = "X-Starling-Deciders"

// RemoteStarlingPreset is the preset's id.
const RemoteStarlingPreset = "remote-starling"

// IsRemoteStarling reports whether the provider is another starling.
func (p Provider) IsRemoteStarling() bool { return p.Preset == RemoteStarlingPreset }

// remoteStarlingEndpoints puts every API on the address given, however it
// was typed: the bare host and port (plain http, which starling serves),
// …/v1, or the Anthropic root.
func (p *Provider) remoteStarlingEndpoints() {
	if !p.IsRemoteStarling() {
		return
	}
	root := ""
	for _, u := range []string{p.Chat, p.Responses, p.Anthropic, p.Decide} {
		if u = strings.TrimRight(strings.TrimSpace(u), "/"); u != "" {
			root = u
			break
		}
	}
	if root == "" {
		return
	}
	if !strings.Contains(root, "://") {
		root = "http://" + root
	}
	for _, suf := range []string{
		"/v1/messages/count_tokens", "/v1/messages",
		"/v1/chat/completions", "/v1/responses", "/v1/systemone",
		"/v1/embeddings", "/v1/rerank",
		"/v1/audio/transcriptions", "/v1/audio/translations",
		"/v1/images/generations", "/v1/images/edits",
		"/v1/videos", "/v1",
	} {
		if b, ok := strings.CutSuffix(root, suf); ok {
			root = b
			break
		}
	}
	p.Chat, p.Responses, p.Anthropic = root+"/v1", root+"/v1", root
	p.Decide = p.Chat
}

// listHeaders are the headers p's model list is asked with: a remote
// starling is asked for its image, video and decision models too, and its prices.
func (p Provider) listHeaders() map[string]string {
	if !p.IsRemoteStarling() {
		return p.Headers
	}
	h := maps.Clone(p.Headers)
	if h == nil {
		h = map[string]string{}
	}
	h[DrawersHeader] = "1"
	h[VideomakersHeader] = "1"
	h[PricesHeader] = "1"
	h[DecidersHeader] = "1"
	return h
}

// listHeadersAt is listHeaders for a list asked at u: at OpenCode's
// gateway with a session of the list's own (OpenCodeSession), unless the
// user set one on the provider.
func (p Provider) listHeadersAt(u string) map[string]string {
	h := p.listHeaders()
	if !OpenCodeURL(u) {
		return h
	}
	for k := range h {
		if strings.EqualFold(k, "x-opencode-session") {
			return h
		}
	}
	h = maps.Clone(h)
	if h == nil {
		h = map[string]string{}
	}
	id := openCodeSession("")
	h["X-Opencode-Session"], h["X-Opencode-Session-Id"] = id, id
	return h
}

// remotePrice is what a remote starling's list said the model costs there,
// if it said.
func (p Provider) remotePrice(model string) (catalog.Price, bool) {
	if !p.IsRemoteStarling() {
		return catalog.Price{}, false
	}
	for _, m := range p.liveByID(model) {
		if m.Price != nil {
			return *m.Price, true
		}
	}
	return catalog.Price{}, false
}

// RemotePriced reports whether the model's list price is the one a remote
// starling's list gave, for the editor to say whose it is.
func (p Provider) RemotePriced(model string) bool {
	_, ok := p.remotePrice(model)
	return ok
}
