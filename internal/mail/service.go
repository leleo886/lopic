package mail

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/jordan-wright/email"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
)

//go:embed templates/*.html
var templatesFS embed.FS

type MailService struct {
	enabled      bool
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	from         string
	fromName     string
}

func NewMailService(cfg *models.MailConfig) *MailService {
	return &MailService{
		enabled:      cfg.Enabled,
		smtpHost:     cfg.SMTP.Host,
		smtpPort:     cfg.SMTP.Port,
		smtpUsername: cfg.SMTP.Username,
		smtpPassword: cfg.SMTP.Password,
		from:         cfg.SMTP.From,
		fromName:     cfg.SMTP.FromName,
	}
}

// UpdateConfig 更新邮件配置
func (s *MailService) UpdateConfig(cfg *models.MailConfig) {
	s.enabled = cfg.Enabled
	s.smtpHost = cfg.SMTP.Host
	s.smtpPort = cfg.SMTP.Port
	s.smtpUsername = cfg.SMTP.Username
	s.smtpPassword = cfg.SMTP.Password
	s.from = cfg.SMTP.From
	s.fromName = cfg.SMTP.FromName
}

func (s *MailService) GetSMTPPWD() string {
	return s.smtpPassword
}

func (s *MailService) IsEnabled() bool {
	return s.enabled
}

func (s *MailService) SendResetPasswordCode(to, code, locale string) error {
	data := struct {
		Code string
		To   string
	}{
		Code: code,
		To:   to,
	}

	if locale == "zh" {
		return s.sendTemplate("reset_password_zh.html", "重置密码", to, data)
	}

	return s.sendTemplate("reset_password.html", "Reset Password", to, data)
}

func (s *MailService) SendEmailVerification(to, username, verifyLink, locale string) error {
	data := struct {
		Username   string
		VerifyLink string
		To         string
	}{
		Username:   username,
		VerifyLink: verifyLink,
		To:         to,
	}

	if locale == "zh" {
		return s.sendTemplate("email_verification_zh.html", "邮箱验证", to, data)
	}

	return s.sendTemplate("email_verification.html", "Email Verification", to, data)
}

func (s *MailService) sendTemplate(templateName, subject, to string, data interface{}) error {
	tpl, err := template.ParseFS(templatesFS, "templates/"+templateName)
	if err != nil {
		return cerrors.ErrInternalServer
	}

	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return cerrors.ErrInternalServer
	}

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.fromName, s.from)
	e.To = []string{to}
	e.Subject = subject
	e.HTML = body.Bytes()

	smtpAuth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	smtpAddr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	err = e.Send(smtpAddr, smtpAuth)
	if err != nil {
		log.Errorf("failed to send email: to=%s, error=%v", to, err)
		return cerrors.ErrFailedToSendEmail
	}

	log.Infof("email sent successfully: to=%s, subject=%s", to, subject)
	return nil
}
