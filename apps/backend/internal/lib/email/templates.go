package email

type Template string

const (
	TemplateWelcome             Template = "welcome"
	TemplateDueDateReminder     Template = "due_date_reminder"
	TemplateOverdueNotification Template = "overdue_notification"
	TemplateWeeklyReport        Template = "weekly_report"
)
