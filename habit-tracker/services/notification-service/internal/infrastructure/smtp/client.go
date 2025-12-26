package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"path/filepath"

	"notification-service/internal/config"

	"gopkg.in/gomail.v2"
)

type Client struct {
	cfg       *config.SMTPConfig
	emailCfg  *config.EmailConfig
	templates map[string]*template.Template
}

// NewClient creates a new SMTP client
func NewClient(cfg *config.SMTPConfig, emailCfg *config.EmailConfig) (*Client, error) {
	client := &Client{
		cfg:       cfg,
		emailCfg:  emailCfg,
		templates: make(map[string]*template.Template),
	}

	if err := client.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return client, nil
}

// loadTemplates loads all email templates
func (c *Client) loadTemplates() error {
	verificationTemplate, err := template.ParseFiles(
		filepath.Join(c.emailCfg.TemplatesPath, "verification.html"),
	)
	if err != nil {
		verificationTemplate, err = template.New("verification").Parse(defaultVerificationTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default verification template: %w", err)
		}
	}
	c.templates["verification"] = verificationTemplate

	resetTemplate, err := template.ParseFiles(
		filepath.Join(c.emailCfg.TemplatesPath, "password_reset.html"),
	)
	if err != nil {
		resetTemplate, err = template.New("password_reset").Parse(defaultPasswordResetTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default password reset template: %w", err)
		}
	}
	c.templates["password_reset"] = resetTemplate

	changedTemplate, err := template.ParseFiles(
		filepath.Join(c.emailCfg.TemplatesPath, "password_changed.html"),
	)
	if err != nil {
		changedTemplate, err = template.New("password_changed").Parse(defaultPasswordChangedTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default password changed template: %w", err)
		}
	}
	c.templates["password_changed"] = changedTemplate

	return nil
}

// SendVerificationEmail sends an email verification email
func (c *Client) SendVerificationEmail(ctx context.Context, to, username, firstName, verificationToken string) error {
	verificationURL := fmt.Sprintf("%s?token=%s", c.emailCfg.VerificationURL, verificationToken)

	data := map[string]interface{}{
		"Username":        username,
		"FirstName":       firstName,
		"VerificationURL": verificationURL,
	}

	body, err := c.renderTemplate("verification", data)
	if err != nil {
		return fmt.Errorf("failed to render verification email: %w", err)
	}

	subject := "Verify Your Email - Habit Tracker"
	return c.send(to, subject, body)
}

// SendPasswordResetEmail sends a password reset email
func (c *Client) SendPasswordResetEmail(ctx context.Context, to, username, firstName, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", c.emailCfg.VerificationURL, resetToken)

	data := map[string]interface{}{
		"Username":  username,
		"FirstName": firstName,
		"ResetURL":  resetURL,
	}

	body, err := c.renderTemplate("password_reset", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset email: %w", err)
	}

	subject := "Reset Your Password - Habit Tracker"
	return c.send(to, subject, body)
}

// SendPasswordChangedEmail sends a password changed notification email
func (c *Client) SendPasswordChangedEmail(ctx context.Context, to string, wasReset bool) error {
	data := map[string]interface{}{
		"WasReset": wasReset,
	}

	body, err := c.renderTemplate("password_changed", data)
	if err != nil {
		return fmt.Errorf("failed to render password changed email: %w", err)
	}

	subject := "Password Changed - Habit Tracker"
	return c.send(to, subject, body)
}

// send sends an email using gomail
func (c *Client) send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", c.cfg.FromName, c.cfg.FromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(c.cfg.Host, c.cfg.Port, c.cfg.Username, c.cfg.Password)

	// TLS настройки
	// Если UseTLS = true, используем STARTTLS (порт 587)
	// Если UseTLS = false, используем SSL (порт 465)
	if c.cfg.UseTLS {
		d.SSL = false // STARTTLS для порта 587
		// Настройка TLS для Gmail и других современных SMTP серверов
		d.TLSConfig = &tls.Config{
			ServerName: c.cfg.Host,
		}
	} else {
		d.SSL = true // SSL для порта 465
		d.TLSConfig = &tls.Config{
			ServerName: c.cfg.Host,
		}
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// renderTemplate renders an email template with the provided data
func (c *Client) renderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, exists := c.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Default email templates
const defaultVerificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #4CAF50;">Welcome to Habit Tracker!</h2>
        <p>Hi {{.FirstName}},</p>
        <p>Thank you for signing up! Please verify your email address by clicking the button below:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationURL}}" style="background-color: #4CAF50; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Verify Email</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">{{.VerificationURL}}</p>
        <p>If you didn't create an account, please ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">This is an automated email, please do not reply.</p>
    </div>
</body>
</html>
`

const defaultPasswordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #FF5722;">Reset Your Password</h2>
        <p>Hi {{.FirstName}},</p>
        <p>We received a request to reset your password. Click the button below to reset it:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background-color: #FF5722; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">{{.ResetURL}}</p>
        <p>If you didn't request a password reset, please ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">This is an automated email, please do not reply.</p>
    </div>
</body>
</html>
`

const defaultPasswordChangedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Changed</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2196F3;">Password Changed Successfully</h2>
        <p>Hello,</p>
        <p>This email confirms that your password was {{if .WasReset}}reset{{else}}changed{{end}} successfully.</p>
        <p>If you did not make this change, please contact our support team immediately.</p>
        <div style="background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <p style="margin: 0;"><strong>Security Tip:</strong> For your security, all active sessions have been logged out. Please log in again with your new password.</p>
        </div>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">This is an automated email, please do not reply.</p>
    </div>
</body>
</html>
`
