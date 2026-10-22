// Package all registers every built-in plugin. Only magpie's main imports
// it: core never imports a plugin, and a plugin never imports another.
package all

import (
	_ "github.com/lulu58588/magpie/plugins/budget"
	_ "github.com/lulu58588/magpie/plugins/gateway"
	_ "github.com/lulu58588/magpie/plugins/groups"
	_ "github.com/lulu58588/magpie/plugins/library"
	_ "github.com/lulu58588/magpie/plugins/middleware"
	_ "github.com/lulu58588/magpie/plugins/otel"
	_ "github.com/lulu58588/magpie/plugins/redact"
	_ "github.com/lulu58588/magpie/plugins/rotation"
	_ "github.com/lulu58588/magpie/plugins/routing"
	_ "github.com/lulu58588/magpie/plugins/search"
	_ "github.com/lulu58588/magpie/plugins/sessions"
	_ "github.com/lulu58588/magpie/plugins/tools"
	_ "github.com/lulu58588/magpie/plugins/usage"
	_ "github.com/lulu58588/magpie/plugins/warmups"
)
