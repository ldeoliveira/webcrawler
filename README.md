# web crawler application

## Task Description

To get the 10 most valuable companies from the fundamentus website

## How to run

From inside the project directory, execute:

`docker-compose build && docker-compose up`

After it finishes to crawl all URLs (it takes around 5min), an endpoint should be available. To get the companies information execute:

`curl -X GET http://localhost:6060/companies`

## Considerations

I first started an implementation to crawl URLs in parallel, but the Fundamentus server was not able to process concurrent requests and started to return a lot of 503 errors. That said, I decided to keep the serial implementation as primary-- but I kept the parallel functions in `companies-parallel.go` so you can take a look on what was done. I implemented only a single unit tests just to get familiar with testing in Go. This project is literally my very first lines in Go, so it may not look as idiomatic as possible.
