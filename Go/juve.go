package main

import (
	"fanProject/scraper"
	"github.com/GeertJohan/go.rice"  // package to embed web content in binary
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

var (
	// reddit sites that will be scraped for web content.
	// Format: [/r/soccer, /r/soccer top results for team, team sub-reddit top results]
	redditSites = [3]string{
		"https://old.reddit.com/r/soccer/",
		"https://old.reddit.com/r/soccer/search?q=juventus&restrict_sr=on&sort=top&t=day",
		"https://old.reddit.com/r/Juve/top/",
	}

	// team schedule from Bleacher Report url
	brSite = "https://bleacherreport.com/juventus/schedule"
	// team soccer reference url
	soccerRef = "https://fbref.com/en/squads/e0652b02/Juventus"

	// below are sites that are only present as a list of links on site and can be changed out

	// official team twitter page url
	twitter = "https://twitter.com/juventusfcen"
	// official team instagram page url
	instagram = "https://www.instagram.com/juventus/?hl=en"
	// team ESPN page url
	espn = "http://www.espn.com/soccer/team/_/id/111/"
	// official team website url
	teamSite = "https://www.juventus.com/en/"
	// team whoScored page url
	whoScored = "https://www.whoscored.com/Teams/87/Show/Italy-Juventus"

	box *rice.Box  // initialize box that will store static web content
)

type Stats struct {
	Country, League, Record, Home, Points, Goals, Position, Diff string
}

type Tables struct {
	Soccer, SoccerTop, Team, Schedule, Roster template.HTML
}

type Socials struct {
	Twitter, Instagram, Team, Espn, Who string
}

type Results struct {
	Table Tables
	Stat Stats
	Sites Socials
}

// opens the specified URL in the default browser of the user.  Taken from:
// https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// handles request for website content.  Calls onto helper
// functions to parse websites and generate necessary table
// information.
func ServeSite (w http.ResponseWriter, _ *http.Request) {
	templateData, _ := box.String("index.html")
	soccer := scraper.GetReddit(redditSites[0])  // get /r/soccer results table template
	soccerTop := scraper.GetReddit(redditSites[1])  // get /r/soccer/top team results table template
	team := scraper.GetReddit(redditSites[2])  // get /r/team results table template
	schedule := scraper.GetBRSchedule(brSite)  // get Bleacher Report schedule table template
	roster := scraper.GetSoccerRefRoster(soccerRef)  // get team roster table template

	values := scraper.GetSoccerRefStats(soccerRef)  // get team summary statistics string

	// create Stats struct for execution into html later
	statistics := Stats{
		Country: 	values.Country,
		League: 	values.League,
		Record: 	values.Record,
		Home: 		values.Home,
		Points: 	values.Points,
		Goals: 		values.Goals,
		Position: 	values.Position,
		Diff: 		values.Diff,
	}

	// create Tables struct for execution into html later
	tables := Tables{
		Soccer:soccer,
		SoccerTop:soccerTop,
		Team:team,
		Schedule:schedule,
		Roster:roster,
	}

	// create Socials struct for execution into html later
	site := Socials{
		Twitter: twitter,
		Instagram: instagram,
		Team: teamSite,
		Who: whoScored,
		Espn: espn,
	}

	// create Results struct to consolidate all structs needed for executing into one struct to be passed into html
	res := Results{
		Table: tables,
		Stat: statistics,
		Sites: site,
	}

	// create template and parse un-executed html
	t, err := template.New("t").Parse(templateData)
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	// execute Results struct into html and respond with content
	err = t.Execute(w, res)
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

// runs a go routine to open the website in default browser
// while creating a new router to listen on port 8080 and
// serve local web page content
func main() {
	go open("http://localhost:8080")
	// create box of files that will be used - pass in root directory
	box  = rice.MustFindBox("website")
	r := mux.NewRouter()  // create router
	r.HandleFunc("/", ServeSite)  // serve web content
	r.PathPrefix("/style.css").Handler(http.FileServer(box.HTTPBox()))
	r.PathPrefix("/juventus-logo.png").Handler(http.FileServer(box.HTTPBox()))
	r.PathPrefix("/juventus-favicon.png").Handler(http.FileServer(box.HTTPBox()))
	log.Fatal(http.ListenAndServe(":8080", r))  // listen on port 8080 for a request
}
