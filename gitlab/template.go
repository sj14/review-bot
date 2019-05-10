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

// Exec the reminder message for the given merge request.
func execTemplate(template *template.Template, project gitlab.Project, reminders []reminder) string {
	data := struct {
		Project   gitlab.Project
		Reminders []reminder
	}{
		project,
		reminders,
	}

	buffer := bytes.NewBuffer([]byte{})

	if err := template.Execute(buffer, data); err != nil {
		log.Fatalf("failed executing template: %v", err)
	}

	return buffer.String()
}
