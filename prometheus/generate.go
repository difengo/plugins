package prometheus

import (
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/design"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPlugin("prometheus", "gen", AddMetricsEndpoint, UpdateServerFiles)
}

type fileToModify struct {
	file           *codegen.File
	path           string
	ServiceName    string
	isServer       bool
	isService      bool
	isEndpoints    bool
	isEncodeDecode bool
}

// AddMetricsEndpoint adds a new metrics endpoint to the roots during the preparation phase
func AddMetricsEndpoint(genpkg string, roots []eval.Root) error {

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			for _, svc := range r.HTTPServices {

				method := &design.MethodExpr{
					Name:             "metrics",
					Description:      "Prometheus metrics endpoint.",
					Service:          svc.ServiceExpr,
					Result:           &design.AttributeExpr{Type: design.String},
					Payload:          &design.AttributeExpr{Type: design.Empty},
					StreamingPayload: &design.AttributeExpr{Type: design.Empty},
				}

				svc.ServiceExpr.Methods = append(svc.ServiceExpr.Methods, method)

				route := &httpdesign.RouteExpr{Method: "GET", Path: "/metrics"}
				routes := make([]*httpdesign.RouteExpr, 1)
				routes[0] = route

				responses := []*httpdesign.HTTPResponseExpr{}
				responses = append(responses, &httpdesign.HTTPResponseExpr{
					StatusCode: httpdesign.StatusOK,
					Headers:    design.NewEmptyMappedAttributeExpr(),
					Body:       &design.AttributeExpr{Type: design.String},
				})
				endpoint := &httpdesign.EndpointExpr{
					MethodExpr:    method,
					Service:       svc,
					Routes:        routes,
					Params:        design.NewEmptyMappedAttributeExpr(),
					Headers:       design.NewEmptyMappedAttributeExpr(),
					Body:          &design.AttributeExpr{Type: design.Empty},
					StreamingBody: &design.AttributeExpr{Type: design.Empty},
					Responses:     responses,
				}

				responses[0].Parent = endpoint
				route.Endpoint = endpoint

				svc.HTTPEndpoints = append(svc.HTTPEndpoints, endpoint)
			}
		}
	}

	return nil
}

// UpdateServerFiles adds prometheus handler to generated server code
func UpdateServerFiles(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

	filesToModify := []*fileToModify{}

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			for _, svr := range r.Design.API.Servers {
				for _, svc := range svr.Services {
					serverPkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
					serverPath := filepath.Join("gen", "http", serverPkg, "server", "server.go")
					encodeDecodePath := filepath.Join("gen", "http", serverPkg, "server", "encode_decode.go")
					servicePath := filepath.Join("gen", serverPkg, "service.go")
					endpointsPath := filepath.Join("gen", serverPkg, "endpoints.go")
					filesToModify = append(filesToModify, &fileToModify{path: serverPath, ServiceName: svc, isServer: true})
					filesToModify = append(filesToModify, &fileToModify{path: servicePath, ServiceName: svc, isService: true})
					filesToModify = append(filesToModify, &fileToModify{path: endpointsPath, ServiceName: svc, isEndpoints: true})
					filesToModify = append(filesToModify, &fileToModify{path: encodeDecodePath, ServiceName: svc, isEncodeDecode: true})
				}
			}

			// Update the added files
			for _, fileToModify := range filesToModify {
				for _, file := range files {
					if file.Path == fileToModify.path {
						fileToModify.file = file
						updateFile(genpkg, r, fileToModify)
						break
					}
				}
			}
		}
	}

	return files, nil
}

func updateFile(genpkg string, root *httpdesign.RootExpr, f *fileToModify) {

	sections := f.file.SectionTemplates
	header := sections[0]

	if f.isServer {
		codegen.AddImport(header, &codegen.ImportSpec{
			Name: "promhttp",
			Path: "github.com/prometheus/client_golang/prometheus/promhttp",
		})

		for _, s := range f.file.Section("server-init") {
			s.Source = serverInitT
		}

		for _, s := range f.file.Section("server-handler-init") {
			s.Source = serverHandlerInitT
		}
	}

	if f.isService {
		for _, s := range f.file.Section("service") {
			newSource := strings.Replace(s.Source, serviceOldT, serviceNewT, 1)
			s.Source = newSource
		}
	}

	if f.isEncodeDecode {
		for _, s := range f.file.Section("response-encoder") {
			source := `{{- if eq .Method.Name "metrics" }}{{- else }}
			` + s.Source
			source = strings.Replace(source, `{{ define "response" -}}`, `{{- end }}
			{{ define "response" -}}`, 1)
			s.Source = source
		}
	}

	if f.isEndpoints {
		for _, s := range sections {

			if s.Name == "endpoints-struct" || s.Name == "endpoints-init" || s.Name == "endpoints-use" {

				data := *s.Data.(*service.EndpointsData)

				for i, m := range data.Methods {
					if m.Name == "metrics" {
						data.Methods = append(data.Methods[:i], data.Methods[i+1:]...)
					}
				}

				s.Data = data
			}

			if s.Name == "endpoint-method" {
				source := `{{- if eq .Name "metrics" }}{{- else }}` + s.Source + `{{- end }}`
				s.Source = source
			}
		}
	}

	f.file.SectionTemplates = sections
}

