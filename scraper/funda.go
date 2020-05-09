package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"regexp"
	"scraper/db"
	"scraper/utils"
	"strings"
	"time"
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
	maxPrice := 1800
	terrace := "dakterras"
	garden := "tuin"
	name := "Funda"
	searchQuery := fmt.Sprintf("/en/huur/%s/%d-%d/%s/%s/sorteer-datum-af/", location, minPrice, maxPrice, terrace, garden)
	url := fmt.Sprintf("https://%s%s", domain, searchQuery)

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	f.Hunter.CacheDir = "./_hunter/funda"

	f.Hunter.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36"

	f.Hunter.AllowedDomains = []string{domain}
	linkVisitor := f.Hunter.Clone()
	linkVisitor.Async = false
	linkVisitor.IgnoreRobotsTxt = true
	linkVisitor.CacheDir = "./_hunter/funda/links"
	//// Limit the number of threads started by colly to two
	//// when visiting links which domains' matches "*funda.*" glob
	linkVisitor.Limit(&colly.LimitRule{
		DomainGlob:  "*funda.*",
		Parallelism: 2,
		RandomDelay: 2 * time.Second,
	})

	// visit the page and get the details.
	linkVisitor.OnHTML("#content > div > header > div > div.object-header__details", func(e *colly.HTMLElement) {
		// extract the name
		name := e.ChildText("div.object-header__details-info > h1.object-header__container > span.object-header__title")
		zipCode := e.ChildText("div.object-header__details-info > h1.object-header__container > span.object-header__subtitle")
		price := e.ChildText("div.object-header__pricing > strong.object-header__price")

		apartment := utils.PotentialApartment{
			URL:      e.Request.URL.String(),
			Rent:     price,
			Location: name,
			ZipCode:  zipCode,
		}
		fmt.Println(apartment)
		if !f.Db.CheckApartmentExists(name) {
			if err := f.Db.AddApartment(&apartment); err != nil {
				f.Queue.Channel <- apartment
			}
		}

	})
	f.Hunter.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println(link)
		if !strings.HasPrefix(link, "/en/huur/") && !strings.Contains(link, "appartement") && !strings.Contains(link, "huis") || strings.Contains(link, searchQuery) {
			return
		}
		// funda links generally have a longer char length
		if len(link) < 35 {
			return
		}
		fullUrl := fmt.Sprintf("https://%s%s", domain, link)
		fmt.Println(fullUrl)
		address := extractAddressFromLink(fullUrl)
		if !f.Db.CheckApartmentExists(address) {
			linkVisitor.Visit(fullUrl)
		}
	})
	f.Hunter.OnRequest(func(r *colly.Request) {
		fmt.Println(name, " Visiting...", r.URL)

		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
		r.Headers.Set("Accept", "text/html")
		r.Headers.Set("Sec-Fetch-Site", "same-origin")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
		r.Headers.Set("Cookie", "html-classes=js supports-placeholder; language_preference=en; rtb-platform=improve; __RequestVerificationToken=PRQCamierZAQxqaxCpxmNehW21FKliaomY3f8KrKo2YCHduw658A3-jxps7FCcqejeEjirkmdWU27zlVLUFb44oQ_J41; cookiePolicy_18=allowPersonalisatie=True&allowAdvertenties=True&explicitAcceptOfCookies=True; DG_SID=66.202.46.7:1TY2zYRYGlFii7Y8P/MNSE2XojBADXeSVyuzlfmSUmI; language-preference=en; INLB=01-008; .ASPXANONYMOUS=khl4AT5WGj4ZXKY3ey8dw2p0ScPJk-vo6tmOqWx1BGWiBez6pgItuBZYuIfzjxZ0S3Cy3RgVxOwd4aCK7yDAMSRX_ja6sjdVgxgJtEUMUSUvycj7AyPE5doD3gYdyHWzfnvoZH1tjtAI9KQe35t8pYuHaJo1; googtrans=/en/en; googtrans=/en/en; OptanonAlertBoxClosed=2020-05-04T12:34:27.052Z; eupubconsent=BOy4HY-Oy4HY-AcABBNLDHAAAAAvR7_f_v___9______9uz_Ov_v_f__33e8__9v_l_7_-___u_-33d4u_1vf99yfm1-7atr3tp_87ues2_Xur__71__3z3_9pxP78k89r7327Ew_v-_v-b7BCPN9Y3v-8KA; fonts-loaded=true; DG_ZID=85363BEA-8B58-3E9E-8773-D6860E10A1FD; DG_ZUID=9AA70BAB-31AB-3215-8E98-192BEFE1636B; DG_HID=09827BE0-78F6-3AD9-86BE-1C852792A63D; objectnotfound=objectnotfound=false; OptanonConsent=isIABGlobal=false&datestamp=Fri+May+08+2020+10%3A09%3A09+GMT%2B0200+(Central+European+Summer+Time)&version=5.15.0&landingPath=NotLandingPage&groups=F01%3A1%2CF02%3A1%2CF03%3A1%2CF05%3A1&hosts=&legInt=&AwaitingReconsent=false&geolocation=NL%3BGE; InMemoryreferrer=ObjectContactReferrer_5358086=https%3a%2f%2fwww.funda.nl%2fen%2fkoop%2famsterdam%2fappartement-41703449-boterdiepstraat-31-h%2f&ObjectContactReferrer_5413990=https%3a%2f%2fwww.funda.nl%2fen%2fhuur%2famsterdam%2fappartement-87222999-admiralengracht-42-h%2f&ObjectContactReferrer_5452629=https%3a%2f%2fwww.funda.nl%2fen%2fhuur%2famsterdam%2fappartement-41807082-narva-eiland-68%2f&ObjectContactReferrer_5453469=https%3a%2f%2fwww.funda.nl%2fen%2fhuur%2famsterdam%2fappartement-87262468-kanaalstraat-155-h%2f&ObjectContactReferrer_5455486=https%3a%2f%2fwww.funda.nl%2fen%2fhuur%2famsterdam%2fappartement-41800849-vasco-da-gamastraat-26-hs%2f; DG_IID=19DDABC2-1F92-3B31-B63B-0AEE698D5B4C; DG_UID=67488688-26F6-35C1-A42A-456515B034A6; sr=0%7ctrue%7chuur; lzo_sort=huur=%7b%22Key%22%3a%22datum%22%2c%22Value%22%3a%22Descending%22%7d; eupubconsent=BOzJTNmOzJTNmAcABBNLDIAAAAAvaAAA; OptanonConsent=isIABGlobal=false&datestamp=Sat+May+09+2020+19%3A42%3A34+GMT%2B0200+(Central+European+Summer+Time)&version=6.0.0&landingPath=NotLandingPage&groups=F01%3A1%2CF02%3A1%2CF03%3A1%2CF05%3A1&hosts=&legInt=&AwaitingReconsent=false&geolocation=NL%3BGE; lzo=koop=%2fkoop%2famsterdam%2ftuindorp-oostzaan%2f&huur=%2fhuur%2famsterdam%2f1000-1800%2fdakterras%2ftuin%2f; SNLB2=12-002")

	})
	f.Hunter.OnError(func(response *colly.Response, err error) {
		fmt.Println(string(response.Body))
		fmt.Println(response.StatusCode)
		fmt.Println(err)
	})

	// enable this for debugging the response
	//f.Hunter.OnResponse(func(response *colly.Response) {
	//	fmt.Println(string(response.Body))
	//})
	f.Hunter.Visit(url)
}
func extractAddressFromLink(link string) string {
	re := regexp.MustCompile(`[0-9]{8}(.*)$`)
	matches := re.FindStringSubmatch(link)
	if len(matches) > 0 {
		lastMatch := matches[len(matches)-1]
		//now we need to trim the first and last characters
		m := strings.TrimLeft(lastMatch, "-")
		m = strings.TrimRight(m, "/")
		return m
	}
	return ""
}
 func (f *Funda) addApartment(apartment *utils.PotentialApartment) error {
	//check if it already exists
	if !f.Db.CheckApartmentExists(apartment.Location) {
		return f.Db.AddApartment(apartment)
	}
	return nil
}
