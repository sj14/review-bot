package github

import (
	"bytes"
	"log"
	"text/template"

	"github.com/google/go-github/v25/github"
)

// DefaultTemplate contains a project header and reminder messages.
func DefaultTemplate() *template.Template {
	const defaultTemplate = `
# [{{.Repository.Name}}]({{.Repository.URL}})

**How-To**: *Got reminded? Just normally review the given pull request.*

---

{{range .Reminders}}
**[{{.PR.Title}}]({{.PR.HTMLURL}})**
{{if .Discussions}} {{.Discussions}} 💬 {{end}} {{range $emoji, $count := .Emojis}} {{$count}} :{{$emoji}}: {{end}} {{range .Missing}}{{.}} {{else}}You got all reviews, {{.PR.Owner}}.{{end}}
{{end}}
`
	return template.Must(template.New("default").Parse(defaultTemplate))
}

// Exec the reminder message for the given merge request.
func execTemplate(template *template.Template, repository *github.Repository, reminders []reminder) string {
	data := struct {
		Repository *github.Repository
		Reminders  []reminder
	}{
		repository,
		reminders,
	}
	buffer := bytes.NewBuffer([]byte{})

	if err := template.Execute(buffer, data); err != nil {
		log.Fatalf("failed executing template: %v", err)
	}

	return buffer.String()
}
