package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"
)

var (
	ErrSMTPNotConfigured = errors.New("SMTP not configured")
	ErrInvalidRecipient  = errors.New("invalid recipient email")
	ErrSendFailed        = errors.New("failed to send email")
)

// EmailConfig holds the configuration for email sending
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	UseTLS       bool
}

// ObitoNotificationData represents the data for an obito notification email
type ObitoNotificationData struct {
	HospitalNome  string
	Setor         string
	HoraObito     time.Time
	TempoRestante string
	OccurrenceID  string
	Prioridade    int
	DashboardURL  string
}

// InfrastructureAlertData represents the data for an infrastructure alert email
type InfrastructureAlertData struct {
	ServiceName    string
	Status         string
	PreviousStatus string
	Timestamp      time.Time
	Message        string
	DashboardURL   string
}

// EmailService handles sending emails
type EmailService struct {
	config *EmailConfig
}

// NewEmailService creates a new EmailService
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// IsConfigured returns true if SMTP is properly configured
func (s *EmailService) IsConfigured() bool {
	return s.config != nil &&
		s.config.SMTPHost != "" &&
		s.config.SMTPPort > 0 &&
		s.config.SMTPFrom != ""
}

// SendObitoNotification sends an email notification for a new eligible obito
func (s *EmailService) SendObitoNotification(ctx context.Context, to string, data *ObitoNotificationData) error {
	if !s.IsConfigured() {
		return ErrSMTPNotConfigured
	}

	if to == "" || !strings.Contains(to, "@") {
		return ErrInvalidRecipient
	}

	subject := fmt.Sprintf("[URGENTE] Nova Ocorrencia Elegivel - %s", data.HospitalNome)
	body, err := s.renderObitoTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, to, subject, body)
}

// SendInfrastructureAlert sends an email alert for infrastructure issues
func (s *EmailService) SendInfrastructureAlert(ctx context.Context, to string, data *InfrastructureAlertData) error {
	if !s.IsConfigured() {
		return ErrSMTPNotConfigured
	}

	if to == "" || !strings.Contains(to, "@") {
		return ErrInvalidRecipient
	}

	// Set default dashboard URL
	if data.DashboardURL == "" {
		data.DashboardURL = "http://localhost:3000/dashboard/status"
	}

	subject := fmt.Sprintf("[ALERTA] %s - %s", data.ServiceName, data.Status)
	body, err := s.renderInfrastructureAlertTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, to, subject, body)
}

// renderObitoTemplate renders the HTML template for obito notification
func (s *EmailService) renderObitoTemplate(data *ObitoNotificationData) (string, error) {
	tmpl, err := template.New("obito_notification").Parse(obitoNotificationTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderInfrastructureAlertTemplate renders the HTML template for infrastructure alert
func (s *EmailService) renderInfrastructureAlertTemplate(data *InfrastructureAlertData) (string, error) {
	tmpl, err := template.New("infrastructure_alert").Parse(infrastructureAlertTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendEmail sends an email via SMTP
func (s *EmailService) sendEmail(ctx context.Context, to, subject, body string) error {
	headers := make(map[string]string)
	headers["From"] = s.config.SMTPFrom
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	var auth smtp.Auth
	if s.config.SMTPUser != "" && s.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	}

	// Use TLS if configured
	if s.config.UseTLS || s.config.SMTPPort == 465 {
		return s.sendEmailTLS(addr, auth, to, message.Bytes())
	}

	// Standard SMTP (with STARTTLS if supported)
	err := smtp.SendMail(addr, auth, s.config.SMTPFrom, []string{to}, message.Bytes())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return nil
}

// sendEmailTLS sends email using TLS connection
func (s *EmailService) sendEmailTLS(addr string, auth smtp.Auth, to string, message []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%w: authentication failed: %v", ErrSendFailed, err)
		}
	}

	if err := client.Mail(s.config.SMTPFrom); err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return client.Quit()
}

// FormatTimeRemaining formats the remaining time for display
func FormatTimeRemaining(expiresAt time.Time) string {
	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return "Expirado"
	}

	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dmin", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dmin", minutes)
}

