package github

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/google/go-github/v90/github"
)

// DefaultTemplate contains a project header and reminder messages.
func DefaultTemplate() *template.Template {
	const defaultTemplate = `
# [{{.Repository.Name}}]({{.Repository.HTMLURL}})

**How-To**: *Got reminded? Just normally review the given pull request.*

---

{{range .Reminders}}
**[{{.PR.Title}}]({{.PR.HTMLURL}})**
{{if .Discussions}} {{.Discussions}} 💬 {{end}} {{range $emoji, $count := .Emojis}} {{$count}} :{{$emoji}}: {{end}} {{range .Missing}}{{.}} {{else}}You got all reviews, {{.Owner}}.{{end}}
{{end}}
`
	return template.Must(template.New("default").Parse(defaultTemplate))
}

// Exec the reminder message for the given merge request.
func ExecTemplate(template *template.Template, repository *github.Repository, reminders []reminder) (string, error) {
	data := struct {
		Repository *github.Repository
		Reminders  []reminder
	}{
		repository,
		reminders,
	}
	buffer := bytes.NewBuffer([]byte{})

	if err := template.Execute(buffer, data); err != nil {
		return "", fmt.Errorf("failed executing template: %w", err)
	}

	return buffer.String(), nil
}
