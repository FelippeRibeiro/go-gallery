package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type Config struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func loadConfig() Config {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	return Config{
		Host: os.Getenv("SMTP_HOST"),
		Port: port,
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("SMTP_FROM"),
	}
}

func buildMsg(from, to, subject, htmlBody string) []byte {
	encoded := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	lines := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + encoded,
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"",
		htmlBody,
	}
	return []byte(strings.Join(lines, "\r\n"))
}

func send(to, subject, htmlBody string) error {
	cfg := loadConfig()
	if cfg.Host == "" {
		return fmt.Errorf("SMTP_HOST não configurado")
	}

	msg := buildMsg(cfg.From, to, subject, htmlBody)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	addr := cfg.Host + ":" + cfg.Port

	// Porta 465 → SSL implícito (tls.Dial)
	if cfg.Port == "465" {
		return sendSSL(addr, cfg.Host, auth, cfg.From, to, msg)
	}

	// Porta 587 / 25 → STARTTLS via smtp.SendMail
	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg)
}

// sendSSL estabelece uma conexão TLS direta (porta 465) e envia o e-mail.
func sendSSL(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("tls.Dial: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp.NewClient: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return w.Close()
}


func EnviarConfirmacaoPedido(toEmail, albumTitulo string, downloadURL string) error {
	body := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px;color:#333">
  <h2 style="color:#7c3aed">Pedido confirmado!</h2>
  <p>Seu pedido de fotos do álbum <strong>%s</strong> foi confirmado.</p>
  <p>Clique no botão abaixo para acessar e baixar suas fotos originais:</p>
  <a href="%s" style="display:inline-block;background:#7c3aed;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:bold">
    Baixar fotos
  </a>
  <p style="margin-top:24px;font-size:12px;color:#888">Os links de download expiram em 24 horas.</p>
</body></html>`, albumTitulo, downloadURL)
	return send(toEmail, "Pedido confirmado: "+albumTitulo, body)
}

func EnviarConviteAlbum(toEmail, albumTitulo, token, capaURL string) error {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173"
	}
	link := fmt.Sprintf("%s/accept-invite?token=%s", appURL, token)

	capaHTML := ""
	if capaURL != "" {
		capaHTML = fmt.Sprintf(`<img src="%s" alt="Capa do álbum" style="width:100%%;max-height:300px;object-fit:cover;border-radius:8px;margin-bottom:16px">`, capaURL)
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px;color:#333">
  %s
  <h2 style="color:#7c3aed">Você foi convidado!</h2>
  <p>Você recebeu um convite para visualizar o álbum <strong>%s</strong>.</p>
  <p>Clique no botão abaixo para criar sua conta e acessar o álbum:</p>
  <a href="%s"
     style="display:inline-block;background:#7c3aed;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:bold">
    Acessar álbum
  </a>
  <p style="margin-top:24px;font-size:12px;color:#888">Este link expira em 7 dias.</p>
</body>
</html>`, capaHTML, albumTitulo, link)

	return send(toEmail, "Convite para o álbum: "+albumTitulo, body)
}
