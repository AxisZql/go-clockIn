package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-clockIn/config"
	"go-clockIn/internal/models"
	"go-clockIn/internal/tools"
	"log"
	"time"
)

func main() {
	log.Println(fmt.Sprintf("打卡任务启动......"))
	_ = config.GetConfig()
	cr := cron.New(cron.WithSeconds())
	// 此处是每天7:30开始打卡，其他时间以此类推
	cr.AddFunc("0 30 7 * * ?", func() {
		now := time.Now()
		log.Println(now.Format(time.RFC3339))
		users, err := models.GetUsers()
		if err != nil {
			log.Println(err)
			return
		}
		for _, u := range users {
			for i := 0; i < 3; i++ {
				if _, err := tools.CollyHealthClockIn(u.Username, u.Password); err != nil {
					// 每个人打卡重试3次
					log.Println(fmt.Sprintf("打卡失败:[%v],err=%v", u.Username, err))
					if i < 2 {
						continue
					}
					tools.SentMsgByEmail(fmt.Sprintf("打卡失败！😭<br>失败详情：<br> %v", err), u.Email)
				} else {
					log.Println(fmt.Sprintf("打卡成功:[%v]", u.Username))
					tools.SentMsgByEmail(fmt.Sprintf("打卡成功！🥳<br> Have a nice day!"), u.Email)
					break
				}
			}
		}
		log.Println(fmt.Sprintf("总耗时:%v", time.Since(now)))
	})
	cr.Start()
	tick := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-tick.C:
			tick.Reset(20 * time.Second)
		}
	}
}
