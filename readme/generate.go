package readme

import (
	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPlugin("readme", "example", nil, Example)
}

// Example generates test files for the api
func Example(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = append(files, GenerateReadmeFile(genpkg, r))
		}
	}

	return files, nil
}

// GenerateReadmeFile returns the generated readme file.
func GenerateReadmeFile(genpkg string, root *httpdesign.RootExpr) *codegen.File {

	filePath := "README.md"
	sections := []*codegen.SectionTemplate{}

	apiData := map[string]interface{}{
		"Api": root.Design.API,
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "readme-api",
		Source: apiT,
		Data:   apiData,
	})

	for _, svc := range root.HTTPServices {
		for _, mtd := range svc.ServiceExpr.Methods {
			for _, ept := range svc.HTTPEndpoints {
				for _, route := range ept.Routes {

					routeData := map[string]interface{}{
						"Method":   mtd,
						"Endpoint": ept,
						"Route":    route,
					}

					sections = append(sections, &codegen.SectionTemplate{
						Name:   "readme-endpoint",
						Source: endpointT,
						Data:   routeData,
					})
				}
			}
		}
	}

	return &codegen.File{Path: filePath, SectionTemplates: sections}
}

const apiT = `
# {{ .Api.Title }}{{- if .Api.Version }} (v{{ .Api.Version }}){{- end }}

{{ .Api.Description }}

The service exposes the following methods and endpoints:
`

const endpointT = `
| Method        | Verb          | Path         | Description      |
| ------------- |---------------|--------------|------------------|
| {{ .Method.Name }}           | {{ .Route.Method }}           | {{ .Route.Path }} | {{ .Method.Description }} |
`
