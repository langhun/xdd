package models

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
)

var version = "2021092004"
var describe = "不改了~不改了~~真的不改了..."
var AppName = "xdd"
var pname = regexp.MustCompile(`/([^/\s]+)`).FindStringSubmatch(os.Args[0])[1]

func initVersion() {
	//if !Cdle {
	//	cmd("mv ../../xdd/.xdd.db ../../.xdd.db", &Sender{})
	//	cmd("rm -rf ../../xdd", &Sender{})
	//	panic("bye bye")
	//}
	if Config.Version != "" {
		version = Config.Version
	}
	logs.Info("检查更新" + version)
	value, err := httplib.Get(GhProxy + "https://raw.githubusercontent.com/langhun/xdd/x/models/version.go").String()
	if err != nil {
		logs.Info("获取版本失败🤣")
	} else {
		// name := AppName + "_" + runtime.GOOS + "_" + runtime.GOARCH
		if match := regexp.MustCompile(`var version = "(\d{10})"`).FindStringSubmatch(value); len(match) != 0 {
			des := regexp.MustCompile(`var describe = "([^"]+)"`).FindStringSubmatch(value)
			if len(des) != 0 {
				describe = des[1]
			}
			if match[1] > version {
				(&JdCookie{}).Push("小弟弟有更新了呢~😍：" + describe)
				err := Update(&Sender{})
				if err != nil {
					logs.Warn("更新失败😭", err)
					return
				}
				Daemon()
			}
		}
	}
}

func Update(sender *Sender) error {
	sender.Reply("小弟弟要开始拉取更新代码了😊。")
	rtn, err := exec.Command("sh", "-c", "cd "+ExecPath+" && git stash && git pull").Output()
	if err != nil {
		return errors.New("怎么回事？小弟弟拉取代更新失败了呢😭" + err.Error())
	}
	t := string(rtn)
	if !strings.Contains(t, "changed") {
		if strings.Contains(t, "Already") || strings.Contains(t, "已经是最新") {
			return errors.New("小弟弟已是最新版啦👌")
		} else {
			return errors.New("小弟弟拉取代失败😒" + t)
		}
	} else {
		sender.Reply("小弟弟拉取代码成功啦~😋")
	}
	sender.Reply("小弟弟正在努力加工中💪")
	rtn, err = exec.Command("sh", "-c", "cd "+ExecPath+" && go build -o "+pname).Output()
	if err != nil {
		return errors.New("小弟弟编译失败：" + err.Error())
	} else {
		sender.Reply("小弟弟要准备起来了😎")
	}
	return nil
}
