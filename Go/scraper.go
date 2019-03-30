package scraper

import (
	"bytes"
	"errors"
	"github.com/gocolly/colly"  // package for scraping html content
	"golang.org/x/net/html"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type reddit struct {
	Url string
	Title template.HTML  // format as html so characters are not escaped
}

type player struct {
	Name, Nation, Position, Age string
}

type stats struct {
	Country, League, Record, Home, Points, Goals, Position, Diff string
}

// takes in a reddit url, parses the page's posts, and creates a slice of
// a reddit struct with their title and url. Takes the slice and cuts it
// down to no more than 10 results, and passes it into a function that parses the
// content into raw html. This raw html is then returned as template.html
func GetReddit(url string) template.HTML{  // return html template template.HTML
	redditString := "https://old.reddit.com"  // used to make relative reddit url an absolute url
	results := []reddit{}  // slice for reddit results

	// create default collector
	c := colly.NewCollector(
		colly.Async(true),
	)

	// On every a element which has .top-matter attribute call callback.
	// This func is unique to the div that holds all information about a
	// post on common reddit pages (/, /top, /new, etc.)
	c.OnHTML(".top-matter", func(e *colly.HTMLElement) {
		temp := reddit{}
		temp.Url = e.ChildAttr("a[data-event-action=\"title\"]", "href")
		// if url is relative, make it absolute
		if !strings.Contains(temp.Url,"https"){
			temp.Url = redditString + temp.Url
		}
		temp.Title = template.HTML(e.ChildText("a[data-event-action=\"title\"]"))
		// if title is too long, shorten it and make it clear title is longer
		if len(temp.Title) > 50{
			temp.Title = temp.Title[:47] + "..."
		}
		results = append(results, temp)
	})

	// On every a element which has .search-result-header attribute call callback.
	// This func is unique to the div that holds all information about a
	// post on the reddit search results
	c.OnHTML(".search-result-header", func(e *colly.HTMLElement) {
		temp := reddit{}
		temp.Url = e.ChildAttr("a", "href")
		// if url is relative, make it absolute
		if !strings.Contains(temp.Url,"https"){
			temp.Url = redditString + temp.Url
		}
		temp.Title = template.HTML(e.ChildText("a"))
		// if title is too long, shorten it and make it clear title is longer
		if len(temp.Title) > 50{
			temp.Title = temp.Title[:47] + "..."
		}
		results = append(results, temp)
	})

	c.Visit(url)  // have collector visit site

	c.Wait()  // let collector jobs finish

	// popular sub-reddit where first 2 posts are typically , daily, pinned posts,
	// only take 10 highest post after pinned
	if url == "https://old.reddit.com/r/soccer/"{
		results = results[2:12]
	}

	// only want a max of 10 results
	if len(results) > 10{
		results = results[:10]
	}

	// pass slice into html parser and return results
	return parseReddit(results)
}

// takes in a slice of reddit structs, parses the structs into a table
// in raw html, and returns the resulting html
func parseReddit(results []reddit) template.HTML{
	var parsedHTML bytes.Buffer  // where html is stored after execution

	// html we will parse over, fill, and return
	var htmlTemplate = `{{range .}}<tr>
<td><a href="{{.Url}}" target="_blank">{{.Title}}</a></td>
</tr>
{{end}}`

	// create the template and parse in the html that will be filled
	t, err := template.New("t").Parse(htmlTemplate)
	if err != nil {
		log.Print(err)
	}

	// execute reddit structs into template and store results in bytes buffer
	err = t.Execute(&parsedHTML, results)
	if err != nil {
		log.Print(err)
	}

	// change bytes buffer to string and then return as template.html
	return template.HTML(parsedHTML.String())
}

// takes in the soccer reference website (https://fbref.com/en/) team
// url for scraping and gets team's roster information. Calls onto
// helper function that parses the slice of player results and produces
// an html table that is then returned as a template.html
func GetSoccerRefRoster(url string) template.HTML{
	players := []player{}  // slice for players

	// create default collector
	c := colly.NewCollector(
		colly.Async(true),
	)

	// find the "stats_player" table and filter through each row element pulling player
	// information from the table
	c.OnHTML("table[id=\"stats_player\"]", func(e *colly.HTMLElement) {
		e.ForEach("tr",  func(_ int, elem *colly.HTMLElement){
			temp := player{}
			temp.Name = elem.ChildText("th[data-stat=\"player\"]")
			temp.Nation = elem.ChildText("td[data-stat=\"nationality\"]")
			temp.Position = elem.ChildText("td[data-stat=\"position\"]")
			temp.Age = elem.ChildText("td[data-stat=\"age\"]")
			// if nation isn't empty, select nation short-name
			if temp.Nation != ""{
				temp.Nation = strings.Split(temp.Nation, " ")[1]
			}
			players = append(players, temp)
		})
	})

	c.Visit(url)  // have collector visit site

	c.Wait()  // let collector jobs finish

	// remove site's table header rows, and totals row
	players = players[2:len(players) - 1]

	// pass slice into html parser and return results
	return parseSoccerRef(players)
}

// takes in a slice of player structs, parses the structs into a table
// in raw html, and returns the resulting html
func parseSoccerRef (players []player) template.HTML{
	var parsedHTML bytes.Buffer  // where html is stored after execution

	// html we will parse over and return
	var htmlTemplate = `{{range .}}<tr>
	<td>{{.Name}}</td>
	<td>{{.Nation}}</td>
	<td>{{.Position}}</td>
	<td>{{.Age}}</td>
</tr>
{{end}}`

	// create the template and parse in the html that will filled in
	t, err := template.New("t").Parse(htmlTemplate)
	if err != nil {
		log.Print(err)
	}

	// execute player structs into template and store results in bytes buffer
	err = t.Execute(&parsedHTML, players)
	if err != nil {
		log.Print(err)
	}

	// change bytes buffer to string and then return as template.html
	return template.HTML(parsedHTML.String())
}

// recursively parses html and returns all content that
// is within the tbody element
func getBody(doc *html.Node) (*html.Node, error) {
	var b *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		// if node is html and tbody, get content within tbody
		if n.Type == html.ElementNode && n.Data == "tbody" {
			b = n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if b != nil {
		return b, nil
	}
	return nil, errors.New("Missing <tbody> in the node tree")
}

// takes in html node and returns string format
func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

// takes in bleacher report url to team's schedule, gets the html content,
// parses it for tbody html, and returns it as a template.html
func GetBRSchedule(url string) template.HTML{
	resp, err := http.Get(url)  // get url content
	if err != nil {
		log.Println(err)  // log errors for troubleshooting
	}

	body, err := ioutil.ReadAll(resp.Body)  // parse response body
	if err != nil {
		log.Println(err)
	}

	res := string(body)

	doc, err := html.Parse(strings.NewReader(res))  // parse reader into html
	if err != nil {
		log.Println(err)
	}

	bn, err := getBody(doc)  // call onto helper function to get table body
	if err != nil {
		log.Println(err)
	}

	bod := renderNode(bn)  // call helper function to change html tbody node into string

	bod = bod[7:len(bod) - 8]  // remove tbody html tags

	return template.HTML(bod)
}

// takes in the soccer reference website (https://fbref.com/en/) team
// url and gets the team statistics from the top of the page. Takes
// the statistics and stores them in a stats struct and returns it.
func GetSoccerRefStats(url string) stats{
	values := stats{}  // struct of stats to fill and return

	// create default collector
	c := colly.NewCollector(
		colly.Async(true),
	)

	// find the "stats_player" table and filter through each row element pulling player
	// information from the table
	c.OnHTML("div[data-template=\"Partials/Teams/Summary\"]", func(e *colly.HTMLElement) {
		e.ForEach("p",  func(line int, elem *colly.HTMLElement){
			if line == 0 {
				values.Country = strings.Split(elem.Text, " ")[2]
			} else if line == 1 {
				currLine := strings.Split(elem.Text, ",")
				values.Record = strings.TrimSpace(strings.Split(currLine[0], ":")[1])
				values.Points = strings.TrimSpace(strings.Split(currLine[1], "points")[0])
				values.Position = strings.TrimSpace(strings.Split(currLine[2], "in")[0])
				values.League = strings.TrimSpace(elem.ChildText("a"))
			} else if line == 2 {
				home := strings.Replace(strings.Split(elem.Text, " ")[7], "(", "", -1)
				values.Home = strings.Replace(home, ")", "", -1)
			} else if line == 3 {
				currLine := strings.Split(elem.Text, ",")
				values.Goals = strings.Split(currLine[0], " ")[1]
				values.Diff = strings.TrimSpace(strings.Split(currLine[2], " ")[1])
			}
		})
	})

	c.Visit(url)  // have collector visit site

	c.Wait()  // let collector jobs finish

	return values
}
