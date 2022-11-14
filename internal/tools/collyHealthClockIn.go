package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly"
	"go-clockIn/pkg/constant"
	"golang.org/x/net/html"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SerialResp struct {
	Errno    int      `json:"errno"`
	Ecode    string   `json:"ecode"`
	Error    string   `json:"error"`
	Entities []string `json:"entities"`
}

func CollyHealthClockIn(username, password string) (ok bool, err error) {
	var client = colly.NewCollector(colly.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36`))
	log.Println(fmt.Sprintf("开始打卡流程：[%s]......", username))
	var (
		snURL           string
		clockInData     map[string]interface{}
		serialNumberURL = make(chan string, 1)
		loginOK         = make(chan int, 1)
		completeFlag    = make(chan string, 1)
		stepId          string
		csrfToken       string
	)

	client.OnResponse(func(r *colly.Response) {
		if r.StatusCode == http.StatusOK {
			if r.Request.URL.String() == constant.LoginURL && r.Request.Method == "GET" {
				// 登陆部分
				req, e := getLoginReq(strings.NewReader(string(r.Body)), username, password)
				if e != nil {
					log.Println(fmt.Sprintf("login got err:%v", e))
					return
				}
				client.Post(constant.LoginURL, req)
			} else if r.Request.URL.String() == constant.IndexPageURL && r.Request.Method == "POST" {
				// 登陆成功
				loginOK <- 1
			} else if r.Request.URL.String() == constant.ClockInURL && r.Request.Method == "GET" {
				// 获取流水号
				doc, _ := html.Parse(strings.NewReader(string(r.Body)))
				p1 := htmlquery.FindOne(doc, `//meta[@itemscope='csrfToken']/@content`)
				client.Post(constant.PreCommitURL, map[string]string{
					"idc":       "XNYQSB",
					"release":   "",
					"csrfToken": htmlquery.SelectAttr(p1, "content"),
					"lang":      "zh",
				})
			} else if r.Request.URL.String() == constant.PreCommitURL && r.Request.Method == "POST" {
				var ser SerialResp
				json.Unmarshal(r.Body, &ser)
				if len(ser.Entities) != 0 {
					serialNumberURL <- ser.Entities[0]
				}
			} else if r.Request.URL.String() == snURL && r.Request.Method == "GET" {
				rg, _ := regexp.Compile("formStepId = (.*?);")
				data := rg.FindAll(r.Body, -1)
				tmp := strings.Split(string(data[0]), "= ")
				stepId = tmp[1][:len(tmp[1])-1]
				doc, _ := html.Parse(strings.NewReader(string(r.Body)))
				p1 := htmlquery.FindOne(doc, `//meta[@itemscope='csrfToken']/@content`)
				csrfToken = htmlquery.SelectAttr(p1, "content")
				client.Post(constant.RenderURL, map[string]string{
					"stepId":     stepId,
					"instanceId": "",
					"admin":      "false",
					"rand":       strconv.Itoa(rand.Intn(999)),
					"width":      "960",
					"lang":       "zh",
					"csrfToken":  csrfToken,
				})
			} else if r.Request.URL.String() == constant.RenderURL && r.Request.Method == "POST" {
				json.Unmarshal(r.Body, &clockInData)
				if data, ok := clockInData["entities"].([]interface{})[0].(map[string]interface{})["data"]; ok {
					clockInData = data.(map[string]interface{})
					clockInData["fieldCXXXsftjhb"] = "2"  //7天内是否前往疫情重点地区
					clockInData["fieldJKMsfwlm"] = "1"    // 是否绿码
					clockInData["fieldSTQKbrstzk1"] = "1" // 本人身体状况
					clockInData["fieldCNS"] = "true"      // 确认按钮
				}
				client.Visit(constant.BeforeClockInGet)
			} else if r.Request.URL.String() == constant.BeforeClockInGet && r.Request.Method == "GET" {
				client.Post(constant.BeforeClockInPost, map[string]string{
					"stepId":       stepId,
					"includingTop": "true",
					"csrfToken":    csrfToken,
					"lang":         "zh",
				})
			} else if r.Request.URL.String() == constant.BeforeClockInPost && r.Request.Method == "POST" {
				// 第一次提交打卡数据
				data, _ := json.Marshal(clockInData)
				first := map[string]string{
					"stepId":      stepId,
					"actionId":    "1",
					"formData":    string(data),
					"timestamp":   strconv.Itoa(int(time.Now().Unix())),
					"rand":        strconv.Itoa(rand.Intn(9999)),
					"boundFields": "fieldYMTGSzd,fieldSTQKjtcyzd,fieldSTQKfrtw,fieldCXXXjtfslc,fieldSTQKxgqksm,fieldJKMsfwlm,fieldJKHDDzt,fieldYQJLzhycjcsj,fieldSTQKfl,fieldSTQKhxkn,fieldCXXXsfylk,fieldSTQKglfs,fieldCXXXsfjcgyshqzbl,fieldSTQKjtcyfx,fieldCXXXszsqsfyyshqzbl,fieldJCDDs,fieldSTQKjtcyfs,fieldSTQKjtcyzljgmc,fieldSQSJ,fieldJBXXnj,fieldSTQKfx,fieldSTQKfs,fieldYQJLjcddshi,fieldHQRQ,fieldSTQKjtcyqtms,fieldCXXXksjcsj,fieldSTQKjtcyxm,fieldZJYCHSJCYXJGRQzd,fieldqjymsjtqk,fieldCXXXjcdr,fieldCXXXsftjhbjtdz,fieldJCDDq,fieldSFJZYM,fieldSTQKjtcyclfs,fieldSTQKxm,fieldSTQKjtcyzdjgmcc,fieldSTQKqt,fieldCXXXlksj,fieldJBXXfdy,fieldSTQKjtcyjmy,fieldCXXXsftjhbq,fieldSTQKqtms,fieldYCFDY,fieldCXXXjtfspc,fieldSTQKbrstzk1,fieldCXXXssh,fieldJBXXjgjdwbk,fieldLYYZM,fieldCNS,fieldJBXXjzdz,fieldSTQKclfs,fieldSTQKjtcyfl,fieldSTQKjtcyzdjgmc,fieldJBXXbj,fieldSTQKjtcyfxx,fieldJBXXcsny,fieldCXXXdqszd,fieldSTQKjtcystzk,fieldSTQKjtcypcsj,fieldJBXXqu,fieldJBXXjgshi,fieldYQJLjcddq,fieldYQJLjcdds,fieldCXXXjtzz,fieldCXXXjtfsqt,fieldJTCZDZqu,fieldDQSJ,fieldSTQKzdjgmc,fieldJTCZDZxxdz,fieldSTQKjtcyglkssj,fieldCXXXsftjhb,fieldJTCZDZJDcode,fieldzgzjzdzjtdz,fieldJCDDqmsjtdd,fieldSHENGYC,fieldYQJLksjcsj,fieldJBXXjgsjtdz,fieldSTQKbrstzk,fieldSTQKjtcyqt,fieldJBXXlxfs,fieldSTQKpcsj,fieldYQJLsfjcqtbl,fieldJTCZDZsheng,fieldJBXXbz,fieldFLid,fieldjgs,fieldJCDDshi,fieldSTQKrytsqkqsm,fieldzgzjzdzs,fieldzgzjzdzq,fieldJZDZC,fieldSTQKjtcyzdkssj,fieldYQJLjcdry,fieldCXXXjtfsdb,fieldCXXXcxzt,fieldCXXXjtjtzz,fieldCXXXsftjhbs,fieldSTQKzdkssj,fieldSTQKfxx,fieldJTCZDZJDwbk,fieldSTQKjtcyzysj,fieldjgshi,fieldJBXXsheng,fieldJBXXdrsfwc,fieldJBXXdw,fieldCXXXjtgjbc,fieldJBXXjgjdcode,fieldSTQKjtcygldd,fieldzgzjzdzshi,fieldSTQKzd,fieldSTQKjtcyfrsj,fieldCXXXjtfsqtms,fieldSTQKjtcyzdmc,fieldCXXXjtfsfj,fieldJBXXxm,fieldJKMjt,fieldSTQKzljgmc,fieldCXXXzhycjcsj,fieldJBXXxb,fieldSTQKglkssj,fieldYCBJ,fieldSTQKzysj,fieldJBXXgh,fieldCXXXfxxq,fieldSTQKqtqksm,fieldCXXXqjymsxgqk,fieldYCBZ,fieldSTQKjmy,fieldSTQKjtcyxjwjjt,fieldJBXXxnjzbgdz,fieldCXXXddsj,fieldSTQKfrsj,fieldSTQKgldd,fieldCXXXfxcfsj,fieldJTCZDZshi,fieldSTQKks,fieldCXXXjtzzq,fieldJBXXJG,fieldCXXXjtzzs,fieldJBXXshi,fieldSTQKjtcyfrtw,fieldSTQKjtcystzk1,fieldCXXXjcdqk,fieldSTQKzdmc,fieldSFJZYMyczd,fieldSTQKjtcyks,fieldCXXXjtfshc,fieldYMTGSzdqt,fieldCXXXcqwdq,fieldSTQKxjwjjt,fieldSTQKlt,fieldYMJZRQzd,fieldYQJLjcdryjkqk,fieldSTQKjtcyhxkn,fieldJBXXjgq,fieldJBXXjgs,fieldSTQKjtcylt,fieldSTQKzdjgmcc,fieldJBXXqjtxxqk,fieldSTQKjtcyglfs",
					"csrfToken":   csrfToken,
					"lang":        "zh",
				}
				client.Post(constant.FirstCommitURL, first)
			} else if r.Request.URL.String() == constant.FirstCommitURL && r.Request.Method == "POST" {
				data, _ := json.Marshal(clockInData)
				second := map[string]string{
					"actionId":    "1",
					"formData":    string(data),
					"rand":        strconv.Itoa(rand.Intn(9999)),
					"nextUsers":   "{}",
					"stepId":      stepId,
					"timestamp":   strconv.Itoa(int(time.Now().Unix())),
					"boundFields": "fieldYMTGSzd,fieldSTQKjtcyzd,fieldSTQKfrtw,fieldCXXXjtfslc,fieldSTQKxgqksm,fieldJKMsfwlm,fieldJKHDDzt,fieldYQJLzhycjcsj,fieldSTQKfl,fieldSTQKhxkn,fieldCXXXsfylk,fieldSTQKglfs,fieldCXXXsfjcgyshqzbl,fieldSTQKjtcyfx,fieldCXXXszsqsfyyshqzbl,fieldJCDDs,fieldSTQKjtcyfs,fieldSTQKjtcyzljgmc,fieldSQSJ,fieldJBXXnj,fieldSTQKfx,fieldSTQKfs,fieldYQJLjcddshi,fieldHQRQ,fieldSTQKjtcyqtms,fieldCXXXksjcsj,fieldSTQKjtcyxm,fieldZJYCHSJCYXJGRQzd,fieldqjymsjtqk,fieldCXXXjcdr,fieldCXXXsftjhbjtdz,fieldJCDDq,fieldSFJZYM,fieldSTQKjtcyclfs,fieldSTQKxm,fieldSTQKjtcyzdjgmcc,fieldSTQKqt,fieldCXXXlksj,fieldJBXXfdy,fieldSTQKjtcyjmy,fieldCXXXsftjhbq,fieldSTQKqtms,fieldYCFDY,fieldCXXXjtfspc,fieldSTQKbrstzk1,fieldCXXXssh,fieldJBXXjgjdwbk,fieldLYYZM,fieldCNS,fieldJBXXjzdz,fieldSTQKclfs,fieldSTQKjtcyfl,fieldSTQKjtcyzdjgmc,fieldJBXXbj,fieldSTQKjtcyfxx,fieldJBXXcsny,fieldCXXXdqszd,fieldSTQKjtcystzk,fieldSTQKjtcypcsj,fieldJBXXqu,fieldJBXXjgshi,fieldYQJLjcddq,fieldYQJLjcdds,fieldCXXXjtzz,fieldCXXXjtfsqt,fieldJTCZDZqu,fieldDQSJ,fieldSTQKzdjgmc,fieldJTCZDZxxdz,fieldSTQKjtcyglkssj,fieldCXXXsftjhb,fieldJTCZDZJDcode,fieldzgzjzdzjtdz,fieldJCDDqmsjtdd,fieldSHENGYC,fieldYQJLksjcsj,fieldJBXXjgsjtdz,fieldSTQKbrstzk,fieldSTQKjtcyqt,fieldJBXXlxfs,fieldSTQKpcsj,fieldYQJLsfjcqtbl,fieldJTCZDZsheng,fieldJBXXbz,fieldFLid,fieldjgs,fieldJCDDshi,fieldSTQKrytsqkqsm,fieldzgzjzdzs,fieldzgzjzdzq,fieldJZDZC,fieldSTQKjtcyzdkssj,fieldYQJLjcdry,fieldCXXXjtfsdb,fieldCXXXcxzt,fieldCXXXjtjtzz,fieldCXXXsftjhbs,fieldSTQKzdkssj,fieldSTQKfxx,fieldJTCZDZJDwbk,fieldSTQKjtcyzysj,fieldjgshi,fieldJBXXsheng,fieldJBXXdrsfwc,fieldJBXXdw,fieldCXXXjtgjbc,fieldJBXXjgjdcode,fieldSTQKjtcygldd,fieldzgzjzdzshi,fieldSTQKzd,fieldSTQKjtcyfrsj,fieldCXXXjtfsqtms,fieldSTQKjtcyzdmc,fieldCXXXjtfsfj,fieldJBXXxm,fieldJKMjt,fieldSTQKzljgmc,fieldCXXXzhycjcsj,fieldJBXXxb,fieldSTQKglkssj,fieldYCBJ,fieldSTQKzysj,fieldJBXXgh,fieldCXXXfxxq,fieldSTQKqtqksm,fieldCXXXqjymsxgqk,fieldYCBZ,fieldSTQKjmy,fieldSTQKjtcyxjwjjt,fieldJBXXxnjzbgdz,fieldCXXXddsj,fieldSTQKfrsj,fieldSTQKgldd,fieldCXXXfxcfsj,fieldJTCZDZshi,fieldSTQKks,fieldCXXXjtzzq,fieldJBXXJG,fieldCXXXjtzzs,fieldJBXXshi,fieldSTQKjtcyfrtw,fieldSTQKjtcystzk1,fieldCXXXjcdqk,fieldSTQKzdmc,fieldSFJZYMyczd,fieldSTQKjtcyks,fieldCXXXjtfshc,fieldYMTGSzdqt,fieldCXXXcqwdq,fieldSTQKxjwjjt,fieldSTQKlt,fieldYMJZRQzd,fieldYQJLjcdryjkqk,fieldSTQKjtcyhxkn,fieldJBXXjgq,fieldJBXXjgs,fieldSTQKjtcylt,fieldSTQKzdjgmcc,fieldJBXXqjtxxqk,fieldSTQKjtcyglfs",
					"csrfToken":   csrfToken,
					"lang":        "zh",
				}
				client.Post(constant.SecondCommitURL, second)
			} else if r.Request.URL.String() == constant.SecondCommitURL && r.Request.Method == "POST" {
				completeFlag <- string(r.Body)
			}
		}
	})
	client.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Referer", constant.ClockInURL)
	})
	client.Visit(constant.LoginURL)
	tk := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-loginOK:
			// 获取流水号
			client.Visit(constant.ClockInURL)
		case u, _ := <-serialNumberURL:
			snURL = u
			client.Visit(u)
		case b, _ := <-completeFlag:
			if strings.Contains(b, "SUCCEED") {
				log.Println(b)
				return true, nil
			} else {
				return false, errors.New("打卡失败:" + b)
			}
		case <-tk.C:
			return false, errors.New("打卡失败:打卡时间超时")
		}
	}
}

func getLoginReq(data io.Reader, username, password string) (map[string]string, error) {
	doc, err := html.Parse(data)
	if err != nil {
		return nil, err
	}
	p1 := htmlquery.FindOne(doc, `//div[@class="login-tab-details"]/input[4]/@value`)
	rsaStr := username + password + htmlquery.SelectAttr(p1, "value")
	rsaStr, _ = GenerateRsa(rsaStr)
	lt := htmlquery.FindOne(doc, `//div[@class="login-tab-details"]/input[4]/@value`)
	execution := htmlquery.FindOne(doc, `//div[@class="login-tab-details"]/input[5]/@value`)
	return map[string]string{
		"rsa":       rsaStr,
		"ul":        strconv.Itoa(len(username)),
		"pl":        strconv.Itoa(len(password)),
		"lt":        htmlquery.SelectAttr(lt, "value"),
		"execution": htmlquery.SelectAttr(execution, "value"),
		"_eventId":  "submit",
	}, nil
}
