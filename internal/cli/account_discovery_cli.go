package cli

import (
	"fmt"

	"github.com/lulu58588/magpie/internal/provider"
	"github.com/lulu58588/magpie/internal/settings"
)

// accountsDiscoveryCmd is `chongyun accounts discovery [on|off]` (#1535):
// whether magpie reads the agents' own sign-ins (their credential files,
// the Keychain) and serves them as subscriptions; Settings → Privacy's
// Subscription accounts. With no argument it says which is in effect.
func accountsDiscoveryCmd(args []string) error {
	if len(args) > 1 || len(args) == 1 && args[0] != "on" && args[0] != "off" {
		return fmt.Errorf("usage: chongyun accounts discovery [on|off]")
	}
	if len(args) == 1 {
		s := settings.Load()
		if args[0] == "on" {
			s.AccountDiscovery = true
			s.NoAccountDiscovery = false
		} else {
			s.AccountDiscovery = false
			s.NoAccountDiscovery = true
		}
		if err := settings.Save(s); err != nil {
			return err
		}
		provider.ForgetAccounts()
	}
	switch {
	case !settings.AccountDiscoveryOff():
		fmt.Println("account discovery is on: chongyun reads your agents' sign-ins and serves them as subscriptions")
	case settings.Load().NoAccountDiscovery:
		fmt.Println("account discovery is off: chongyun reads none of your agents' sign-ins and serves no subscription; API-key providers work as before")
	case settings.Load().AccountDiscovery:
		fmt.Println("account discovery is off for this run: MAGPIE_NO_ACCOUNTS is set (saved: on)")
	default:
		fmt.Println("account discovery is off: this build reads none of your agents' sign-ins unless you turn it on (chongyun accounts discovery on); API-key providers work as before")
	}
	return nil
}
