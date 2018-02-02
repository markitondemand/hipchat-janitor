package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tbruyelle/hipchat-go/hipchat"
)

var (
	tokenFlag    = flag.String("token", "", "The HipChat API token.")
	urlFlag      = flag.String("url", "", "The HipChat server URL.")
	intervalFlag = flag.Int("interval", 24, "How often cleanups are attempted, in hours.")
	maxFlag      = flag.Int("max", 30, "The maximum amount of time to keep a room, in days.")
	insecureFlag = flag.Bool("insecure", false, "Skip certificate verification for HTTPS requests.")
)

func main() {

	flag.Parse()

	// validate flags
	if *tokenFlag == "" {
		log.Fatal("The 'token' flag is required for communications with HipChat.")
	}
	if *urlFlag == "" {
		log.Fatal("The 'url' flag must be set to a valid HipChat server URL.")
	}
	if *intervalFlag <= 0 {
		log.Fatal("The 'interval' flag must be set to a number greater than zero.")
	}
	if *maxFlag <= 0 {
		log.Fatal("The 'max' flag must be set to a number greater than zero.")
	}

	// ensure that we can parse the provided URL flag
	hipchatURL, err := url.Parse(*urlFlag)
	if err != nil {
		log.Fatalf("Fatal: could not parse URL flag: %v", err)
	}

	// setup health server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	http.ListenAndServe(":3000", nil)

	// setup metrics server
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3001", nil)

	// create a HipChat client
	client := hipchat.NewClient(*tokenFlag)
	client.BaseURL = hipchatURL

	// if the insecure flag is set, customize the underlying HTTP client to skip TLS verification
	if *insecureFlag {
		client.SetHTTPClient(&http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}})
	}

	// create a new ticker based on the interval configuration
	ticker := time.NewTicker(time.Duration(*intervalFlag) * time.Hour)
	log.Printf("HipChat Janitor Started. Archiving any rooms untouched for %d days every %d hours.", *maxFlag, *intervalFlag)
	for {

		// get rooms
		rooms, _, err := client.Room.List(&hipchat.RoomsListOptions{IncludePrivate: true, IncludeArchived: false})
		if err != nil {
			log.Fatalf("Fatal: Could not get rooms list from HipChat: %v", err)
		}

		// loop over each room and check for archivability
		for _, item := range rooms.Items {

			// the item ID needs to be converted to a string
			idString := strconv.Itoa(item.ID)

			// get room details
			room, _, err := client.Room.Get(idString)
			if err != nil {
				log.Printf("Error: Could not retrieve room '%s': %v", item.Name, err)
				break
			}

			// ensure that we only archive private rooms
			if room.Privacy == "private" {

				// get room statistics
				stats, _, err := client.Room.GetStatistics(idString)
				if err != nil {
					log.Printf("Error: Could not retrieve statistics for room '%s': %v", room.Name, err)
					break
				}

				// get lastactive date as a time
				lastActiveTime, err := time.Parse(time.RFC3339, stats.LastActive)
				if err != nil {
					log.Printf("Error: Could not parse LastActive time '%s' for room '%s': %s", stats.LastActive, room.Name, err)
					break
				}

				// if the last update was > max days ago, archive the room
				if time.Now().Sub(lastActiveTime).Hours()/24 >= float64(*maxFlag) {
					log.Printf("Archiving room '%s', not touched in >= %d days.", room.Name, *maxFlag)
					_, err := client.Room.Update(idString, &hipchat.UpdateRoomRequest{
						Name:          room.Name,
						IsArchived:    true,
						IsGuestAccess: room.IsGuestAccessible,
						Owner:         hipchat.ID{ID: strconv.Itoa(room.Owner.ID)},
						Privacy:       room.Privacy,
						Topic:         room.Topic,
					})
					if err != nil {
						log.Printf("Error: Could not archive room '%s': %v", idString, err)
					}
				}

			}
		}

		<-ticker.C
	}
}
