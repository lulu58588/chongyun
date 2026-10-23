// Package all registers every built-in plugin. Only starling's main imports
// it: core never imports a plugin, and a plugin never imports another.
package all

import (
	_ "github.com/lulu58588/starling/plugins/budget"
	_ "github.com/lulu58588/starling/plugins/gateway"
	_ "github.com/lulu58588/starling/plugins/groups"
	_ "github.com/lulu58588/starling/plugins/library"
	_ "github.com/lulu58588/starling/plugins/middleware"
	_ "github.com/lulu58588/starling/plugins/otel"
	_ "github.com/lulu58588/starling/plugins/redact"
	_ "github.com/lulu58588/starling/plugins/rotation"
	_ "github.com/lulu58588/starling/plugins/routing"
	_ "github.com/lulu58588/starling/plugins/search"
	_ "github.com/lulu58588/starling/plugins/sessions"
	_ "github.com/lulu58588/starling/plugins/tools"
	_ "github.com/lulu58588/starling/plugins/usage"
	_ "github.com/lulu58588/starling/plugins/warmups"
)
