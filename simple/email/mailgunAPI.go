package email

import (
	"context"
	"fmt"
	"github.com/mailgun/mailgun-go"
	"time"
)

const welcomeHTML = `<a href="support@simpleapes.com"">Click here for help</a>`

type Client struct {
	from string
	mg   mailgun.Mailgun
}

func (c *Client) Welcome(toName, toEmail string) error {
	message := c.mg.NewMessage(c.from, "welcome to SimpleApes", welcomeHTML, buildEmail(toName, toEmail))
	message.SetHtml(welcomeHTML)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, err := c.mg.Send(ctx, message)
	return err
}

func WithMailgun(domain, apiKey string) ClientConfig {
	return func(c *Client) {
		mg := mailgun.NewMailgun(domain, apiKey)
		c.mg = mg
	}

}

func WithSender(name, email string) ClientConfig {
	return func(c *Client) {
		c.from = buildEmail(name, email)
	}
}

type ClientConfig func(*Client)

func NewClient(opts ...ClientConfig) *Client {
	client := Client{
		from: "support@simpleapes.com",
	}
	for _, opt := range opts {
		opt(&client)
	}
	return &client
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
