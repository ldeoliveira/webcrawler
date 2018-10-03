package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/antchfx/htmlquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
)


type Company struct {
	CompanyName string `json:"companyName"`
	StockName string   `json:"stockName"`
	MarketValue int64  `json:"marketValue"`
	Oscillation string `json:"oscillation"`
	Error error		   `json:"-"`
}

var db *sql.DB

const dataSourceName  = "user:password@tcp(db:3306)/companies"
const dbDriver        = "mysql"
const baseUrl         = "https://www.fundamentus.com.br/"
const homePage        = baseUrl + "detalhes.php"

func main() {

	initDb()
	defer db.Close()

	urls := buildUrlList()

	largestMarketValueCompanies := findMostValuableCompanies(urls, 10)

	persistCompanies(largestMarketValueCompanies)

	initEndpoint()
}

func initEndpoint() {
	router := mux.NewRouter()
	router.HandleFunc("/companies", GetCompanies).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}


func initDb() {
	var err error
	db, err = sql.Open(dbDriver, dataSourceName)
	if err != nil {
		panic(err)
	}
	_, errDel := db.Exec("delete from company")

	if errDel != nil {
		log.Println("could not clean table campanies...")
	}
}

/**
	Find the K most valuable companies, where K is given by the parameter numberOfCompanies
 */

func findMostValuableCompanies(urls []string, numberOfCompanies int) (largestMarketValueCompanies []Company) {

	for _, url := range urls {
		company := findCompanyInfo(url)
		if company.Error != nil {
			continue
		}
		if len(largestMarketValueCompanies) < numberOfCompanies {
			largestMarketValueCompanies = append(largestMarketValueCompanies, company)
		} else {
			minIndex, minErr := min(largestMarketValueCompanies)
			if minErr != nil {
				continue
			}
			if company.MarketValue > largestMarketValueCompanies[minIndex].MarketValue {
				largestMarketValueCompanies[minIndex] = company
			}
		}
	}

	return
}

/**
	Insert the list of companies in the DB
 */
func persistCompanies(companies []Company) {

	for _, company := range companies {

		stmt, errStatement := db.Prepare("INSERT INTO company (company_name, stock_name, market_value, oscillation) VALUES (?, ?, ?, ?)")
		if errStatement != nil {
			log.Println(errStatement)
			continue
		}
		_, errorExec := stmt.Exec(company.CompanyName, company.StockName, company.MarketValue, company.Oscillation)

		if errorExec != nil {
			log.Println(errorExec)
		} else {
			log.Printf("Company %s inserted successfully", company.CompanyName)
		}

		stmt.Close()
	}
}

/*
	Get the companies present in the db
 */
func GetCompanies(w http.ResponseWriter, r *http.Request) {

	var companies []Company

	rows, errorQuery := db.Query("SELECT company_name, stock_name, oscillation, market_value FROM company")

	if errorQuery != nil {
		panic(errorQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			companyName string
			stockName string
			oscillation string
			marketValue int64
		)
		err := rows.Scan(&companyName, &stockName, &oscillation, &marketValue)
		if err != nil {
			log.Fatal(err)
		}
		company := Company{CompanyName: companyName, StockName: stockName, Oscillation: oscillation, MarketValue: marketValue}
		companies = append(companies, company)
	}
	json.NewEncoder(w).Encode(companies)
}

/**
	Load the homepage and build the list of urls to crawl
 */
func buildUrlList() (urls []string) {

	doc, err := htmlquery.LoadURL(homePage)
	if err != nil {
		panic(err)
	}
	nodes := htmlquery.Find(doc, "//tr")[1:]
	for _, n := range nodes {
		a := htmlquery.FindOne(n, "//a")
		ref := htmlquery.SelectAttr(a, "href")
		if strings.HasPrefix(ref, "detalhes") {
			urls = append(urls, baseUrl+ref)
		}
	}
	return
}

/**
	Find the least valuable company in an array of companies and return its index
 */
func min(companies []Company) (minIndex int, e error) {

	if len(companies) == 0 {
		minIndex = -1
		e = errors.New("empty slice")
		return
	}
	minIndex = 0
	for i, c := range companies {
		if c.MarketValue < companies[minIndex].MarketValue {
			minIndex = i
		}
	}
	return
}

/**
	Assemble company information given the company page url
 */
func findCompanyInfo(baseUrl string) (company Company) {
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

	return
}


/**
	Test min() function
 */
func testMin(t *testing.T) {
	var companies []Company
	expectedIndex := 0
	companies = append(companies, Company{MarketValue: 1})
	companies = append(companies, Company{MarketValue: 2})
	companies = append(companies, Company{MarketValue: 3})
	companies = append(companies, Company{MarketValue: 4})
	minIndex, e := min(companies)

	if e != nil {
		t.Error("There was an error executing the min() function", e)
	}

	if minIndex != expectedIndex {
		t.Errorf("Min index was %d but was expecting %d", minIndex, expectedIndex)
	}

}

