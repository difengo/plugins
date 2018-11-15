# JaegerTracer Plugin

The `JaegerTracer` plugin is a [goa v2](https://github.com/goadesign/goa/tree/v2) plugin
that adds initialization code in the generated example main file for a [Jaeger](https://https://github.com/jaegertracing/jaeger-client-go) tracer.

It works in combination with the `OpenTracing` plugin.  

## Enabling the Plugin

To enable it, import the `OpenTracing` plugin and the `JaegerTracer` plugin in your design.go file using the blank identifier `_` as follows:

```go

package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"
import _ "goa.design/plugins/opentracing" # Enables the opentracing plugin
import _ "goa.design/plugins/jaegertracer" # Enables the plugin

var _ = API("...
```

and generate as usual:

```bash
goa gen PACKAGE
goa example PACKAGE
```

where `PACKAGE` is the Go import path of the design package.