// obitoNotificationTemplate is the HTML template for obito notification emails
const obitoNotificationTemplate = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SIDOT - Nova Ocorrencia</title>
</head>
<body style="font-family: 'Segoe UI', Arial, sans-serif; margin: 0; padding: 0; background-color: #f3f4f6;">
    <table width="100%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <!-- Header -->
        <tr>
            <td style="background-color: #0EA5E9; padding: 20px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 24px;">SIDOT</h1>
                <p style="color: #e0f2fe; margin: 5px 0 0 0; font-size: 14px;">Sistema de Notificacao de Obitos</p>
            </td>
        </tr>

        <!-- Alert Banner -->
        <tr>
            <td style="background-color: #EF4444; padding: 15px; text-align: center;">
                <span style="color: #ffffff; font-size: 18px; font-weight: bold;">
                    NOVA OCORRENCIA ELEGIVEL
                </span>
            </td>
        </tr>

        <!-- Content -->
        <tr>
            <td style="padding: 30px;">
                <h2 style="color: #1f2937; margin: 0 0 20px 0; font-size: 20px;">
                    Detalhes da Ocorrencia
                </h2>

                <table width="100%" cellpadding="10" cellspacing="0" style="background-color: #f9fafb; border-radius: 8px; margin-bottom: 20px;">
                    <tr>
                        <td style="border-bottom: 1px solid #e5e7eb; color: #6b7280; font-size: 14px;">
                            <strong>Hospital:</strong>
                        </td>
                        <td style="border-bottom: 1px solid #e5e7eb; color: #1f2937; font-size: 14px;">
                            {{.HospitalNome}}
                        </td>
                    </tr>
                    <tr>
                        <td style="border-bottom: 1px solid #e5e7eb; color: #6b7280; font-size: 14px;">
                            <strong>Setor:</strong>
                        </td>
                        <td style="border-bottom: 1px solid #e5e7eb; color: #1f2937; font-size: 14px;">
                            {{.Setor}}
                        </td>
                    </tr>
                    <tr>
                        <td style="border-bottom: 1px solid #e5e7eb; color: #6b7280; font-size: 14px;">
                            <strong>Hora do Obito:</strong>
                        </td>
                        <td style="border-bottom: 1px solid #e5e7eb; color: #1f2937; font-size: 14px;">
                            {{.HoraObito.Format "02/01/2006 15:04"}}
                        </td>
                    </tr>
                    <tr>
                        <td style="color: #6b7280; font-size: 14px;">
                            <strong>Tempo Restante:</strong>
                        </td>
                        <td style="color: #EF4444; font-size: 14px; font-weight: bold;">
                            {{.TempoRestante}}
                        </td>
                    </tr>
                </table>

                <!-- Priority Badge -->
                <div style="text-align: center; margin-bottom: 20px;">
                    <span style="background-color: {{if ge .Prioridade 80}}#EF4444{{else if ge .Prioridade 60}}#F59E0B{{else}}#10B981{{end}}; color: #ffffff; padding: 8px 16px; border-radius: 20px; font-size: 12px; font-weight: bold;">
                        PRIORIDADE: {{.Prioridade}}
                    </span>
                </div>

                <!-- CTA Button -->
                <div style="text-align: center;">
                    <a href="{{.DashboardURL}}" style="display: inline-block; background-color: #0EA5E9; color: #ffffff; text-decoration: none; padding: 14px 28px; border-radius: 8px; font-size: 16px; font-weight: bold;">
                        Acessar Dashboard
                    </a>
                </div>
            </td>
        </tr>

        <!-- Warning -->
        <tr>
            <td style="padding: 20px 30px; background-color: #fef3c7;">
                <p style="color: #92400e; font-size: 13px; margin: 0; text-align: center;">
                    <strong>ATENCAO:</strong> A janela de captacao de corneas e de 6 horas apos o obito.
                    Acao imediata e necessaria.
                </p>
            </td>
        </tr>

        <!-- Footer -->
        <tr>
            <td style="background-color: #f3f4f6; padding: 20px; text-align: center;">
                <p style="color: #6b7280; font-size: 12px; margin: 0;">
                    SIDOT - Sistema de Gestao de Doacao de Corneas
                </p>
                <p style="color: #9ca3af; font-size: 11px; margin: 10px 0 0 0;">
                    Esta e uma mensagem automatica. Por favor, nao responda diretamente.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>`

// infrastructureAlertTemplate is the HTML template for infrastructure alert emails
const infrastructureAlertTemplate = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SIDOT - Alerta de Infraestrutura</title>
</head>
<body style="font-family: 'Segoe UI', Arial, sans-serif; margin: 0; padding: 0; background-color: #f3f4f6;">
    <table width="100%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <!-- Header -->
        <tr>
            <td style="background-color: #1f2937; padding: 20px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 24px;">SIDOT</h1>
                <p style="color: #9ca3af; margin: 5px 0 0 0; font-size: 14px;">Alerta de Infraestrutura</p>
            </td>
        </tr>

        <!-- Alert Banner -->
        <tr>
            <td style="background-color: #DC2626; padding: 20px; text-align: center;">
                <span style="color: #ffffff; font-size: 24px; font-weight: bold;">
                    SERVICO FORA DO AR
                </span>
            </td>
        </tr>

        <!-- Content -->
        <tr>
            <td style="padding: 30px;">
                <h2 style="color: #1f2937; margin: 0 0 20px 0; font-size: 20px;">
                    Detalhes do Alerta
                </h2>

                <table width="100%" cellpadding="12" cellspacing="0" style="background-color: #fef2f2; border: 2px solid #fecaca; border-radius: 8px; margin-bottom: 20px;">
                    <tr>
                        <td style="border-bottom: 1px solid #fecaca; color: #6b7280; font-size: 14px; width: 40%;">
                            <strong>Servico:</strong>
                        </td>
                        <td style="border-bottom: 1px solid #fecaca; color: #991b1b; font-size: 14px; font-weight: bold;">
                            {{.ServiceName}}
                        </td>
                    </tr>
                    <tr>
                        <td style="border-bottom: 1px solid #fecaca; color: #6b7280; font-size: 14px;">
                            <strong>Status Atual:</strong>
                        </td>
                        <td style="border-bottom: 1px solid #fecaca; color: #DC2626; font-size: 14px; font-weight: bold;">
                            {{.Status}}
                        </td>
                    </tr>
                    <tr>
                        <td style="border-bottom: 1px solid #fecaca; color: #6b7280; font-size: 14px;">
                            <strong>Status Anterior:</strong>
                        </td>
                        <td style="border-bottom: 1px solid #fecaca; color: #1f2937; font-size: 14px;">
                            {{.PreviousStatus}}
                        </td>
                    </tr>
                    <tr>
                        <td style="color: #6b7280; font-size: 14px;">
                            <strong>Detectado em:</strong>
                        </td>
                        <td style="color: #1f2937; font-size: 14px;">
                            {{.Timestamp.Format "02/01/2006 15:04:05"}}
                        </td>
                    </tr>
                </table>

                {{if .Message}}
                <div style="background-color: #fef3c7; border: 1px solid #fcd34d; border-radius: 8px; padding: 15px; margin-bottom: 20px;">
                    <p style="color: #92400e; font-size: 14px; margin: 0;">
                        <strong>Mensagem:</strong> {{.Message}}
                    </p>
                </div>
                {{end}}

                <!-- Action Required -->
                <div style="background-color: #fee2e2; border-radius: 8px; padding: 20px; margin-bottom: 20px; text-align: center;">
                    <p style="color: #991b1b; font-size: 16px; font-weight: bold; margin: 0 0 10px 0;">
                        ACAO IMEDIATA NECESSARIA
                    </p>
                    <p style="color: #7f1d1d; font-size: 14px; margin: 0;">
                        Verifique o servico e tome as medidas corretivas necessarias para restaurar a operacao normal.
                    </p>
                </div>

                <!-- CTA Button -->
                <div style="text-align: center;">
                    <a href="{{.DashboardURL}}" style="display: inline-block; background-color: #1f2937; color: #ffffff; text-decoration: none; padding: 14px 28px; border-radius: 8px; font-size: 16px; font-weight: bold;">
                        Ver Status do Sistema
                    </a>
                </div>
            </td>
        </tr>

        <!-- Recommended Actions -->
        <tr>
            <td style="padding: 0 30px 30px 30px;">
                <h3 style="color: #1f2937; margin: 0 0 15px 0; font-size: 16px;">
                    Acoes Recomendadas:
                </h3>
                <ol style="color: #4b5563; font-size: 14px; margin: 0; padding-left: 20px;">
                    <li style="margin-bottom: 8px;">Verificar logs do servico no servidor</li>
                    <li style="margin-bottom: 8px;">Confirmar conectividade de rede e banco de dados</li>
                    <li style="margin-bottom: 8px;">Reiniciar o servico se necessario</li>
                    <li style="margin-bottom: 8px;">Monitorar recuperacao no dashboard de status</li>
                </ol>
            </td>
        </tr>

        <!-- Footer -->
        <tr>
            <td style="background-color: #1f2937; padding: 20px; text-align: center;">
                <p style="color: #9ca3af; font-size: 12px; margin: 0;">
                    SIDOT - Sistema de Gestao de Doacao de Corneas
                </p>
                <p style="color: #6b7280; font-size: 11px; margin: 10px 0 0 0;">
                    Alerta automatico de monitoramento de infraestrutura
                </p>
            </td>
        </tr>
    </table>
</body>
</html>`
