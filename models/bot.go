package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
)

var SendQQ = func(a int64, b interface{}) {

}
var SendQQGroup = func(a int64, b int64, c interface{}) {

}
var ListenQQPrivateMessage = func(uid int64, msg string) {
	SendQQ(uid, handleMessage(msg, "qq", int(uid)))
}

var ListenQQGroupMessage = func(gid int64, uid int64, msg string) {
	if gid == Config.QQGroupID {
		if Config.QbotPublicMode {
			SendQQGroup(gid, uid, handleMessage(msg, "qqg", int(uid), int(gid)))
		} else {
			SendQQ(uid, handleMessage(msg, "qq", int(uid)))
		}
	}
}

var replies = map[string]string{}

func InitReplies() {
	f, err := os.Open(ExecPath + "/conf/reply.php")
	if err == nil {
		defer f.Close()
		data, _ := ioutil.ReadAll(f)
		ss := regexp.MustCompile("`([^`]+)`\\s*=>\\s*`([^`]+)`").FindAllStringSubmatch(string(data), -1)
		for _, s := range ss {
			replies[s[1]] = s[2]
		}
	}
	if _, ok := replies["壁纸"]; !ok {
		replies["壁纸"] = "https://acg.toubiec.cn/random.php"
	}
}

var handleMessage = func(msgs ...interface{}) interface{} {
	msg := msgs[0].(string)
	args := strings.Split(msg, " ")
	head := args[0]
	contents := args[1:]
	sender := &Sender{
		UserID:   msgs[2].(int),
		Type:     msgs[1].(string),
		Contents: contents,
	}
	if len(msgs) >= 4 {
		sender.ChatID = msgs[3].(int)
	}
	if sender.Type == "tgg" {
		sender.MessageID = msgs[4].(int)
		sender.Username = msgs[5].(string)
		sender.ReplySenderUserID = msgs[6].(int)
	}
	if sender.UserID == Config.TelegramUserID || sender.UserID == int(Config.QQID) {
		sender.IsAdmin = true
	}
	for i := range codeSignals {
		for j := range codeSignals[i].Command {
			if codeSignals[i].Command[j] == head {
				return func() interface{} {
					if codeSignals[i].Admin && !sender.IsAdmin {
						return "你没有权限操作"
					}
					return codeSignals[i].Handle(sender)
				}()
			}
		}
	}
	switch msg {
	default:
		{ //tyt
			ss := regexp.MustCompile(`packetId=(\S+)(&|&amp;)currentActId`).FindStringSubmatch(msg)
			if len(ss) > 0 {
				if !sender.IsAdmin {
					coin := GetCoin(sender.UserID)
					if coin < 8 {
						return "推一推需要8个许愿币。"
					}
					RemCoin(sender.UserID, 8)
					sender.Reply("推一推即将开始，已扣除8个许愿币。")
				}
				runTask(&Task{Path: "jd_tyt.js", Envs: []Env{
					{Name: "tytpacketId", Value: ss[1]},
				}}, sender)
				return nil
			}
		}
		{ //ptkey
			ss := regexp.MustCompile(`pt_key=([^;=\s]+);pt_pin=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
			if len(ss) > 0 {

				xya := 0
				xyb := 0
				xyc := 0
				xyd := 0
				xye := 0
				for _, s := range ss {
					xya++
					ck := JdCookie{
						PtKey: s[1],
						PtPin: s[2],
					}
					if CookieOK(&ck) {
						if sender.IsQQ() {
							ck.QQ = sender.UserID
						} else if sender.IsTG() {
							ck.Telegram = sender.UserID
						}
						if HasKey(ck.PtKey) {
							xyd++
							sender.Reply(fmt.Sprintf("重复提交"))
						} else {
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								xyc++
								nck.InPool(ck.PtKey)
								msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
								(&JdCookie{}).Push(msg)
								logs.Info(msg)
							} else {
								if Cdle {
									ck.Hack = True
								}
								xyb++
								NewJdCookie(&ck)
								msg := fmt.Sprintf("添加账号，%s", ck.PtPin)
								sender.Reply(fmt.Sprintf("很棒，许愿币+1，余额%d", AddCoin(sender.UserID)))
								sender.Reply(ck.Query())
								logs.Info(msg)
							}
						}
					} else {
						xye++
						sender.Reply(fmt.Sprintf("无效，许愿币-1，余额%d", RemCoin(sender.UserID, 1)))
					}
				}
				go func() {
					Save <- &JdCookie{}
				}()
				(&JdCookie{}).Push(fmt.Sprintf("添加完成，共%d个，新用户%s个，更新%s个，重复%s个，失败%s个", xya, xyb, xyc, xyd, xye))
				return nil
			}
		}
		//wskey
		ss := regexp.MustCompile(`pin=([^;=\s]+);wskey=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
		xya := 0
		xyb := 0
		xyc := 0
		xyd := 0
		if len(ss) > 0 {

			for _, s := range ss {
				xya++
				pinkey := "pin=" + s[1] + ";wskey=" + s[2] + ";"
				rsp := cmd(fmt.Sprintf(`python3 wspt.py "%s"`, pinkey), &Sender{})
				logs.Info(rsp)
				if strings.Contains(rsp, "错误") || strings.Contains(rsp, "失效") {
					logs.Error("wskey错误")
					sender.Reply(fmt.Sprintf("wskey错误"))
				} else {
					ptKey := FetchJdCookieValue("pt_key", rsp)
					logs.Info(ptKey)
					ptPin := FetchJdCookieValue("pt_pin", rsp)
					if len(ptKey) > 0 {
						ck := JdCookie{
							PtPin: ptPin,
							PtKey: ptKey,
							WsKey: s[2],
						}
						if CookieOK(&ck) {
							if sender.IsQQ() {
								ck.QQ = sender.UserID
							} else if sender.IsTG() {
								ck.Telegram = sender.UserID
							}
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								xyc++
								nck.InPool(ck.PtKey)
								if nck.WsKey == "" || len(nck.WsKey) == 0 {
									if sender.IsQQ() {
										ck.Update(QQ, ck.QQ)
									}
									nck.Update(WsKey, ck.WsKey)
									msg := fmt.Sprintf("写入WsKey，并更新账号%s", ck.PtPin)
									sender.Reply(fmt.Sprintf(msg))
									(&JdCookie{}).Push(msg)
									logs.Info(msg)
								} else {
									if nck.WsKey == ck.WsKey {
										msg := fmt.Sprintf("重复wskey,更新ptkey")
										sender.Reply(fmt.Sprintf(msg))
										//(&JdCookie{}).Push(msg)
										logs.Info(msg)
									} else {
										nck.Update(WsKey, ck.WsKey)
										msg := fmt.Sprintf("更新WsKey，并更新账号%s", ck.PtPin)
										sender.Reply(fmt.Sprintf(msg))
										(&JdCookie{}).Push(msg)
										logs.Info(msg)
									}
								}
							} else {
								xyb++
								NewJdCookie(&ck)
								msg := fmt.Sprintf("添加账号成功，用户名：%s\n ptkey=%s", ck.PtPin, ck.PtKey)
								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}
								sender.Reply(fmt.Sprintf(msg))
								sender.Reply(ck.Query())
								logs.Info(msg)
							}
						}
					} else {
						xyd++
						msg := fmt.Sprintf("转换失败,pin=%s", s[1])
						//sender.Reply(fmt.Sprintf(msg))
						(&JdCookie{}).Push(msg)
						logs.Info(msg)
					}
				}
				xyd++
			}
			go func() {
				Save <- &JdCookie{}
			}()
			(&JdCookie{}).Push(fmt.Sprintf("添加完成，共%d个，新用户%s个，重复%s个，失败%s个", xya, xyb, xyc, xyd))
			return nil
		}
		{
			o := findShareCode(msg)
			if o != "" {
				return "导入互助码成功"
			}
		}
		for k, v := range replies {
			if regexp.MustCompile(k).FindString(msg) != "" {
				if strings.Contains(msg, "妹") && time.Now().Unix()%10 == 0 {
					v = "https://pics4.baidu.com/feed/d833c895d143ad4bfee5f874cfdcbfa9a60f069b.jpeg?token=8a8a0e1e20d4626cd31c0b838d9e4c1a"
				}
				if regexp.MustCompile(`^https{0,1}://[^\x{4e00}-\x{9fa5}\n\r\s]{3,}$`).FindString(v) != "" {
					url := v
					rsp, err := httplib.Get(url).Response()
					if err != nil {
						return nil
					}
					ctp := rsp.Header.Get("content-type")
					if ctp == "" {
						rsp.Header.Get("Content-Type")
					}
					if strings.Contains(ctp, "text") || strings.Contains(ctp, "json") {
						data, _ := ioutil.ReadAll(rsp.Body)
						return string(data)
					}
					return rsp
				}
				return v
			}
		}
	}
	return nil
}
