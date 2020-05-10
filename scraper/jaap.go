package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"scraper/db"
	"scraper/utils"
)

type Jaap struct {
	Hunter *colly.Collector
	Queue  *utils.Queue
	Db     *db.ApartmentDb
}


func (j *Jaap) Visit() {
	domain := "www.jaap.nl"
	minPrice := 1000
	maxPrice := 2000
	garden := "tuin"
	minArea := "50+-woonopp"
	name := "Jaap"
	maxPages := 5
	currentPage := 1
	searchQuery := fmt.Sprintf("/huurhuizen/noord+holland/groot-amsterdam/amsterdam/%d-%d/%s/%s",minPrice, maxPrice, minArea, garden)
	url := fmt.Sprintf("https://%s%s", domain, searchQuery )
	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	j.Hunter.CacheDir = "./_hunter/jaap"

	j.Hunter.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36"

	j.Hunter.AllowedDomains = []string{domain}
	j.Hunter.OnHTML("div.property-list", func(e *colly.HTMLElement) {
		e.ForEach("div.property ", func(_ int, element *colly.HTMLElement) {

			url := element.ChildAttr("a.property-inner", "href")
			location := element.ChildText("a.property-inner > div.property-info > div > h2.property-address-street")
			zipCode := element.ChildText("a.property-inner > div.property-info > div > h2.property-address-zipcity")
			price := element.ChildText("a.property-inner > div.property-info > div > div.price-info > div.property-price")
			apartment := utils.PotentialApartment{
				URL:      url,
				Rent:     price,
				Location: location,
				ZipCode:  zipCode,
			}
			if !j.Db.CheckApartmentExists(location) {
				err := j.Db.AddApartment(&apartment)
				if err == nil {
					j.Queue.Channel <- apartment
				}
			}
		})
	})
	j.Hunter.OnRequest(func(r *colly.Request) {
		log.Println(name, " Visiting...", r.URL)

		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
		r.Headers.Set("Accept", "text/html")
		r.Headers.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	})
	// find the next page
	j.Hunter.OnHTML("a.navigation-button", func(e *colly.HTMLElement) {

		if e.Attr("rel") != "next" {
			return
		}
		if currentPage > maxPages {
			return
		}
		url := e.Attr("href")
		fullUrl := fmt.Sprintf("https://%s%s", domain, url)
		j.Hunter.Visit(fullUrl)
		currentPage += 1
	})

	j.Hunter.OnError(func(response *colly.Response, err error) {
		log.Println(string(response.Body))
		log.Println(response.StatusCode)
		log.Println(err)
	})

	j.Hunter.Visit(url)
}