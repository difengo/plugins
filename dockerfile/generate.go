package dockerfile

import (
	"net"
	"net/url"

	"goa.design/goa/codegen"
	goadesign "goa.design/goa/design"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginLast("dockerfile", "example", Example)
}

// Example generates a dockerfile for the example service using API information
func Example(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = append(files, GenerateDockerFile(genpkg, r))
		}
	}

	return files, nil
}

// GenerateDockerFile returns the generated docker file.
func GenerateDockerFile(genpkg string, root *httpdesign.RootExpr) *codegen.File {
	path := "Dockerfile"

	var server *goadesign.ServerExpr
	var ports, email string

	if len(root.Design.API.Servers) > 0 {
		server = root.Design.API.Servers[0]
	} else {
		server = root.Design.API.DefaultServer()
	}

	for _, host := range server.Hosts {
		for _, uri := range host.URIs {
			p := getPortFromURI(string(uri))

			if p != "" {

				if ports == "" {
					ports = ports + p
				} else {
					ports = ports + " " + p
				}
			}
		}
	}

	if root.Design.API.Contact != nil {
		email = root.Design.API.Contact.Email
	}

	data := map[string]interface{}{
		"ApiName":        root.Design.API.Name,
		"ApiDescription": root.Design.API.Description,
		"ApiVersion":     root.Design.API.Version,
		"ApiContact":     root.Design.API.Contact,
		"ServerName":     server.Name,
		"Company":        "wiserskills",
		"Ports":          ports,
		"ContactEmail":   email,
	}
	sections := []*codegen.SectionTemplate{
		&codegen.SectionTemplate{
			Name:   "dockerfile",
			Source: dockerfileT,
			Data:   data,
		},
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func getPortFromURI(uri string) string {

	u, err := url.Parse(uri)

	if err != nil {
		return ""
	}

	_, port, _ := net.SplitHostPort(u.Host)

	return port
}

const dockerfileT = `# Dockerfile for {{ .ApiName }} micro-service
FROM golang:1.10.3-alpine AS builder

RUN apk add --no-cache git && \
    go version

COPY . /go/src/{{ .Company }}/{{ .ApiName }}

WORKDIR /go/src/{{ .Company }}/{{ .ApiName }}

RUN go get github.com/golang/dep/cmd/dep && \
    dep init && \
    dep ensure -v

WORKDIR /go/src/{{ .Company }}/{{ .ApiName }}/cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./{{ .ServerName }} .

FROM scratch

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.name="{{ .ApiName }}"
LABEL org.label-schema.description="{{ .ApiDescription }}"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.vcs-ref=$VCS_REF
LABEL org.label-schema.vendor="{{ .Company }}"
LABEL org.label-schema.version="{{ .ApiVersion }}"
{{ if .ContactEmail }}
LABEL maintainer="{{ .ContactEmail }}"
{{ end }}

WORKDIR /root/

COPY --from=builder /go/src/{{ .Company }}/{{ .ApiName }}/cmd/{{ .ServerName }} .

EXPOSE {{ .Ports }}
ENTRYPOINT ["./{{ .ServerName }}"]
`
