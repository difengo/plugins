package buildfile

import (
	"strings"

	"goa.design/goa/codegen"
	goadesign "goa.design/goa/design"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginLast("buildfile", "example", nil, Example)
}

// Example generates a dockerfile for the example service using API information
func Example(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = append(files, GenerateBuildFiles(genpkg, r)...)
		}
	}

	return files, nil
}

// GenerateBuildFiles returns the generated docker file.
func GenerateBuildFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {

	var server *goadesign.ServerExpr
	var company, email string

	if len(root.Design.API.Servers) > 0 {
		server = root.Design.API.Servers[0]
	} else {
		server = root.Design.API.DefaultServer()
	}

	if root.Design.API.Contact != nil {
		company = strings.ToLower(root.Design.API.Contact.Name)
		email = strings.ToLower(root.Design.API.Contact.Email)
	}

	data := map[string]interface{}{
		"ApiName":        root.Design.API.Name,
		"ApiDescription": root.Design.API.Description,
		"ApiVersion":     root.Design.API.Version,
		"ApiContact":     root.Design.API.Contact,
		"ServerName":     server.Name,
		"Company":        company,
		"ContactEmail":   email,
	}

	gitFileSections := []*codegen.SectionTemplate{
		&codegen.SectionTemplate{
			Name:   "gitignore",
			Source: gitignoreT,
			Data:   data,
		},
	}

	buildFileSections := []*codegen.SectionTemplate{
		&codegen.SectionTemplate{
			Name:   "buildfile",
			Source: buildfileT,
			Data:   data,
		},
	}

	files := []*codegen.File{}
	files = append(files, &codegen.File{Path: ".gitignore", SectionTemplates: gitFileSections})
	files = append(files, &codegen.File{Path: "Makefile", SectionTemplates: buildFileSections})

	return files
}

const gitignoreT = `# ignore build folder
build/
`

const buildfileT = `#! /usr/bin/make
#
# Makefile for {{ .ApiName }}
#
# Targets:
# - "server" builds the micro-service server
# - "client" builds the micro-service client
# - "docker" builds the micro-service docker image 

server:
	go build -a -o ./build/{{ .ServerName }} ./cmd/{{ .ServerName }}

client:
	go build -a -o ./build/{{ .ServerName }}-cli ./cmd/{{ .ServerName }}-cli

docker:
	docker build -t "{{ .Company }}/{{.ApiName}}:{{.ApiVersion}}" .
`
