package main

import (
	"errors"
	"github.com/antchfx/htmlquery"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"strings"
	"sync"
)

func crawlParallel() {

	initDb()
	defer db.Close()

	var companies []Company

	urls := buildUrlList()

	c := make(chan Company, 3)

	mutex := &sync.Mutex{}

	go scrapNodesParallel(urls, c)

	for company := range c {

		if company.Error != nil {
			continue
		}
		mutex.Lock()
		if len(companies) < 10 {
			companies = append(companies, company)
		} else {
			minIndex, minErr := min(companies)
			if minErr != nil {
				continue
			}
			if company.MarketValue > companies[minIndex].MarketValue {
				companies[minIndex] = company
			}
		}
		mutex.Unlock()
	}

	persistCompanies(companies)


}


func scrapNodesParallel(urlToProcess []string, cchan chan Company) {

	defer close(cchan)

	var companies []chan Company

	for i, url := range urlToProcess {
		log.Printf("URL: %s", url)
		companies = append(companies, make(chan Company))
		go findCompanyInfoParallel(url, companies[i])
	}

	for i := range companies {
		for c1 := range companies[i] {
			cchan <- c1
		}
	}

}

func findCompanyInfoParallel(baseUrl string, cchan chan Company) {

	defer close(cchan)
	var company Company
	log.Printf("Crawling URL: %s", baseUrl)
	baseDoc, err := htmlquery.LoadURL(baseUrl)
	company.Error = err

	if baseDoc == nil {
		company.Error = errors.New("page has format different than expected")
		return
	}

	stockNameNode := htmlquery.FindOne(baseDoc, "//tr[1]/td[2]")
	companyNameNode := htmlquery.FindOne(baseDoc, "//tr[3]/td[2]")
	oscillationNode := htmlquery.FindOne(baseDoc, "//tr[9]/td[2]")
	marketValueNode := htmlquery.FindOne(baseDoc, "//tr[6]/td[2]")

	if stockNameNode == nil && companyNameNode == nil && oscillationNode == nil && marketValueNode == nil {
		company.Error = errors.New("page has format different than expected")
		return
	}

	marketValueStr := strings.Replace(htmlquery.InnerText(marketValueNode), ".", "", -1)
	marketValueInt, e := strconv.ParseInt(marketValueStr, 10, 64)
	if e != nil {
		company.Error = e
		return
	}

	oscillationStr := htmlquery.InnerText(oscillationNode)
	companyNameStr := htmlquery.InnerText(companyNameNode)
	stockNameStr := htmlquery.InnerText(stockNameNode)

	company.MarketValue = marketValueInt
	company.Oscillation = oscillationStr
	company.CompanyName = companyNameStr
	company.StockName = stockNameStr

	cchan <- company
}


