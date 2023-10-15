package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/edgarsbalodis/adbuddy/internal/utils"
	"github.com/gocolly/colly/v2"
)

type Residence struct {
	Url         string  `json:"url"`
	Description string  `json:"description"`
	Address     string  `json:"address"`
	Area        uint16  `json:"area"`
	Floors      uint8   `json:"floors"`
	Rooms       float32 `json:"rooms"`
	Land        uint16  `json:"land"`
	Price       float32 `json:"price"`
}
type PropertyFilter struct {
	Type            string // apartment|residence|car
	Region          string // region/city for example: Riga || Riga region
	Subregion       string // sub region for example: Center || Babites pagasts todo: implement []string for multiple subregions
	Price           string // maximum price for property
	NumOfRooms      string // minimum amount of rooms
	TransactionType string // transaction type for property sell/rent
	Timestamp       string // the last time the filter was applied
}

// todo: rename to NewFilter
func NewEmptyFilter(datatype string) interface{} {
	switch datatype {
	case "flat", "residence":
		return &PropertyFilter{}
	default:
		return nil
	}
}

func ScrapeDataNew(filter interface{}) []string {
	// Instantiate default collector
	c := colly.NewCollector()

	var (
		shouldStop  bool   // based on timestamp
		initialPage string // if initialPage is "", that means there is no filters or problem with filter, so return empty slice
		// adType          string
		priceFilter string
		roomFilter  string
		// transactionType string
		timestamp string
		rf        float32 // room filter to float32
		nextHref  string
		urls      []string
		subr      string
	)
	// Create another collector to scrape ad details
	detailCollector := c.Clone()

	// creates url based on filter
	baseUrl := "https://www.ss.com"
	if filter != nil {
		switch f := filter.(type) {
		case *PropertyFilter:
			initialPage += "/lv/real-estate/"
			if f.Type == "residence" {
				initialPage += "homes-summer-residences/"
			} else {
				initialPage += "flats/"
			}
			// adType = f.Type
			initialPage += f.Region + "/"
			subr = f.Subregion
			initialPage += f.Subregion + "/"
			if f.TransactionType == "buy" {
				initialPage += "sell/"
			} else {
				initialPage += "hand_over/"
			}
			// transactionType = f.TransactionType
			priceFilter = f.Price
			roomFilter = f.NumOfRooms
			timestamp = f.Timestamp
		}
	}
	rf, _ = utils.StringToFloat32(roomFilter)
	pf, _ := utils.StringToFloat32(priceFilter)

	// find pages
	// structure in ss.com
	// 		1 page is like initialPage
	// 		next page is initalPage/pageX.html
	// 		if page is last page, then next value is like initial value
	c.OnHTML("form#filter_frm > div.td2 a", func(h *colly.HTMLElement) {
		nextHref = h.Attr("href")
		// Check if nextHref is not empty
		if nextHref == "" {
			shouldStop = true
			return
			// Next href exists, do something
		}
	})

	c.OnScraped(func(r *colly.Response) {
		if shouldStop || nextHref == "" {
			return
		}
		if nextHref != initialPage {
			c.Visit(baseUrl + nextHref)
		} else {
			return
		}
	})

	// data to scrape
	// get all urls where matches price filter (price column is always last)
	// this data is from base catalog
	c.OnHTML("form#filter_frm > :nth-child(3) > tbody > tr[id^='tr_']", func(h *colly.HTMLElement) {
		if shouldStop {
			return
		}

		link := h.ChildAttr("td:nth-child(3) a", "href")
		price := h.ChildText("td:last-child")

		// clean everything after €
		parts := strings.Split(price, "€")
		price = strings.TrimSpace(parts[0])
		// remove comma
		price = strings.ReplaceAll(price, ",", "")

		priceFloat, err := strconv.ParseFloat(price, 32)
		if err != nil {
			fmt.Printf("Error converting price: %v\n", err)
			return
		}

		// if price filter matches request then go to next step and visit details in ad
		if float32(priceFloat) <= pf {
			url := baseUrl + link
			if url == baseUrl {
				return
			}
			detailCollector.Visit(url)
		} else {
			return
		}
	})
	// var dateFromAd string
	// get date from ad
	// I can refactor this, so I can get all information from this detail collector
	detailCollector.OnHTML("table#page_main > tbody > tr:nth-child(2) > td > table > tbody > tr:nth-child(2) > td:nth-child(2)", func(e *colly.HTMLElement) {
		date := e.Text
		t := formatDate(date)
		// TODO: format properly
		layout := "02.01.2006 15:04 Europe/Riga"
		// Load the "Europe/Riga" time zone
		rigaLocation, err := time.LoadLocation("Europe/Riga")
		if err != nil {
			fmt.Println("Error loading location:", err)
		}
		// // Parse the input string into a time.Time object
		// t, err := time.Parse(layout, dateFromAd)
		// if err != nil {
		// 	fmt.Println("Error parsing time:", err)
		// 	return
		// }

		// 02.10.2023 15:00  ad date
		// 02.10.2023 14:00 last timestamp
		// 01.10.2023 pirms 02.10.2023
		lastUpdatedTime, _ := time.ParseInLocation(layout, timestamp+" Europe/Riga", rigaLocation)
		// fmt.Printf("Time from ad: %v \n Time from filter: %v\n\n", t, lastUpdatedTime)
		if t.Before(lastUpdatedTime) {
			shouldStop = true
		}
	})

	// Get all Details about ad (right now, just rooms)
	detailCollector.OnHTML("table.options_list > tbody > tr > td > table > tbody > tr", func(e *colly.HTMLElement) {
		// fmt.Printf("should stop: %v\n", shouldStop)
		if shouldStop {
			return
		}
		key := e.ChildText("td.ads_opt_name") // "Istabas"
		value := e.ChildText("td.ads_opt")    // "5"

		// Clean up the key and value by removing unnecessary spaces and colons
		key = strings.TrimSpace(strings.TrimSuffix(key, ":"))
		value = strings.TrimSpace(value)

		if key == "Istabas" {
			rooms, err := strconv.ParseFloat(value, 32)
			if err != nil {
				fmt.Printf("Error converting rooms: %v\n", err)
				return
			}

			if float32(rooms) >= rf {
				url := e.Request.URL.String()
				// Append the URL to urls
				urls = append(urls, url)
			}

			return
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(baseUrl + initialPage)
	fmt.Printf("Scraped from %v: %v\n", subr, urls)
	return urls
}

func formatDate(date string) time.Time {
	// remove Datums: string and andd timezone for latvia
	dateFromAd := strings.SplitN(date, "Datums: ", 2)[1] + " Europe/Riga"

	// Load the "Europe/Riga" time zone
	rigaLocation, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		fmt.Println("Error loading location:", err)
	}
	// updated layout for right timezone
	layout := "02.01.2006 15:04 Europe/Riga"

	// Parse the input string into a time.Time object
	t, err := time.ParseInLocation(layout, dateFromAd, rigaLocation)

	// t, err := time.Parse(layout, dateFromAd)
	if err != nil {
		fmt.Println("Error parsing time:", err)
	}
	return t
}

// think better solution what to return
// func processResidence(el *colly.HTMLElement, priceFilter, roomFilter, transactionType string) string {
// 	url := el.ChildAttr("td:nth-child(3) a", "href")
// 	description := el.ChildText("td:nth-child(3)")
// 	address := el.ChildText("td:nth-child(4)")
// 	area := el.ChildText("td:nth-child(5)")
// 	floors := el.ChildText("td:nth-child(6)")
// 	rooms := el.ChildText("td:nth-child(7)")
// 	land := el.ChildText("td:nth-child(8)")
// 	price := el.ChildText("td:nth-child(9)")
// 	baseUrl := "https://www.ss.com"

// 	areaInt, err := strconv.ParseUint(area, 10, 16)
// 	if err != nil {
// 		fmt.Printf("Error converting area: %v\n", err)
// 		return ""
// 	}

// 	floorsInt, err := strconv.ParseUint(floors, 10, 8)
// 	if err != nil {
// 		fmt.Printf("Error converting floors: %v\n", err)
// 		return ""
// 	}

// 	roomsInt, err := strconv.ParseFloat(rooms, 32)
// 	if err != nil {
// 		fmt.Printf("Error converting rooms: %v\n", err)
// 		return ""
// 	}

// 	land = strings.Replace(land, " m²", "", -1)
// 	landInt, err := strconv.ParseUint(land, 10, 16)
// 	if err != nil {
// 		fmt.Printf("Error converting land: %v\n", err)
// 		return ""
// 	}

// 	if transactionType == "sell" {
// 		price = strings.Replace(price, "  €", "", -1)
// 	} else {
// 		parts := strings.Split(price, "  €")
// 		price = parts[0]
// 	}
// 	price = strings.ReplaceAll(price, ",", "")
// 	priceFloat, err := strconv.ParseFloat(price, 32)
// 	if err != nil {
// 		fmt.Printf("Error converting price: %v\n", err)
// 		return ""
// 	}

// 	pf, _ := utils.StringToFloat32(priceFilter)
// 	rf, _ := utils.StringToFloat32(roomFilter)
// 	url = baseUrl + url
// 	residence := Residence{Url: url, Description: description, Address: address, Area: uint16(areaInt), Floors: uint8(floorsInt), Rooms: float32(roomsInt), Land: uint16(landInt), Price: float32(priceFloat)}

// 	if residence.Price <= pf && residence.Rooms >= rf {
// 		return residence.Url
// 	}
// 	return ""
// }

// func processFlat(el *colly.HTMLElement, priceFilter, roomFilter, transactionType string) string {
// 	url := el.ChildAttr("td:nth-child(3) a", "href")
// 	description := el.ChildText("td:nth-child(3)")
// 	address := el.ChildText("td:nth-child(4)")
// 	rooms := el.ChildText("td:nth-child(5)")
// 	area := el.ChildText("td:nth-child(6)")
// 	// floor := el.ChildText("td:nth-child(7)")
// 	// land := el.ChildText("td:nth-child(9)")
// 	price := el.ChildText("td:nth-child(10)")
// 	baseUrl := "https://www.ss.com"

// 	areaInt, err := strconv.ParseUint(area, 10, 16)
// 	if err != nil {
// 		fmt.Printf("Error converting area: %v\n", err)
// 		return ""
// 	}

// 	// floorsInt, err := strconv.ParseUint(floor, 10, 8)
// 	// if err != nil {
// 	// 	fmt.Printf("Error converting floors: %v\n", err)
// 	// 	return ""
// 	// }

// 	roomsInt, err := strconv.ParseFloat(rooms, 32)
// 	if err != nil {
// 		fmt.Printf("Error converting rooms: %v\n", err)
// 		return ""
// 	}

// 	// land = strings.Replace(land, " m²", "", -1)
// 	// landInt, err := strconv.ParseUint(land, 10, 16)
// 	// if err != nil {
// 	// 	fmt.Printf("Error converting land: %v\n", err)
// 	// 	return ""
// 	// }

// 	if transactionType == "sell" {
// 		price = strings.Replace(price, "  €", "", -1)
// 	} else {
// 		parts := strings.Split(price, "€")
// 		price = strings.ReplaceAll(parts[0], " ", "")
// 	}
// 	price = strings.ReplaceAll(price, ",", "")
// 	priceFloat, err := strconv.ParseFloat(price, 32)
// 	if err != nil {
// 		fmt.Printf("Error converting price: %v\n", err)
// 		return ""
// 	}
// 	pf, _ := utils.StringToFloat32(priceFilter)
// 	rf, _ := utils.StringToFloat32(roomFilter)
// 	url = baseUrl + url
// 	residence := Residence{Url: url, Description: description, Address: address, Area: uint16(areaInt), Floors: uint8(0), Rooms: float32(roomsInt), Land: uint16(0), Price: float32(priceFloat)}

// 	if residence.Price <= pf && residence.Rooms >= rf {
// 		return residence.Url
// 	}
// 	return ""
// }
