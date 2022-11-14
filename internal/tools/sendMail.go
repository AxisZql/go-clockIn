package tools

import (
	"fmt"
	"go-clockIn/config"
	"gopkg.in/gomail.v2"
)

func SentMsgByEmail(msg string, toUserEmail ...string) error {
	conf := config.GetConfig()
	mailTo := make([]string, 0) //收件人列表
	mailTo = append(mailTo, toUserEmail...)
	title := `健康打卡`
	body := fmt.Sprintf("Hi👋,健康打卡结果如下：<br> %s", msg)
	m := gomail.NewMessage()
	m.SetHeader(`From`, conf.Mail.Username)
	m.SetHeader(`To`, mailTo...)
	m.SetHeader(`Subject`, title)
	m.SetBody(`text/html`, body)
	err := gomail.NewDialer(conf.Mail.Host, conf.Mail.Port, conf.Mail.Username, conf.Mail.Password).DialAndSend(m)
	if err != nil {
		return err
	}
	return nil
}
