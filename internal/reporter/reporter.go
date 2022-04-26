package reporter

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/yohamta/jobctl/internal/config"
	"github.com/yohamta/jobctl/internal/mail"
	"github.com/yohamta/jobctl/internal/models"
	"github.com/yohamta/jobctl/internal/scheduler"
)

type Reporter struct {
	*Config
}

type Config struct {
	Mailer mail.Mailer
}

func New(config *Config) *Reporter {
	return &Reporter{
		Config: config,
	}
}

func (rp *Reporter) ReportStep(cfg *config.Config, status *models.Status, node *scheduler.Node) error {
	st := node.ReadStatus()
	if st != scheduler.NodeStatus_None {
		log.Printf("%s %s", node.Name, status.StatusText)
	}
	if st == scheduler.NodeStatus_Error && node.MailOnError {
		return rp.sendError(cfg, status)
	}
	return nil
}

func (rp *Reporter) ReportSummary(status *models.Status, err error) {
	var buf bytes.Buffer
	buf.Write([]byte("\n"))
	buf.Write([]byte("Summary ->\n"))
	buf.Write([]byte(renderSummary(status, err)))
	buf.Write([]byte("\n"))
	buf.Write([]byte("Details ->\n"))
	buf.Write([]byte(renderTable(status.Nodes)))
	log.Printf(buf.String())
}

func (rp *Reporter) ReportMail(cfg *config.Config, status *models.Status) error {
	if (status.Status != scheduler.SchedulerStatus_Error &&
		status.Status != scheduler.SchedulerStatus_Cancel) && cfg.MailOn.Failure {
		return rp.sendError(cfg, status)
	} else if cfg.MailOn.Success {
		return rp.sendInfo(cfg, status)
	}
	return nil
}

func (rp *Reporter) sendInfo(cfg *config.Config, status *models.Status) error {
	mailConfig := cfg.InfoMail
	jobName := cfg.Name
	subject := fmt.Sprintf("%s %s (%s)", mailConfig.Prefix, jobName, status)
	body := renderHTML(status.Nodes)

	return rp.Mailer.SendMail(
		cfg.InfoMail.From,
		[]string{cfg.InfoMail.To},
		subject,
		body,
	)
}

func (rp *Reporter) sendError(cfg *config.Config, status *models.Status) error {
	mailConfig := cfg.ErrorMail
	jobName := cfg.Name
	subject := fmt.Sprintf("%s %s (%s)", mailConfig.Prefix, jobName, status)
	body := renderHTML(status.Nodes)

	return rp.Mailer.SendMail(
		cfg.ErrorMail.From,
		[]string{cfg.ErrorMail.To},
		subject,
		body,
	)
}

func renderSummary(status *models.Status, err error) string {
	t := table.NewWriter()
	var errText = ""
	if err != nil {
		errText = err.Error()
	}
	t.AppendHeader(table.Row{"Name", "Started At", "Finished At", "Status", "Params", "Error"})
	t.AppendRow(table.Row{
		status.Name,
		status.StartedAt,
		status.FinishedAt,
		status.Status,
		status.Params,
		errText,
	})
	return t.Render()
}

func renderTable(nodes []*models.Node) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "Step", "Started At", "Finished At", "Status", "Command", "Error"})
	for i, n := range nodes {
		var command = n.Command
		if n.Args != nil {
			command = strings.Join([]string{n.Command, strings.Join(n.Args, " ")}, " ")
		}
		t.AppendRow(table.Row{
			fmt.Sprintf("%d", i+1),
			n.Name,
			n.StartedAt,
			n.FinishedAt,
			n.StatusText,
			command,
			n.Error,
		})
	}
	return t.Render()
}

func renderHTML(nodes []*models.Node) string {
	var buffer bytes.Buffer
	addValFunc := func(val string) {
		buffer.WriteString(
			fmt.Sprintf("<td align=\"center\" style=\"padding: 10px;\">%s</td>",
				val))
	}
	buffer.WriteString(`
	<table border="1" style="border-collapse: collapse;">
		<thead>
			<tr>
				<th align="center" style="padding: 10px;">Name</th>
				<th align="center" style="padding: 10px;">Started At</th>
				<th align="center" style="padding: 10px;">Finished At</th>
				<th align="center" style="padding: 10px;">Status</th>
				<th align="center" style="padding: 10px;">Error</th>
			</tr>
		</thead>
		<tbody>
	`)
	addStatusFunc := func(status scheduler.NodeStatus) {
		style := ""
		switch status {
		case scheduler.NodeStatus_Error:
			style = "color: #D01117;font-weight:bold;"
		}
		buffer.WriteString(
			fmt.Sprintf("<td align=\"center\" style=\"padding: 10px; %s\">%s</td>",
				style, status))
	}
	for _, n := range nodes {
		buffer.WriteString("<tr>")
		addValFunc(n.Name)
		addValFunc(n.StartedAt)
		addValFunc(n.FinishedAt)
		addStatusFunc(n.Status)
		addValFunc(n.Error)
		buffer.WriteString("</tr>")
	}
	buffer.WriteString("</table>")
	return buffer.String()
}
