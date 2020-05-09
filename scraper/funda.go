package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"scraper/db"
	"scraper/utils"
	"strings"
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
	//f.Hunter.CacheDir = "./funda"

	f.Hunter.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36"

	//f.Hunter.AllowedDomains = []string{domain}
	linkVisitor := f.Hunter.Clone()

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
		if !strings.HasPrefix(link, "/en/huur/") && !strings.Contains(link, "appartement") && !strings.Contains(link, "huis") || strings.Contains(link, searchQuery) {
			return
		}
		// funda links generally have a longer char length
		if len(link) < 35 {
			return
		}
		fullUrl := fmt.Sprintf("https://%s%s", domain, link)

		linkVisitor.Visit(fullUrl)

	})
	f.Hunter.OnRequest(func(r *colly.Request) {
		fmt.Println(name, " Visiting...", r.URL)
	})
	f.Hunter.OnError(func(response *colly.Response, err error) {
		fmt.Println(string(response.Body))
		fmt.Println(response.StatusCode)
		fmt.Println(err)
	})
	f.Hunter.OnResponse(func(response *colly.Response) {
		fmt.Println(string(response.Body))
	})
	f.Hunter.Visit(url)
}

func (f *Funda) addApartment(apartment *utils.PotentialApartment) error {
	//check if it already exists
	if !f.Db.CheckApartmentExists(apartment.Location) {
		return f.Db.AddApartment(apartment)
	}
	return nil
}
