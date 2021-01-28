package service

import (
	"github.com/go-gomail/gomail"
	"jd_seckill/conf"
	"strconv"
)

type Email struct {
	host string
	port string
	user string
	pass string
}

func NerEmail(conf *conf.Config) *Email {
	host := conf.Read("smtp", "email_host")
	port := conf.Read("smtp", "port")
	user := conf.Read("smtp", "email_user")
	pass := conf.Read("smtp", "email_pwd")
	return &Email{host: host, port: port, user: user, pass: pass}
}

func (this *Email) SendMail(mailTo []string, subject, body string) error {
	port, _ := strconv.Atoi(this.port)
	m := gomail.NewMessage()
	m.SetHeader("From", "<"+this.user+">")
	m.SetHeader("To", mailTo...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(this.host, port, this.user, this.pass)
	err := d.DialAndSend(m)
	return err
}
