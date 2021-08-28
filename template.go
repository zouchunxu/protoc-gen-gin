package main

import (
	"bytes"
	"strings"
	"text/template"
)

var httpTemplate = `
type {{.ServiceType}}Handler interface {
{{range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{end}}
}

func New{{.ServiceType}}Handler(srv {{.ServiceType}}Handler, route *gin.Engine) {
	{{range .Methods}}
	route.Handle("{{.Method}}", "{{.Path}}", func(c *gin.Context) {
		var in {{.Request}}
		if "{{.Method}}" == "GET" {
			if err := c.ShouldBindQuery(&in); err != nil {
				c.JSON(200, responses.New(500, nil, err.Error()))
				return
			}
		} else {
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(200, responses.New(500, nil, err.Error()))
				return
			}
		}
	
		ctx := context.Background()
		out, err := srv.{{.Name}}(ctx, &in)
		if err != nil {
			if apiError, ok := err.(*api_errors.ApiError); ok {
				c.JSON(200, responses.New(apiError.Code(), apiError.Data(), apiError.Message()))
				return
			}
			c.JSON(200, err.Error())
			return
        }
		c.JSON(200, responses.New(0, out, "ok"))
	})
	{{end}}
}
`

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name    string
	Num     int
	Vars    []string
	Forms   []string
	Request string
	Reply   string
	// http_rule
	Path         string
	Method       string
	Body         string
	ResponseBody string
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}
