package templates

import (
	"bytes"
	"log"
	"text/template"

	gitlablib "github.com/xanzy/go-gitlab"
)

// Get the template for the reminder message.
func Get() *template.Template {
	// TODO: allow to load any template from a file

	const defaultTemplate = `
**[{{.MR.Title}}]({{.MR.WebURL}})**
{{if .Discussions}} {{.Discussions}} 💬 {{end}} {{range $key, $value := .Emojis}} {{$value}} :{{$key}}: {{end}} {{range .Missing}}{{.}} {{else}}You got all reviews, {{.Owner}}.{{end}}
`
	return template.Must(template.New("default").Parse(defaultTemplate))
}

// Exec the reminder message for the given merge request.
func Exec(template *template.Template, mr *gitlablib.MergeRequest, owner string, contacts []string, discussions int, emojis map[string]int) string {
	data := struct {
		MR          *gitlablib.MergeRequest
		Missing     []string
		Discussions int
		Owner       string
		Emojis      map[string]int
	}{
		mr,
		contacts,
		discussions,
		owner,
		emojis,
	}

	var buf []byte
	buffer := bytes.NewBuffer(buf)

	if err := template.Execute(buffer, data); err != nil {
		log.Fatalf("failed executing template: %v", err)
	}

	return buffer.String()
}
