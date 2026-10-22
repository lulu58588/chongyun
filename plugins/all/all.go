// Package all registers every built-in plugin. Only magpie's main imports
// it: core never imports a plugin, and a plugin never imports another.
package all

import (
	_ "github.com/lulu58588/chongyun/plugins/budget"
	_ "github.com/lulu58588/chongyun/plugins/gateway"
	_ "github.com/lulu58588/chongyun/plugins/groups"
	_ "github.com/lulu58588/chongyun/plugins/library"
	_ "github.com/lulu58588/chongyun/plugins/middleware"
	_ "github.com/lulu58588/chongyun/plugins/otel"
	_ "github.com/lulu58588/chongyun/plugins/redact"
	_ "github.com/lulu58588/chongyun/plugins/rotation"
	_ "github.com/lulu58588/chongyun/plugins/routing"
	_ "github.com/lulu58588/chongyun/plugins/search"
	_ "github.com/lulu58588/chongyun/plugins/sessions"
	_ "github.com/lulu58588/chongyun/plugins/tools"
	_ "github.com/lulu58588/chongyun/plugins/usage"
	_ "github.com/lulu58588/chongyun/plugins/warmups"
)
