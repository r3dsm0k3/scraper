package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"scraper/db"
	"scraper/utils"
)

type Funda struct {
	Hunter *colly.Collector
	Queue  *utils.Queue
	Db     *db.ApartmentDb
}

func (f *Funda) Visit() {
	domain := "www.funda.nl"
	location := "amsterdam"
	minPrice := 1000
	maxPrice := 2000
	terrace := "dakterras"
	garden := "tuin"
	name := "Funda"
	maxPages := 5
	currentPage := 1
	searchQuery := fmt.Sprintf("/en/huur/%s/%d-%d/%s/%s/sorteer-datum-af/", location, minPrice, maxPrice, terrace, garden)
	url := fmt.Sprintf("https://%s%s", domain, searchQuery)

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	f.Hunter.CacheDir = "./_hunter/funda"

	f.Hunter.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36"
	f.Hunter.Async = true
	f.Hunter.AllowURLRevisit = true
	f.Hunter.OnHTML("div.search-result-content-inner", func(e *colly.HTMLElement) {

		address := e.ChildText("div.search-result__header > div.search-result__header-title-col > a > h3.search-result__header-title")
		zipCode := e.ChildText("div.search-result__header > div.search-result__header-title-col > a > h4.search-result__header-subtitle")
		url := e.ChildAttr("div.search-result__header > div.search-result__header-title-col > a:nth-child(1)", "href")
		price := e.ChildText("div.search-result-info.search-result-info-price > span.search-result-price")
		fullUrl := fmt.Sprintf("https://%s%s", domain, url)
		apartment := utils.PotentialApartment{
			URL:      fullUrl,
			Rent:     price,
			Location: address,
			ZipCode:  zipCode,
		}
		if !f.Db.CheckApartmentExists(address) {
			err := f.Db.AddApartment(&apartment)
			if err == nil {
				f.Queue.Channel <- apartment
			} else {
				log.Printf("there was an error in adding the apartment %v", err.Error())
			}
		}
	})
	// find the next page
	f.Hunter.OnHTML("#content > form > div.container.search-main > nav > a", func(e *colly.HTMLElement) {
		if currentPage > maxPages {
			return
		}
		url := e.Attr("href")
		fullUrl := fmt.Sprintf("https://%s%s", domain, url)
		f.Hunter.Visit(fullUrl)
		currentPage += 1
	})

	f.Hunter.OnRequest(func(req *colly.Request) {
		fmt.Println(name, " Visiting...", req.URL)

		req.Headers.Add("Connection", "keep-alive")
		req.Headers.Add("Cache-Control", "max-age=0")
		req.Headers.Add("Upgrade-Insecure-Requests", "1")
		req.Headers.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
		req.Headers.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		req.Headers.Add("Sec-Fetch-Site", "none")
		req.Headers.Add("Sec-Fetch-Mode", "navigate")
		req.Headers.Add("Sec-Fetch-User", "?1")
		req.Headers.Add("Sec-Fetch-Dest", "document")
		req.Headers.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
		req.Headers.Add("Cookie", "DG_ZID=66497084-93C0-300D-84BC-79B3DAB109A1; DG_ZUID=62F55E37-F43F-3818-8999-0F6BEF2528DC; DG_HID=1E69CCC7-D73F-3D9D-8339-081B7182500A; DG_SID=213.127.38.125:cygzv0U6TMVenI8n+/sFCbIgWBk2JAvBCnHkIeGzcVI; .ASPXANONYMOUS=yLO2sJpS9gpdLj9YsbNBtwV4cJPkx6fgoo0mENeUevEoS-Mo8e5w4sXtEUUAq92Y-uvgvYtP49Y8Hau-aWo3B55IETNRI75D01IY8O4XcLR5v2BWHzIUJaelD6vpyuWfp6BYLPsLze5rkkLG_kZeYfqSflA1; sr=0%7cfalse; INLB=01-002; html-classes=js supports-placeholder; fonts-loaded=true; DG_IID=19DDABC2-1F92-3B31-B63B-0AEE698D5B4C; DG_UID=A79ADC3B-27CE-3150-A16D-5032DE3F33E0; lzo=huur=%2fhuur%2famsterdam%2f1000-2000%2f; SNLB2=12-001; OptanonAlertBoxClosed=2020-05-10T09:44:47.104Z; OptanonConsent=isIABGlobal=false&datestamp=Sun+May+10+2020+11%3A44%3A47+GMT%2B0200+(Central+European+Summer+Time)&version=6.0.0&landingPath=NotLandingPage&groups=F01%3A1%2CF02%3A1%2CF03%3A1%2CF05%3A1%2CBG9%3A1&hosts=&legInt=&AwaitingReconsent=false; eupubconsent=BOzLgKXOzLgKXAcABBNLDI-AAAAvZ7_______9______9uz_Ov_v_f__33e8__9v_l_7_-___u_-33d4u_1vf99yfm1-7etr3tp_87ues2_Xur__71__3z3_9pxP78k89r7337Ew_v-_v-b7BCPN9Y3v-8Kg")
		req.Headers.Add("Cookie", ".ASPXANONYMOUS=chK3M2tSOjVo8VoSdC6r7UJLlfpl64RG4aFQlGnED6NC2cNZvK2_TNQ2xt-GlnjEGEovN9lyOKDfrU4zFoO5lZeSDgdAV7RXJQHSnsZqAs6pvoFmdRoINe60aeAeeFzvv0051VyBjF_gPvQ-jO6Vz18bFL41; sr=0%7cfalse; SNLB2=12-002; rtb-platform=improve; INLB=01-002")
	})
	f.Hunter.OnError(func(response *colly.Response, err error) {
		log.Println(string(response.Body))
		log.Println(response.StatusCode)
		log.Println(err)
	})
	////enable this for debugging the response
	//f.Hunter.OnResponse(func(response *colly.Response) {
	//	fmt.Println(response.Headers)
	//
	//})
	f.Hunter.Visit(url)
}