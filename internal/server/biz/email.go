package biz

import (
	"context"
	"fmt"
	"net/smtp"
	
	"go.uber.org/fx"
	
	"github.com/looplj/axonhub/internal/log"
)

type EmailServiceParams struct {
	fx.In
	
	Config *EmailConfig
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailService(params EmailServiceParams) *EmailService {
	return &EmailService{
		config: params.Config,
	}
}

type EmailService struct {
	config *EmailConfig
}

// SendRechargeSuccess sends recharge success notification
func (s *EmailService) SendRechargeSuccess(ctx context.Context, to string, amount float64, balance float64) error {
	if s.config == nil || s.config.SMTPHost == "" {
		log.Info(ctx, "Email not configured, skipping notification")
		return nil
	}
	
	subject := "MuskAPI 充值成功通知"
	body := fmt.Sprintf(`
您好！

您的充值已成功完成。

充值金额: ¥%.2f
当前余额: ¥%.2f

感谢您的支持！

MuskAPI 团队
`, amount, balance)
	
	return s.sendEmail(to, subject, body)
}

// SendLowBalanceWarning sends low balance warning
func (s *EmailService) SendLowBalanceWarning(ctx context.Context, to string, balance float64) error {
	if s.config == nil || s.config.SMTPHost == "" {
		log.Info(ctx, "Email not configured, skipping notification")
		return nil
	}
	
	subject := "MuskAPI 余额不足提醒"
	body := fmt.Sprintf(`
您好！

您的账户余额已不足。

当前余额: ¥%.2f

请及时充值以保证服务正常使用。

MuskAPI 团队
`, balance)
	
	return s.sendEmail(to, subject, body)
}

// SendWelcomeEmail sends welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, to string, name string) error {
	if s.config == nil || s.config.SMTPHost == "" {
		log.Info(ctx, "Email not configured, skipping notification")
		return nil
	}
	
	subject := "欢迎加入 MuskAPI"
	body := fmt.Sprintf(`
您好 %s！

欢迎加入 MuskAPI！您现在可以使用我们的 AI 模型接入服务。

体验版用户可获得 ¥10 免费额度，立即开始使用吧！

MuskAPI 团队
`, name)
	
	return s.sendEmail(to, subject, body)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	if s.config == nil {
		return fmt.Errorf("email not configured")
	}
	
	// Build email message
	from := s.config.FromEmail
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	}
	
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", from, to, subject, body)
	
	// SMTP auth
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	
	// Send email
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(msg))
}
