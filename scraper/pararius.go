package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"scraper/db"
	"scraper/utils"
	"log"
)

type Pararius struct {
	Hunter *colly.Collector
	Queue  *utils.Queue
	Db     *db.ApartmentDb
}


func (p *Pararius) Visit() {
	domain := "www.pararius.com"
	location := "amsterdam"
	minPrice := 1000
	maxPrice := 1750
	terrace := "roof-terrace"
	garden := "garden"
	minArea := "50m2"
	name := "Pararius"
	searchQuery := fmt.Sprintf("/apartments/%s/%d-%d/%s/%s/%s",location, minPrice, maxPrice,minArea, garden, terrace  )
	url := fmt.Sprintf("https://%s%s", domain, searchQuery)

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	p.Hunter.CacheDir = "./_hunter/pararius"

	p.Hunter.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36"

	p.Hunter.AllowedDomains = []string{domain}
	p.Hunter.OnHTML("ul.search-list > li.search-list__item--listing", func(e *colly.HTMLElement) {

		link := e.ChildAttr("div.listing-search-item__depiction > a.listing-search-item__link", "href")
		fullUrl := fmt.Sprintf("https://%s%s", domain, link)
		location := e.ChildText("h2.listing-search-item__title > a.listing-search-item__link")
		price := e.ChildText("div.listing-search-item__price-status > span.listing-search-item__price")
		zipCode := e.ChildText("h2.listing-search-item__title > div.listing-search-item__location")

		apartment := utils.PotentialApartment{
			URL:      fullUrl,
			Rent:     price,
			Location: location,
			ZipCode:  zipCode,
		}
		if !p.Db.CheckApartmentExists(name) {
			err := p.Db.AddApartment(&apartment)
			if err == nil {
				p.Queue.Channel <- apartment
			}
		}
	})
	p.Hunter.OnRequest(func(r *colly.Request) {
		log.Println(name, " Visiting...", r.URL)

		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
		r.Headers.Set("Accept", "text/html")
		r.Headers.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	})
	p.Hunter.OnError(func(response *colly.Response, err error) {
		log.Println(string(response.Body))
		log.Println(response.StatusCode)
		log.Println(err)
	})

	//// enable this for debugging the response
	//p.Hunter.OnResponse(func(response *colly.Response) {
	//	fmt.Println(string(response.Body))
	//})
	p.Hunter.Visit(url)
}