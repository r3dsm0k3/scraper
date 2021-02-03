package scraper

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
	"github.com/r3dsm0k3/scraper/db"
	"github.com/r3dsm0k3/scraper/utils"
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
	maxPrice := 2000
	terrace := "roof-terrace"
	garden := "garden"
	minArea := "50m2"
	name := "Pararius"
	maxPages := 5
	currentPage := 1
	searchQuery := fmt.Sprintf("/apartments/%s/%d-%d/%s", location, minPrice, maxPrice, minArea)
	urlTerrace := fmt.Sprintf("https://%s%s/%s", domain, searchQuery, terrace)
	urlGarden := fmt.Sprintf("https://%s%s/%s", domain, searchQuery, garden)
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
			Price:    price,
			Location: location,
			ZipCode:  zipCode,
		}
		if !p.Db.CheckApartmentExists(location) {
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
	// find the next page
	p.Hunter.OnHTML("a.pagination__link.pagination__link--next", func(e *colly.HTMLElement) {
		if currentPage > maxPages {
			return
		}

		url := e.Attr("href")
		fullUrl := fmt.Sprintf("https://%s%s", domain, url)
		p.Hunter.Visit(fullUrl)
		currentPage += 1
	})

	p.Hunter.OnError(func(response *colly.Response, err error) {
		log.Println(string(response.Body))
		log.Println(response.StatusCode)
		log.Println(err)
	})

	p.Hunter.Visit(urlTerrace)
	p.Hunter.Visit(urlGarden)
}
