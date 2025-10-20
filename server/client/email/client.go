package email

import (
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/wneessen/go-mail"
)

type EmailClient struct {
	from              string
	smtpClient        *mail.Client
	logoURL           string
	accountLockTmpl   *template.Template
	accountUnlockTmpl *template.Template
	accountSetupTmpl  *template.Template
	passwordResetTmpl *template.Template
}

func NewEmailClient(config *config.EmailSenderConfig, templateDir, logoUrl string) (*EmailClient, error) {

	tmpl, err := template.ParseGlob(filepath.Join(templateDir, "*.tmpl"))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing templates from directory: %s", templateDir)
		return nil, err
	}

	accountLockTmpl := tmpl.Lookup("account-lock.tmpl")
	accountSetupTmpl := tmpl.Lookup("account-setup.tmpl")
	accountUnlockedTmpl := tmpl.Lookup("account-unlocked.tmpl")
	passwordResetTmpl := tmpl.Lookup("password-reset.tmpl")
	if accountLockTmpl == nil || accountSetupTmpl == nil || accountUnlockedTmpl == nil ||
		passwordResetTmpl == nil {
		log.Logger().Error().Msg("Some of the email templates are empty")
		return nil, fmt.Errorf("some email templates are empty")
	}

	smtpClient, err := mail.NewClient(
		config.SmtpHost,
		mail.WithPort(int(config.SmtpPort)),
		mail.WithUsername(config.SmtpUser),
		mail.WithPassword(config.SmtpPassword),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occured when initializing smtp client")
		return nil, err
	}

	return &EmailClient{
		smtpClient:        smtpClient,
		logoURL:           logoUrl,
		from:              config.FromAddress,
		accountLockTmpl:   accountLockTmpl,
		accountUnlockTmpl: accountUnlockedTmpl,
		accountSetupTmpl:  accountSetupTmpl,
		passwordResetTmpl: passwordResetTmpl,
	}, nil
}

func (ec *EmailClient) sendEmail(to, subject string, selectedTmpl *template.Template,
	templateParams map[string]string) error {

	devConfig := config.GetDevelopmentConfig()
	if devConfig.Enable && devConfig.MockEmail {
		log.Logger().Info().Msgf("Mock Mail in Development Mode: Subject: %s, To: %s, Params: %v", subject, to, templateParams)
		return nil
	}

	msg := mail.NewMsg()
	err := msg.From(ec.from)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when setting `from` email: %s", ec.from)
		return err
	}

	err = msg.To(to)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when setting `to` email: %s", to)
		return err
	}

	msg.Subject(subject)
	msg.SetBodyHTMLTemplate(selectedTmpl, templateParams)

	return ec.smtpClient.DialAndSend(msg)
}

func (ec *EmailClient) SendAccountLockEmail(username, recipient, reason string) error {
	params := map[string]string{
		"LogoURL":  ec.logoURL,
		"Username": username,
		"Reason":   reason,
	}
	return ec.sendEmail(recipient, "Open Image Registry - Account Locked", ec.accountLockTmpl, params)
}

func (ec *EmailClient) SendAccountSetupEmail(username, recipient, passwordSetLink string) error {
	params := map[string]string{
		"LogoURL":         ec.logoURL,
		"Username":        username,
		"Email":           recipient,
		"SetPasswordLink": passwordSetLink,
	}
	return ec.sendEmail(recipient, "Open Image Registry - Account Setup", ec.accountSetupTmpl, params)
}

func (ec *EmailClient) SendAccountUnlockedEmail(username, recipient string) error {
	params := map[string]string{
		"LogoURL":  ec.logoURL,
		"Username": username,
	}
	return ec.sendEmail(recipient, "Open Image Registry - Your Account has been unlocked!", ec.accountUnlockTmpl, params)
}

func (ec *EmailClient) SendPasswordResetEmail(username, recipient, resetLink string) error {
	params := map[string]string{
		"LogoURL":   ec.logoURL,
		"Username":  username,
		"Email":     recipient,
		"ResetLink": resetLink,
	}
	return ec.sendEmail(recipient, "Open Image Registry - Reset your password", ec.passwordResetTmpl, params)
}