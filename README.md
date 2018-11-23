# web crawler application

## Task Description

To get the 10 most valuable companies from the fundamentus website

## How to run

From inside the project directory, execute:

`docker-compose build && docker-compose up`

After it finishes to crawl all URLs (it takes around 5min), an endpoint should be available. To get the companies information execute:

`curl -X GET http://localhost:6060/companies`

