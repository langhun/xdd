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


func wstopt() {

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

				xyb := 0
				for _, s := range ss {
					ck := JdCookie{
						PtKey: s[1],
						PtPin: s[2],
					}
					if CookieOK(&ck) {
						xyb++
						if sender.IsQQ() {
							ck.QQ = sender.UserID
						} else if sender.IsTG() {
							ck.Telegram = sender.UserID
						}
						if HasKey(ck.PtKey) {
							sender.Reply(fmt.Sprintf("重复提交"))
						} else {
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								nck.InPool(ck.PtKey)
								msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
								(&JdCookie{}).Push(msg)
								logs.Info(msg)
							} else {
								if Cdle {
									ck.Hack = True
								}
								NewJdCookie(&ck)
								msg := fmt.Sprintf("添加账号，%s", ck.PtPin)
								sender.Reply(fmt.Sprintf("很棒，许愿币+1，余额%d", AddCoin(sender.UserID)))
								logs.Info(msg)
							}
						}
					} else {
						sender.Reply(fmt.Sprintf("无效，许愿币-1，余额%d", RemCoin(sender.UserID, 1)))
					}
				}
				go func() {
					Save <- &JdCookie{}
				}()
				return nil
			}
		}
		{ //wskey
			ss := regexp.MustCompile(`pin=([^;=\s]+);wskey=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
			if len(ss) > 0 {
				xyb := 0
				for _, s := range ss {
					ck := JdCookie{
						PtPin: s[1],
						WsKey: s[2],
					}
					if HasWsKeys(ck.WsKey) {
						sender.Reply(fmt.Sprintf("11111已有wskey，开始转换..."))
						if nck, err := GetJdCookie(ck.PtPin); err == nil {
							sender.Reply(fmt.Sprintf("已有wskey，开始转换..."))
							wstopt := simpleCmd(fmt.Sprintf(`wskey="pin=%s;wskey=%s;" python3 wspt.py`, ck.PtPin,ck.WsKey))
							wspt := fmt.Sprintf(`"wskey=%s;%s"`, ck.WsKey, wstopt)
							ss1 := regexp.MustCompile(`wskey=([^;=\s]+);pt_key=([^;=\s]+);pt_pin=([^;=\s]+);`).FindStringSubmatch(wspt)
							if ss1 != nil {
								ck1 := JdCookie{PtKey: ss1[2], PtPin: ck.PtPin}
									if CookieOK(&ck1) {
										xyb++
										nck.InPool(ck1.PtKey)
										sender.Reply(fmt.Sprintf("%s;%s",ck.WsKey,ck1.PtKey))
										nck.addwskey(ck.WsKey,ck1.PtKey)
										msg := fmt.Sprintf("更新账号成功：%s", ck.PtPin)
										(&JdCookie{}).Push(msg)
										logs.Info(msg)
									}else {
										sender.Reply(fmt.Sprintf("!!!更新失败!!!\n账号:%s,获取到的ck无效\nwskey过期了？？？", ck.PtPin))
										sender.Reply(fmt.Sprintf("替换wskey中..."))
										nck.addwskey(ck1.WsKey,ck1.PtKey)
										sender.Reply(fmt.Sprintf("替换成功。再次试试？"))
									}
								}
							}
						}else {
							sender.Reply(fmt.Sprintf("没有wskey，"))
							wstopt := simpleCmd(fmt.Sprintf(`wskey="pin=%s;wskey=%s;" python3 wspt.py`, ck.PtPin,ck.WsKey))
							wspt := fmt.Sprintf(`"wskey=%s;%s"`, ck.WsKey, wstopt)
							sender.Reply(fmt.Sprintf("没有wskey，\n%s",wspt))
							ss1 := regexp.MustCompile(`wskey=([^;=\s]+);pt_key=([^;=\s]+);pt_pin=([^;=\s]+);`).FindStringSubmatch(wspt)
							if ss1 != nil {
								ck1 := JdCookie{WsKey: ss1[1], PtPin: ck.PtPin, PtKey: ss1[2]}
								if CookieOK(&ck1) {
									xyb++
									if Cdle {
										ck.Hack = True
									}
									NewWskey(&ck1)
									sender.Reply(fmt.Sprintf("添加账号成功：%s", ck.PtPin))
									sender.Reply(fmt.Sprintf("很棒，许愿币+1，余额%d", AddCoin(sender.UserID)))
								}else {
									sender.Reply(fmt.Sprintf("!!!更新失败!!!\n账号:%s,未获取到 pt_key,执行结果为:%s", ck1.PtPin, ck1))
								}
							}
						}
					}
					go func() {
						Save <- &JdCookie{}
					}()
					return nil
				}
			}
		/*{ //wskey
			if strings.Contains(msg, "wskey=") {
				wstopt := cmd(fmt.Sprintf(`wskey="%s" python3 wspt.py`, msg), sender)
				wspt := fmt.Sprintf(`"%s;%s"`, msg, wstopt)
				ss := regexp.MustCompile(`pin=([^;=\s]+);wskey=([^;=\s]+);pt_key=([^;=\s]+);pt_pin=([^;=\s]+)`).FindAllStringSubmatch(wspt, -1)

				if len(ss) > 0 {
					xyb := 0
					for _, s := range ss {
						ck := JdCookie{
							PtPin: s[1],
							WsKey: s[2],
							PtKey: s[3],
						}
						sender.Reply(fmt.Sprintf("pin--%s wskey--%s ptkey--%s \n", ck.PtPin, ck.WsKey, ck.PtKey))
						if CookieOK(&ck) {
							xyb++
							if sender.IsQQ() {
								ck.QQ = sender.UserID
							} else if sender.IsTG() {
								//ck.Telegram = sender.UserID
							}
							if HasKey(ck.PtKey) {
								sender.Reply(fmt.Sprintf("重复提交"))
							} else {
								if nck, err := GetJdCookie(ck.PtPin); err == nil {
									//nck.InPoolws(ck.WsKey, ck.PtKey)
									nck.addwskey(ck.WsKey, ck.PtKey)
									msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
									(&JdCookie{}).Push(msg)
									logs.Info(msg)
								} else {
									if Cdle {
										ck.Hack = True
									}
									NewWskey(&ck)
									msg := fmt.Sprintf("添加账号，%s", ck.PtPin)
									sender.Reply(fmt.Sprintf("wskey添加成功，%s", ck.WsKey))
									sender.Reply(fmt.Sprintf("很棒，许愿币+1，余额%d", AddCoin(sender.UserID)))
									logs.Info(msg)
								}
							}
						} else {
							sender.Reply(fmt.Sprintf("无效，许愿币-1，余额%d", RemCoin(sender.UserID, 1)))
						}
					}
					go func() {
						Save <- &JdCookie{}
					}()
					return nil
				}
			}
		}
		 */
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
