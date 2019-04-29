package gitlab

import (
	"bytes"
	"log"
	"text/template"

	"github.com/xanzy/go-gitlab"
)

// Get the template for the reminder message.
func DefaultTemplate() *template.Template {
	// TODO: allow to load any template from a file
	const defaultTemplate = `
**[{{.MR.Title}}]({{.MR.WebURL}})**
{{if .Discussions}} {{.Discussions}} 💬 {{end}} {{range $key, $value := .Emojis}} {{$value}} :{{$key}}: {{end}} {{range .Missing}}{{.}} {{else}}You got all reviews, {{.Owner}}.{{end}}
`
	return template.Must(template.New("default").Parse(defaultTemplate))
}

// Exec the reminder message for the given merge request.
func execTemplate(template *template.Template, mr *gitlab.MergeRequest, owner string, contacts []string, discussions int, emojis map[string]int) string {
	data := struct {
		MR          *gitlab.MergeRequest
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