const serviceOldT = `
type Service interface {
{{- range .Methods }}
	{{ comment .Description }}
	{{- if .ViewedResult }}
		{{- if not .ViewedResult.ViewName }}
			{{ comment "The \"view\" return value must have one of the following views" }}
			{{- range .ViewedResult.Views }}
				{{- if .Description }}
					{{ printf "//	- %q: %s" .Name .Description }}
				{{- else }}
					{{ printf "//	- %q" .Name }}
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if .ServerStream }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}, {{ .ServerStream.Interface }}) (err error)
	{{- else }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) ({{ if .Result }}res {{ .ResultRef }}, {{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}{{ end }}err error)
	{{- end }}
{{- end }}
}
`

const serviceNewT = `
type Service interface {
	{{- range .Methods }}
	{{- if eq .Name "metrics" }}
	{{- else }}
		{{ comment .Description }}
		{{- if .ViewedResult }}
			{{- if not .ViewedResult.ViewName }}
				{{ comment "The \"view\" return value must have one of the following views" }}
				{{- range .ViewedResult.Views }}
					{{- if .Description }}
						{{ printf "//	- %q: %s" .Name .Description }}
					{{- else }}
						{{ printf "//	- %q" .Name }}
					{{- end }}
				{{- end }}
			{{- end }}
		{{- end }}
		{{- if .ServerStream }}
			{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}, {{ .ServerStream.Interface }}) (err error)
		{{- else }}
			{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) ({{ if .Result }}res {{ .ResultRef }}, {{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}{{ end }}err error)
		{{- end }}
	{{- end }}
	{{- end }}
}
`

// input: ServiceData
const serverInitT = `{{ printf "%s instantiates HTTP handlers for all the %s service endpoints." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(
	e *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	{{- if streamingEndpointExists . }}
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
	{{- end }}
	{{- range .Endpoints }}
		{{- if .MultipartRequestDecoder }}
	{{ .MultipartRequestDecoder.VarName }} {{ .MultipartRequestDecoder.FuncName }},
		{{- end }}
	{{- end }}
) *{{ .ServerStruct }} {
	return &{{ .ServerStruct }}{
		Mounts: []*{{ .MountPointStruct }}{
			{{- range $e := .Endpoints }}
				{{- range $e.Routes }}
			{"{{ $e.Method.VarName }}", "{{ .Verb }}", "{{ .Path }}"},
				{{- end }}
			{{- end }}
			{{- range .FileServers }}
				{{- $filepath := .FilePath }}
				{{- range .RequestPaths }}
			{"{{ $filepath }}", "GET", "{{ . }}"},
				{{- end }}
			{{- end }}
		},
		{{- range .Endpoints }}
		{{- if eq .Method.VarName "Metrics" }}
		{{ .Method.VarName }}: promhttp.Handler(),
		{{- else }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(e.{{ .Method.VarName }}, mux, {{ if .MultipartRequestDecoder }}{{ .MultipartRequestDecoder.InitName }}(mux, {{ .MultipartRequestDecoder.VarName }}){{ else }}dec{{ end }}, enc, eh{{ if .ServerStream }}, up, connConfigFn{{ end }}),
		{{- end }}
		{{- end }}
	}
}
`

// input: EndpointData
const serverHandlerInitT = `{{- if eq .Method.Name "metrics" }}{{- else }}
{{ printf "%s creates a HTTP handler which loads the HTTP request and calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	{{- if .ServerStream }}
	up goahttp.Upgrader,
	connConfigFn goahttp.ConnConfigureFunc,
	{{- end }}
) http.Handler {
	var (
		{{- if .ServerStream }}
			{{- if .Payload.Ref }}
		decodeRequest  = {{ .RequestDecoder }}(mux, dec)
			{{- end }}
		{{- else }}
			{{- if .Payload.Ref }}
		decodeRequest  = {{ .RequestDecoder }}(mux, dec)
			{{- end }}
		encodeResponse = {{ .ResponseEncoder }}(enc)
		{{- end }}
		encodeError    = {{ if .Errors }}{{ .ErrorEncoder }}{{ else }}goahttp.ErrorEncoder{{ end }}(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

	{{- if .Payload.Ref }}
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}
	{{- end }}

	{{ if .ServerStream }}
		v := &{{ .ServicePkgName }}.{{ .Method.ServerStream.EndpointStruct }}{
			Stream: &{{ .ServerStream.VarName }}{
				upgrader: up,
				connConfigFn: connConfigFn,
				w: w,
				r: r,
			},
		{{- if .Payload.Ref }}
			Payload: payload.({{ .Payload.Ref }}),
		{{- end }}
		}
		_, err = endpoint(ctx, v)
	{{- else }}
		res, err := endpoint(ctx, {{ if .Payload.Ref }}payload{{ else }}nil{{ end }})
	{{- end }}

		if err != nil {
			{{- if .ServerStream }}
			if _, ok := err.(websocket.HandshakeError); ok {
				return
			}
			{{- end }}
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
	{{- if not .ServerStream }}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	{{- end }}
	})
}
{{- end }}
`
