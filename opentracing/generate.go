package opentracing

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

const pluginName = "opentracing"

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPlugin(pluginName, "gen", Generate)
	codegen.RegisterPlugin(pluginName, "example", Example)
}

type fileToModify struct {
	file        *codegen.File
	path        string
	serviceName string
	isMain      bool
}

// Generate generates opentracing specific files.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = append(files, GenerateFiles(genpkg, r)...)
		}
	}
	return files, nil
}

// Example modifies the example generated files by referencing
// the opentracing middleware in the main file and adding tracing code
// in traced methods
func Example(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

	filesToModify := []*fileToModify{}

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {

			// Add the generated main files
			for _, svr := range r.Design.API.Servers {
				pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
				mainPath := filepath.Join("cmd", pkg, "main.go")
				filesToModify = append(filesToModify, &fileToModify{path: mainPath, serviceName: svr.Name, isMain: true})
			}

			// Add the generated service files
			for _, svc := range r.HTTPServices {
				servicePath := codegen.SnakeCase(svc.Name()) + ".go"
				filesToModify = append(filesToModify, &fileToModify{path: servicePath, serviceName: svc.Name(), isMain: false})
			}

			if len(r.Design.Schemes) > 0 {
				servicePath := "auth.go"
				filesToModify = append(filesToModify, &fileToModify{path: servicePath, isMain: false})
			}

			// Update the added files
			for _, fileToModify := range filesToModify {
				for _, file := range files {
					if file.Path == fileToModify.path {
						fileToModify.file = file
						updateExampleFile(genpkg, r, fileToModify)
						break
					}
				}
			}
		}
	}

	return files, nil
}

// GenerateFiles create opentracing support files
func GenerateFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, 1)
	fw[0] = GenerateMiddlewareFile(genpkg)
	return fw
}

// GenerateMiddlewareFile returns the generated opentracing middleware file.
func GenerateMiddlewareFile(genpkg string) *codegen.File {
	path := filepath.Join(codegen.Gendir, "tracing", "opentracing.go")
	title := fmt.Sprint("Opentracing Middleware")
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "tracing", []*codegen.ImportSpec{
			{Path: "net/http"},
			{Path: "github.com/opentracing/opentracing-go", Name: "opentracing"},
			{Path: "github.com/opentracing/opentracing-go/ext", Name: "ext"},
			{Path: "goa.design/goa/http/middleware"},
		}),
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "opentracing-middleware",
		Source: middlewareT,
	})

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func updateExampleFile(genpkg string, root *httpdesign.RootExpr, f *fileToModify) {

	path := filepath.Join(genpkg, "tracing")
	header := f.file.SectionTemplates[0]

	if f.isMain {

		codegen.AddImport(header, &codegen.ImportSpec{Path: path, Name: "tracing"})

		r := regexp.MustCompile(`\s+if\s\*dbg\s\{[^}]+\}[^}]+`)

		for _, s := range f.file.SectionTemplates {
			if r.MatchString(s.Source) {
				text := r.FindString(s.Source)
				newtext := text + "        handler = tracing.OpenTracing()(handler)"
				s.Source = strings.Replace(s.Source, text, newtext, 1)
			}
		}

	} else {

		codegen.AddImport(header, &codegen.ImportSpec{Path: "github.com/opentracing/opentracing-go", Name: "opentracing"})

		r1 := regexp.MustCompile(`\}\}\{\{\sif\snot\s\.Method\.ViewedResult\.ViewName\s\}\}view\sstring\,\s\{\{\send\s\}\}\{\{\send\s\}\}\s\{\{\send\s\}\}err\serror\)\s\{\s+\{\{\-\send\s\}\}`)
		r2 := regexp.MustCompile(`func\s\(s\s\*\{\{\s\$\.VarName\s\}\}Svc\)\s\{\{\s\.Type\s\}\}Auth\(ctx\scontext\.Context\,\s\{\{\sif\seq\s\.Type\s\"Basic\"\s\}\}user\,\spass\{\{\selse\sif\seq\s\.Type\s\"APIKey\"\s\}\}key\{\{\selse\s\}\}token\{\{\send\s\}\}\sstring\,\sscheme\s\*security\.\{\{\s\.Type\s\}\}Scheme\)\s\(context\.Context\,\serror\)\s{`)
		for _, s := range f.file.SectionTemplates {
			if r1.MatchString(s.Source) {
				s.Source = r1.ReplaceAllString(s.Source, methodT)
			}
			if r2.MatchString(s.Source) {
				s.Source = r2.ReplaceAllString(s.Source, authT)
			}
		}
	}
}

const methodT = `}}{{ if not .Method.ViewedResult.ViewName }}view string, {{ end }}{{ end }} {{ end }}err error) {
{{- end }}
span, _ := opentracing.StartSpanFromContext(ctx, "{{ .ServiceVarName }}.{{ .Method.VarName }}")
defer span.Finish()
`

const authT = `func (s *{{ $.VarName }}Svc) {{ .Type }}Auth(ctx context.Context, {{ if eq .Type "Basic" }}user, pass{{ else if eq .Type "APIKey" }}key{{ else }}token{{ end }} string, scheme *security.{{ .Type }}Scheme) (context.Context, error) {
span, _ := opentracing.StartSpanFromContext(ctx, "{{ $.VarName }}.{{ .Type }}Auth")
defer span.Finish()
`

const middlewareT = `
// OpenTracing returns a middleware that traces HTTP requests using the globally defined
// opentracing tracer
func OpenTracing() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			wireCtx, _ := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.Header))

			serverSpan := opentracing.StartSpan(r.URL.Path, ext.RPCServerOption(wireCtx))
			defer serverSpan.Finish()

			r = r.WithContext(opentracing.ContextWithSpan(r.Context(), serverSpan))

			rw := middleware.CaptureResponse(w)
			h.ServeHTTP(rw, r)
		})
	}
}
`
