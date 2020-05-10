package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/yanzay/tbot"
	"github.com/yanzay/tbot/model"
	"os"
	"os/signal"
	database "scraper/db"
	"scraper/scraper"
	"scraper/utils"
	"strconv"
	"strings"
	"syscall"
	"time"
)


func main() {

	queue := utils.Queue{Channel:make(chan utils.PotentialApartment, 1)}
	// defer the close
	defer close(queue.Channel)
	signalChan := make(chan os.Signal, 1)
	// defer the close
	defer close(signalChan)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGKILL,
		syscall.SIGSEGV,
		syscall.SIGSYS,
		syscall.SIGPIPE,
		syscall.SIGTERM,
	)

	// make the db
	db := database.New("./badger-db")
	defer db.Close()

	// setup the main collector
	//
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}),)

	// reject any robots.txt
	c.IgnoreRobotsTxt = true
	c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0"

	funda := scraper.Funda{Hunter:c.Clone(), Queue:&queue, Db:db}
	go funda.Visit()

	pararius := scraper.Pararius{
		Hunter: c.Clone(),
		Queue: &queue,
		Db:     db,
	}
	go pararius.Visit()
	go func() {

		for {
			select {
			case data := <- queue.Channel:
				{
					// just wait a bit between messages
					time.Sleep(2 * time.Second)
					go sendTelegramMessage(data)
				}
			}
		}
	}()
	<- signalChan
}

func sendTelegramMessage(apartment utils.PotentialApartment) {
	botToken := os.Getenv("TELEGRAM_TOKEN")
	chatId, _ := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	bot, _ := tbot.NewServer(botToken)
	mapLink := fmt.Sprintf("https://www.google.com/maps/place/%s", strings.ReplaceAll(apartment.Location + apartment.ZipCode, " ", "+"))
	markdown := fmt.Sprintf(`
*Found new apartment for you!*

*Location* : **%s**

*Rent* : **%s**.

[Click here](%s) for details

[Google Map](%s)
`, apartment.Location, apartment.Rent, apartment.URL, mapLink)

	message := model.Message{
		Type:            0,
		Data:            markdown,
		ChatID:          int64(chatId),
		OneTimeKeyboard: false,
		DisablePreview:  false,
		Markdown: true,

	}
	bot.SendMessage(&message)
}



