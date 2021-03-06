package main

import (
	"database/sql"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"strconv"
	"time"
)

var (
	BotToken string
	Socks5   string
	bot      *tb.Bot
	db       *sql.DB
)

const (
	dbDriverName = "mysql"
	logo         = `
  ______ _____ _____       _     ____        _   
 |  ____| ____/ ____|     | |   |  _ \      | |  
 | |__  | |__| (___  _   _| |__ | |_) | ___ | |_ 
 |  __| |___ \\___ \| | | | '_ \|  _ < / _ \| __|
 | |____ ___) |___) | |_| | |_) | |_) | (_) | |_ 
 |______|____/_____/ \__,_|_.__/|____/ \___/ \__|
`
	//dbName       = "./data.db"
)

func init() {
	//read config
	fmt.Println(logo)
	fmt.Println("Read Config……")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	CheckErr(err)
	BotToken = viper.GetString("bot_token")
	Socks5 = viper.GetString("socks5")
	//set bot
	fmt.Println("Bot Settings……")
	Poller := &tb.LongPoller{Timeout: 15 * time.Second}
	spamProtected := tb.NewMiddlewarePoller(Poller, func(upd *tb.Update) bool {
		if upd.Message == nil {
			return true
		}
		if !upd.Message.Private() {
			return false
		}
		return true
	})
	botsettings := tb.Settings{
		Token:  BotToken,
		Poller: spamProtected,
	}
	//set socks5
	if Socks5 != "" {
		fmt.Println("Proxy:" + Socks5)
		dialer, err := proxy.SOCKS5("tcp", Socks5, nil, proxy.Direct)
		CheckErr(err)
		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		botsettings.Client = httpClient
	}
	//create bot
	bot, err = tb.NewBot(botsettings)
	if err != nil {
		fmt.Println("Create Bot ERROR!")
		return
	}
	fmt.Println("Bot: " + strconv.Itoa(bot.Me.ID) + " " + bot.Me.Username)
}
func main() {
	BotStart()
}
func BotStart() {
	MakeHandle()
	TaskLaunch()
	fmt.Println("Bot Start")
	fmt.Println("------------")
	bot.Start()
}
func MakeHandle() {
	fmt.Println("Make Handle……")
	bot.Handle("/start", bStart)
	bot.Handle("/my", bMy)
	bot.Handle("/bind", bBind)
	bot.Handle("/unbind", bUnBind)
	bot.Handle("/notice", bNotice)
	bot.Handle("/help", bHelp)
	bot.Handle(tb.OnText, bOnText)
	//bot.Handle(tb.InlineButton{Unique: ""})
}
func TaskLaunch() {
	fmt.Println("Begin First SignTask……")
	task := cron.New()
	SignTask()
	//每三小时执行一次
	task.AddFunc("1 */3 * * *", SignTask)
	//  */1 * * * *
	fmt.Println("Cron Task Start……")
	task.Start()
}
