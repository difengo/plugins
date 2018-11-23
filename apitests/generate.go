package apitests

import (
	"fmt"
	"path"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	httpcodegen "goa.design/goa/http/codegen"
	httpdesign "goa.design/goa/http/design"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPlugin("apitests", "example", nil, Example)
}

type payloadParam struct {
	Name  string
	Value interface{}
}

// Example generates test files for the api
func Example(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			for _, svc := range r.HTTPServices {
				files = append(files, GenerateTestFile(genpkg, r, svc))
			}
		}
	}

	return files, nil
}

// GenerateTestFile returns the generated test file.
func GenerateTestFile(genpkg string, root *httpdesign.RootExpr, svc *httpdesign.ServiceExpr) *codegen.File {

	pkg := codegen.Goify(svc.Name(), false)
	filePath := fmt.Sprintf("%s_test.go", pkg)
	apiPkg := strings.ToLower(codegen.Goify(root.Design.API.Name, false))
	svcdata := httpcodegen.HTTPServices.Get(svc.Name())
	pkgName := httpcodegen.HTTPServices.Get(svc.Name()).Service.PkgName

	specs := []*codegen.ImportSpec{
		{Path: "log"},
		{Path: "net/http"},
		{Path: "net/http/httptest"},
		{Path: "os"},
		{Path: "testing"},
		{Path: "github.com/gavv/httpexpect"},
		{Path: "goa.design/goa/http", Name: "goahttp"},
		{Path: path.Join(genpkg, "http", codegen.SnakeCase(svc.Name()), "server"), Name: pkgName + "svr"},
		{Path: path.Join(genpkg, codegen.SnakeCase(svc.Name())), Name: pkgName},
	}

	sections := []*codegen.SectionTemplate{codegen.Header("", pkg, specs)}

	data := map[string]interface{}{
		"APIPkg":  apiPkg,
		"Service": svcdata.Service,
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "apitests-handler",
		Source: handlerT,
		Data:   data,
	})

	for _, ept := range svc.HTTPEndpoints {
		for _, route := range ept.Routes {

			pathAttributes := []*payloadParam{}

			for _, attrName := range ept.MethodExpr.Payload.AllRequired() {
				attr := ept.MethodExpr.Payload.Find(attrName)
				for _, example := range attr.UserExamples {
					if strings.Contains(route.Path, fmt.Sprintf("{%s}", attrName)) {
						pathAttributes = append(pathAttributes, &payloadParam{Name: attrName, Value: example.Value})
					}
				}
			}

			data := map[string]interface{}{
				"APIPkg":      apiPkg,
				"Service":     svcdata.Service,
				"ServiceName": strings.Title(svcdata.Service.Name),
				"MethodName":  strings.Title(ept.MethodExpr.Name),
				"Endpoint":    ept,
				"Route":       route,
				"PathParams":  pathAttributes,
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "apitests-testfunc",
				Source: testfuncT,
				Data:   data,
			})
		}
	}

	return &codegen.File{Path: filePath, SectionTemplates: sections}
}

const handlerT = `
func getServerHandle() http.Handler {

	// create an instance of the server
	logger := log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
	svc := New{{ .Service.StructName }}(logger)
	endpoints := {{ .Service.PkgName }}.NewEndpoints(svc)
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	m := goahttp.NewMuxer()
	server := {{ .Service.PkgName }}svr.New(endpoints, m, dec, enc, nil)
	{{ .Service.PkgName }}svr.Mount(m, server)

	// returns the handler
	return m
}
`

const testfuncT = `
func Test{{ .ServiceName }}{{ .MethodName }}(t *testing.T) {

	// run server using httptest
	server := httptest.NewServer(getServerHandle())
	defer server.Close()

	// create httpexpect instance
	e := httpexpect.New(t, server.URL)

	// TODO: test endpoint
	/* 
		r := e.{{ .Route.Method }}("{{ .Route.Path }}").
		{{- range .PathParams }}
		WithPath("{{ .Name }}", {{ .Value }}).
		{{-  end }}
		Expect().
		Status(http.StatusOK).Body()

		r.Equal("3\n") 
	*/
}
`
