package svc

import (
	"fmt"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Client struct {
	config *Config
}

type Email struct {
	To      []string
	Subject string
	Body    string
}

func NewClient(config *Config) *Client {
	return &Client{config: config}
}

func (c *Client) SendResetEmail(to, resetURL string) error {
	subject := "密码重置请求"
	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>密码重置请求</h2>
            <p>我们收到了您的密码重置请求。请点击下面的链接重置您的密码：</p>
            <a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">重置密码</a>
            <p>如果链接无法点击，请复制以下地址到浏览器中打开：</p>
            <p>%s</p>
            <p><strong>注意：</strong>此链接将在30分钟后失效。</p>
        </body>
        </html>
    `, resetURL, resetURL)

	return c.sendEmail([]string{to}, subject, body)
}

func (c *Client) sendEmail(to []string, subject, body string) error {
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		c.config.From, to[0], subject, body)

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	return smtp.SendMail(addr, auth, c.config.From, to, []byte(msg))
}
