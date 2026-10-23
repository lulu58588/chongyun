package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/lulu58588/starling/internal/provider"
)

// starlingUA is the User-Agent starling's requests go to a vendor with.
func starlingUA() string { return "starling/" + Version }

// userAgentFor is the User-Agent a request to p goes with, ua being the
// one the request came to starling with: the agent's own, as it sent it,
// where p is set to pass it on (provider.PassesUserAgent, #1851), and
// starling's where it isn't, where the agent sent none, and on starling's own
// requests — its classifier's, vision's, search's, the image tool's
// (starlingsOwn) — which no agent sent. A User-Agent the user set among the
// provider's headers is put on after, when the request is signed, and
// stands over either.
func userAgentFor(p provider.Provider, ua string) string {
	if ua != "" && p.PassesUserAgent() && !starlingsOwn(ua) {
		return ua
	}
	return starlingUA()
}

// starlingsOwn is a User-Agent starling's own requests carry: another starling's
// (starling/…), the classifier's, vision's and search's (starling-router/1,
// starling-vision/1, starling-search/1), and the image tool's, which names the
// agent it draws for before its own (claude-code/2.1.0 starling-image/1)
// from what the agent told it, not from a User-Agent the agent sent.
func starlingsOwn(ua string) bool {
	return strings.HasPrefix(ua, "starling/") || strings.HasPrefix(ua, "starling-") || strings.HasSuffix(ua, " "+DrawAgent)
}

// callerUA is the User-Agent the request ctx is for came with (withCaller),
// "" when there is none.
func callerUA(ctx context.Context) string {
	c, _ := ctx.Value(callerCtx{}).(caller)
	return c.ua
}

// stampUA puts on req, a request to p that isn't starling's own forwarded
// one (an embedding, a drawing, a video), the User-Agent forwardOnce's
// carry: starling's, or the agent's where p passes it on. A signed-in
// account's is left to its signing, as before, and another starling's is
// set where it is asked.
func stampUA(ctx context.Context, p provider.Provider, h http.Header) {
	if p.Account != nil || p.IsRemoteStarling() {
		return
	}
	h.Set("User-Agent", userAgentFor(p, callerUA(ctx)))
}
