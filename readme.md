#Scraper

Currently only have funda scraping. 

To run it, you need to have a telegram bot created. For this, refere [here](https://docs.microsoft.com/en-us/azure/bot-service/bot-service-channel-connect-telegram?view=azure-bot-service-4.0)


The idea is to make this bot send you/or a group the information about the new listing every hour or so. So for this you need a chat id

* [Group](https://stackoverflow.com/questions/32423837/telegram-bot-how-to-get-a-group-chat-id) 
* [Personal](https://answers.splunk.com/answers/590658/telegram-alert-action-where-do-you-get-a-chat-id.html)

that the bot can send chats to. 

Export these two to the env vars as `TELEGRAM_TOKEN` and `TELEGRAM_CHAT_ID`

once it is set, run `go run main.go`

Expect bugs



