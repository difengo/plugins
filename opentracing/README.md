# OpenTracing Plugin

The `OpenTracing` plugin is a [goa v2](https://github.com/goadesign/goa/tree/v2) plugin
that adds instrumentation to generated service methods using the [OpenTracing](https://github.com/opentracing/opentracing-go) package.

By default, the plugin is using a NoOp tracer. To have a working configuration, this plugin must be used in conjunction with an implemented tracer plugin.

## Enabling the Plugin

To enable it, import it in your design.go file using the blank identifier `_` as follows:

```go

package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"
import _ "goa.design/plugins/opentracing" # Enables the plugin

var _ = API("...
```

and generate as usual:

```bash
goa gen PACKAGE
goa example PACKAGE
```

where `PACKAGE` is the Go import path of the design package.