package gitlab

import (
	"bytes"
	"fmt"
	"text/template"

	"gitlab.com/gitlab-org/api/client-go/v2"
)

// DefaultTemplate contains a project header and reminder messages.
func DefaultTemplate() *template.Template {
	const defaultTemplate = `
# ![]({{.Project.AvatarURL}} =40x) [{{.Project.Name}}]({{.Project.WebURL}})

**How-To**: *Got reminded? Just normally review the given merge request with 👍/👎 or use 😴 if you don't want to receive a reminder about this merge request.*

---

{{range .Reminders}}
**[{{.MR.Title}}]({{.MR.WebURL}})**
{{if .Discussions}} {{.Discussions}} 💬 {{end}} {{range $emoji, $count := .Emojis}} {{$count}} :{{$emoji}}: {{end}} {{range .Missing}}{{.}} {{else}}You got all reviews, {{.Owner}}.{{end}}
{{end}}
`
	return template.Must(template.New("default").Parse(defaultTemplate))
}

// ExecTemplate execs the reminder message for the given merge requests.
func ExecTemplate(template *template.Template, project gitlab.Project, reminders []reminder) (string, error) {
	data := struct {
		Project   gitlab.Project
		Reminders []reminder
	}{
		project,
		reminders,
	}

	buffer := bytes.NewBuffer([]byte{})

	if err := template.Execute(buffer, data); err != nil {
		return "", fmt.Errorf("failed executing template: %w", err)
	}

	return buffer.String(), nil
}
