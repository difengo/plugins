# API Tests Plugin

The `API Tests` plugin is a [goa v2](https://github.com/goadesign/goa/tree/v2) plugin that generates tests skeleton for a defined API. It uses [HTTPExpect](https://github.com/gavv/httpexpect) to handle request and assert results.

## Enabling the Plugin

To enable it, import it in your design.go file using the blank identifier `_` as follows:

```go

package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"
import _ "goa.design/plugins/apitests" # Enables the plugin

var _ = API("...
```

and generate as usual:

```bash
goa gen PACKAGE
goa example PACKAGE
```

where `PACKAGE` is the Go import path of the design package.