# Local Fan Page
Local fan page scrapes content from various sports/information sites and consolidates the data into a web page that is locally hosted and accessible. This is done by producing a binary that contains all static content and code to generate the dynamic content.  The first version is built around my favorite soccer team - Juventus, but I do plan to create implementations for other teams and sports.

![Juventus Fan Page](https://github.com/dwright20/local-fan-page/blob/master/Examples/juve-fan-page.png)

## Code Summary
- Created a [scraper](https://github.com/dwright20/local-fan-page/blob/master/Go/scraper.go) package that is called to scrape given web page URLs and return the results. 
- [Main.go](https://github.com/dwright20/local-fan-page/blob/master/Go/) is tailored to a specific team and passes resources into scraper functions to produce content for HTML templates that are served to the client's browser via Mux.
- All [HTML](https://github.com/dwright20/local-fan-page/blob/master/HTML) must be tailored to a specific team as well.  Provided code to show how data is being passed. 
### Key Packages/Examples Used
- [go.rice](https://github.com/GeertJohan/go.rice) - embedding static content
- [Mux](https://github.com/gorilla/mux) - serving content in browser
- [Colly](https://github.com/gocolly/colly) - web scraping
- [HTML](https://godoc.org/golang.org/x/net/html) - HTML parsing
- [Open Function](https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang) - automate opening browser to webpage
## Sites Used
- [Bleacher Report](https://bleacherreport.com/)
- [ESPN](http://www.espn.com/)
- [Reddit](https://www.reddit.com/)
- [FB Reference](https://fbref.com/)
