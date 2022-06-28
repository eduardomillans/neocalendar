package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

type Assignment struct {
	title  string
	course string
	date   time.Time
	url    string
}

func main() {
	assignaments := make([]Assignment, 0)

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatalln("could not load env")
	}

	// First, we need to install dirvers and browsers
	if err := playwright.Install(); err != nil {
		log.Fatalf("could not install drivers and browser: %v", err)
	}

	// Lauch playwright
	pw, err := playwright.Run()

	handlePlaywrightError(&pw, nil, "sorry, could not lauch pw", err)

	// Create a browser instance
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})

	handlePlaywrightError(&pw, &browser, "sorry, could not lauch browser", err)

	// Create a browser context
	context, err := browser.NewContext()

	handlePlaywrightError(&pw, &browser, "sorry, could not create browser context", err)

	// Create a page
	page, err := context.NewPage()

	handlePlaywrightError(&pw, &browser, "sorry, could not create page", err)

	// Start automatitation
	host := "https://unid.neolms.com"
	email := os.Getenv("email")
	password := os.Getenv("password")

	page.Goto(host)

	// Auth process
	err = page.Click(".loginHolder a")
	handlePlaywrightError(&pw, &browser, `[click error]: ".loginHolder a"`, err)

	err = page.Click("button#office365_sso_btn")
	handlePlaywrightError(&pw, &browser, `[click error]: "button#office365_sso_btn"`, err)

	err = page.Fill("input[type=email]", email)
	handlePlaywrightError(&pw, &browser, `[fill error]: "input[type=email]"`, err)

	err = page.Click("input[type=submit]")
	handlePlaywrightError(&pw, &browser, `[click error]: "input[type=submit]"`, err)

	err = page.Fill("input[type=password]", password)
	handlePlaywrightError(&pw, &browser, `[fill error]: "input[type=password]"`, err)

	err = page.Click("input[type=submit]")
	handlePlaywrightError(&pw, &browser, `[click error]: "input[type=submit]"`, err)

	err = page.Click("input[type=button]")
	handlePlaywrightError(&pw, &browser, `[click error]: "input[type=button]"`, err)

	// Go to calendar view
	err = page.Click(".quickLinks a[title=Calendario]")
	handlePlaywrightError(&pw, &browser, `[click error]: ".quickLinks a[title=Calendario]"`, err)

	tds, err := page.Locator("table.calendar tbody td")
	handlePlaywrightError(&pw, &browser, `[locator error]: "table.calendar tbody td"`, err)

	// With `GetAttribute` force to pw to wait for the elements
	_, err = tds.GetAttribute("data-add-event")
	handlePlaywrightError(&pw, &browser, `[locator error]: "could not wait for tds"`, err)

	tdsCount, _ := tds.Count()

	for i := 0; i < tdsCount; i++ {
		td, _ := tds.Nth(i)

		// Try to find anchor element with the assignment information
		anchors, _ := td.Locator("a.general_event")

		anchorsCount, _ := anchors.Count()

		for j := 0; j < anchorsCount; j++ {
			var title, course, url string
			var date time.Time

			anchor, _ := anchors.Nth(j)

			raw, _ := anchor.GetAttribute("onmouseover")

			rex := regexp.MustCompile(`\(([^)]+)\)`)
			out := rex.FindString(raw)

			data := strings.Split(out, ",")

			title = data[1]
			title = strings.ReplaceAll(title, "'", "")

			course = data[3]
			course = strings.ReplaceAll(course, "'", "")
			course = strings.ReplaceAll(course, "\\", "")
			course = strip.StripTags(course)
			course = strings.Split(course, " ")[3]

			attr, _ := td.GetAttribute("data-add-event")
			attr = strings.ReplaceAll(attr, ",", "-")

			date, _ = time.Parse("2006-1-2", attr)

			url, _ = anchor.GetAttribute("href")
			url = fmt.Sprintf("%s%s", host, url)

			assignaments = append(assignaments, Assignment{title, course, date, url})
		}
	}

	// Show assignments in table format
	prettyPrint(&assignaments)

	browser.Close()

	pw.Stop()
}

func handlePlaywrightError(pw **playwright.Playwright, browser *playwright.Browser, message string, err error) {
	if err != nil {
		if browser != nil {
			(*browser).Close()
		}

		(*pw).Stop()

		log.Fatalln(message)
	}
}

func prettyPrint(assignments *[]Assignment) {
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Title"},
			{Align: simpletable.AlignLeft, Text: "Course"},
			{Align: simpletable.AlignLeft, Text: "Date"},
			{Align: simpletable.AlignLeft, Text: "URL"},
		},
	}

	for _, assignment := range *assignments {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: assignment.title},
			{Align: simpletable.AlignLeft, Text: assignment.course},
			{Align: simpletable.AlignLeft, Text: assignment.date.Format("2 January 2006")},
			{Align: simpletable.AlignLeft, Text: assignment.url},
		}

		table.Body.Cells = append(table.Body.Cells, r)
	}

	table.SetStyle(simpletable.StyleDefault)

	fmt.Println(table.String())
}
