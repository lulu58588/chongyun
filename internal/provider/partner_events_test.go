package provider

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

)

func countsLine(cs []PartnerCount) string {
	var out []string
	for _, c := range cs {
		out = append(out, fmt.Sprintf("%s %s %d", c.ID, c.What, c.Count))
	}
	return strings.Join(out, ",")
}

// A partner listed now is counted by the day, each kind up to its cap; a
// provider of the user's own, or one no longer listed, never is. The days
// ended are what is sent, and once sent they are gone.
// This build counts nobody: there is nowhere the counts would go, so
// nothing is counted at all.
func TestCountPartner(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	resetPartners()
	t.Cleanup(resetPartners)
	var body atomic.Value
	body.Store(`{"partners":[{"id":"acme","name":"Acme","chat":"https://a.example/v1"}]}`)
	t.Setenv("STARLING_PARTNERS", partnerFeedServer(t, &body).URL)
	fetchPartnersNow()

	for range 25 {
		CountPartner(PartnerShown, "acme", "acme", "deepseek", "mine")
	}
	CountPartner(PartnerOpened, "acme")
	CountPartner("clicked", "acme")

	now := time.Now()
	if got := PartnerCounts(now); len(got) != 0 {
		t.Fatalf("today's counts are kept: %v", got)
	}
	if got := PartnerCounts(now.Add(24 * time.Hour)); len(got) != 0 {
		t.Fatalf("tomorrow's counts are kept: %v", got)
	}
}
