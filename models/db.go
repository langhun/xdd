package models

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var JD_COOKIE = "JD_COOKIE"
var RECORD = "RECORD"
var ENV = "env"
var TASK = "TASK"
var keys map[string]bool
var pins map[string]bool

func initDB() {
	var err error
	var c = &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	if strings.Contains(Config.Database, "@tcp(") {
		db, err = gorm.Open(mysql.Open(Config.Database), c)
	} else if strings.Contains(Config.Database, "dbname=") {
		db, err = gorm.Open(postgres.Open(Config.Database), c)
	} else {
		db, err = gorm.Open(sqlite.Open(Config.Database), c)
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(
		&JdCookie{},
		&JdCookiePool{},
		&User{},
		&UserAgent{},
		&Env{},
		&Wish{},
	)
	pins = make(map[string]bool)
	keys = make(map[string]bool)
	jps := []JdCookiePool{}
	db.Find(&jps)
	for _, jp := range jps {
		keys[jp.PtKey] = true
		pins[jp.PtPin] = true
	}
}

func HasPin(pin string) bool {
	if _, ok := pins[pin]; ok {
		return ok
	}
	pins[pin] = true
	return false
}

func HasKey(key string) bool {
	if _, ok := keys[key]; ok {
		return ok
	}
	keys[key] = true
	return false
}

func HasWsKeys(wskey string) bool {
	if _, ok := keys[wskey]; ok {
		return ok
	}
	keys[wskey] = true
	return false
}

type JdCookie struct {
	ID           int    `gorm:"column:ID;primaryKey"`
	Priority     int    `gorm:"column:Priority;default:1"`
	CreateAt     string `gorm:"column:CreateAt"`
	WsKey        string `gorm:"column:WsKey"`
	PtKey        string `gorm:"column:PtKey"`
	PtPin        string `gorm:"column:PtPin;unique"`
	Note         string `gorm:"column:Note"`
	Available    string `gorm:"column:Available;default:true" validate:"oneof=true false"`
	Nickname     string `gorm:"column:Nickname"`
	BeanNum      string `gorm:"column:BeanNum"`
	QQ           int    `gorm:"column:QQ"`
	PushPlus     string `gorm:"column:PushPlus"`
	Telegram     int    `gorm:"column:Telegram"`
	Fruit        string `gorm:"column:Fruit"`
	Pet          string `gorm:"column:Pet"`
	Bean         string `gorm:"column:Bean"`
	JdFactory    string `gorm:"column:JdFactory"`
	DreamFactory string `gorm:"column:DreamFactory"`
	Jxnc         string `gorm:"column:Jxnc"`
	Jdzz         string `gorm:"column:Jdzz"`
	Joy          string `gorm:"column:Joy"`
	Sgmh         string `gorm:"column:Sgmh"`
	Cfd          string `gorm:"column:Cfd"`
	Cash         string `gorm:"column:Cash"`
	Help         string `gorm:"column:Help;default:false" validate:"oneof=true false"`
	Pool         string `gorm:"-"`
	Hack         string `gorm:"column:Hack"  validate:"oneof=true false"`
	UserLevel    string `gorm:"column:UserLevel"`
	LevelName    string `gorm:"column:LevelName"`
}

type JdCookiePool struct {
	ID       int    `gorm:"column:ID;primaryKey"`
	WsKey    string `gorm:"column:WsKey"`
	PtKey    string `gorm:"column:PtKey;unique"`
	PtPin    string `gorm:"column:PtPin"`
	LoseAt   string `gorm:"column:LoseAt"`
	CreateAt string `gorm:"column:CreateAt"`
}

var UserLevel = "UserLevel"
var LevelName = "LevelName"
var ScanedAt = "ScanedAt"
var LoseAt = "LoseAt"
var CreateAt = "CreateAt"
var Note = "Note"
var Available = "Available"
var UnAvailable = "UnAvailable"
var WsKey = "WsKey"
var PtKey = "PtKey"
var PtPin = "PtPin"
var Priority = "Priority"
var Nickname = "Nickname"
var BeanNum = "BeanNum"
var Pool = "Pool"
var True = "true"
var False = "false"
var QQ = "QQ"
var PushPlus = "PushPlus"
var Save chan *JdCookie
var ExecPath string
var Telegram = "Telegram"
var Hack = "Hack"
var Address = "Address"

const (
	Fruit        = "Fruit"
	Pet          = "Pet"
	Bean         = "Bean"
	JdFactory    = "JdFactory"
	DreamFactory = "DreamFactory"
	Jxnc         = "Jxnc"
	Jdzz         = "Jdzz"
	Joy          = "Joy"
	Sgmh         = "Sgmh"
	Cfd          = "Cfd"
	Cash         = "Cash"
	Help         = "Help"
)

func Date() string {
	return time.Now().Local().Format("2006-01-02")
}

func GetJdCookies(sbs ...func(sb *gorm.DB) *gorm.DB) []JdCookie {
	cks := []JdCookie{}
	tb := db
	for _, sb := range sbs {
		tb = sb(tb)
	}
	tb.Order("priority desc").Find(&cks)
	return cks
}

func GetJdCookie(pin string) (*JdCookie, error) {
	ck := &JdCookie{}
	return ck, db.Where(PtPin+" = ?", pin).First(ck).Error
}

func (ck *JdCookie) Updates(values interface{}) {
	if ck.ID != 0 {
		db.Model(ck).Updates(values)
	}
	if ck.PtPin != "" {
		db.Model(ck).Where(PtPin+" = ?", ck.PtPin).Updates(values)
	}
}

func (ck *JdCookie) Update(column string, value interface{}) {
	if ck.ID != 0 {
		db.Model(ck).Update(column, value)
	}
	if ck.PtPin != "" {
		db.Model(JdCookie{}).Where(PtPin+" = ?", ck.PtPin).Update(column, value)
	}
}

func (ck *JdCookie) Removes(values interface{}) {
	if ck.ID != 0 {
		db.Model(ck).Delete(values)
	}
	if ck.PtPin != "" {
		db.Model(ck).Where(PtPin+" = ?", ck.PtPin).Delete(values)
	}
}

func (ck *JdCookie) InPool(pt_key string) error {
	if ck.ID != 0 {
		date := Date()
		tx := db.Begin()
		jp := &JdCookiePool{}
		if tx.Where(fmt.Sprintf("%s = '%s' and %s = '%s'", PtPin, ck.PtPin, PtKey, pt_key)).First(jp).Error == nil {
			return tx.Rollback().Error
		}
		go test2(fmt.Sprintf("pt_key=%s;pt_pin=%s;", pt_key, ck.PtPin))
		if err := tx.Create(&JdCookiePool{
			PtPin:    ck.PtPin,
			PtKey:    pt_key,
			WsKey:    ck.WsKey,
			CreateAt: date,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
		tx.Model(ck).Updates(map[string]interface{}{
			Available: True,
			PtKey:     pt_key,
		})
		return tx.Commit().Error
	}
	return nil
}

func (ck *JdCookie) OutPool() (string, error) {
	if ck.ID != 0 {
		date := Date()
		tx := db.Begin()
		jp := &JdCookiePool{}
		tx.Model(jp).Where(fmt.Sprintf("%s = '%s' and %s = '%s'", PtPin, ck.PtPin, PtKey, ck.PtKey)).Update(LoseAt, date)
		us := map[string]interface{}{}
		if tx.Where(fmt.Sprintf("%s = '%s' and %s = '%s'", PtPin, ck.PtPin, LoseAt, "")).First(jp).Error != nil {
			us[Available] = False
			us[PtKey] = ""
		} else {
			us[Available] = True
			us[PtKey] = jp.PtKey
		}
		e := tx.Model(ck).Updates(us).RowsAffected
		if e == 0 {
			tx.Rollback()
			return "", nil
		}
		ck.Available = us[Available].(string)
		ck.PtKey = jp.PtKey
		return jp.PtKey, tx.Commit().Error
	}
	return "", nil
}

func NewJdCookie(ck *JdCookie) error {
	if ck.Hack == "" {
		ck.Hack = False
	}
	ck.Priority = Config.DefaultPriority
	date := Date()
	ck.CreateAt = date
	tx := db.Begin()
	if err := tx.Create(ck).Error; err != nil {
		tx.Rollback()
		return err
	}
	go test2(fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin))
	if err := tx.Create(&JdCookiePool{
		PtPin:    ck.PtPin,
		PtKey:    ck.PtKey,
		WsKey:    ck.WsKey,
		CreateAt: date,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func UpdateCookie(ck *JdCookie) error {
	if ck.Hack == "" {
		ck.Hack = False
	}
	ck.Priority = Config.DefaultPriority
	date := Date()
	ck.CreateAt = date
	tx := db.Begin()
	if err := tx.Updates(ck).Error; err != nil {
		tx.Rollback()
		return err
	}
	go test2(fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin))
	if err := tx.Create(&JdCookiePool{
		PtPin:    ck.PtPin,
		PtKey:    ck.PtKey,
		WsKey:    ck.WsKey,
		CreateAt: date,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func updateCookie() {
	cks := GetJdCookies()
	xya := 0
	xyb := 0
	for i := range cks {
		if len(cks[i].WsKey) > 0 {
			xya++
			time.Sleep(time.Second * time.Duration(rand.Intn(20)))
			ck := cks[i]
			var pinkey = fmt.Sprintf("pin=%s;wskey=%s;", ck.PtPin, ck.WsKey)
			rsp := cmd(fmt.Sprintf(`python3 wspt.py "%s"`, pinkey), &Sender{})
			if strings.Contains(rsp, "错误") {
				ck.Push(fmt.Sprintf("Wskey失效账号，%s", ck.PtPin))
			} else {
				ptKey := FetchJdCookieValue("pt_key", rsp)
				ptPin := FetchJdCookieValue("pt_pin", rsp)
				ck := JdCookie{
					PtKey: ptKey,
					PtPin: ptPin,
				}
				if CookieOK(&ck) {
					xyb++
					if HasKey(ck.PtKey) {
						(&JdCookie{}).Push(fmt.Sprintf("重复提交"))
					} else {
						if nck, err := GetJdCookie(ck.PtPin); err == nil {
							nck.InPool(ck.PtKey)
							msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
							//(&JdCookie{}).Push(msg)
							logs.Info(msg)
						} else {
							NewJdCookie(&ck)
							msg := fmt.Sprintf("添加账号，账号名:%s", ck.PtPin)
							logs.Info(msg)
						}
					}
				} else {
					(&JdCookie{}).Push(fmt.Sprintf("无效CK转换失败，%s", ck.PtPin))
				}
			}
		} else {
			//(&JdCookie{}).Push(fmt.Sprintf("转换失败，请重新转换，%s", ck.PtPin))
		}
		go func() {
			Save <- &JdCookie{}
		}()
	}
	(&JdCookie{}).Push(fmt.Sprintf("所有wskey转换完成，共%d个，成功%d个。", xya, xyb))
}

func CheckIn(pin, key string) int {
	if !HasPin(pin) {
		NewJdCookie(&JdCookie{
			PtKey: key,
			PtPin: pin,
			Hack:  False,
		})
		return 0
	} else if !HasKey(key) {
		ck, _ := GetJdCookie(pin)
		ck.InPool(key)
		return 1
	}
	return 2
}

func FetchJdCookieValue(key string, cookies string) string {
	match := regexp.MustCompile(key + `=([^;]*);{0,1}`).FindStringSubmatch(cookies)
	if len(match) == 2 {
		return match[1]
	} else {
		return ""
	}
}
