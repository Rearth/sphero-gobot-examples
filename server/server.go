package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"gobot.io/x/gobot"
)

//BotData contains references to the robots data
type BotData struct {
	SPRK      *Sprkbot
	Robot     *gobot.Robot
	active    bool
	collision bool
}

func (b *BotData) connect() {
	fmt.Println("initing the robot")
	b.Robot, b.SPRK = Create()
	fmt.Println("created robot")
	time.Sleep(150 * time.Millisecond)
	go b.Robot.Start()
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("started robot")
	b.active = true
	b.SPRK.collisionCallback = func() {
		b.collision = true
		fmt.Println("got collision callback")
	}

}

func main() {

	http.Handle("/", http.FileServer(http.Dir("./static")))
	b := BotData{}
	b.active = false
	b.collision = false

	//coordinates clicked
	http.HandleFunc("/event/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "localhost:8081", 301)
		x, _ := strconv.Atoi(strings.Split(r.URL.RawQuery, ",")[0])
		y, _ := strconv.Atoi(strings.Split(r.URL.RawQuery, ",")[1])
		q, _ := strconv.Atoi(strings.Split(r.URL.RawQuery, ",")[2])
		fmt.Printf("got coordinats: %d:%d\n", x, y)
		if b.active {
			if q == 1 {
				b.SPRK.addToQueue(x, y)
			} else {
				b.SPRK.GoToPoint(x, y)
				b.SPRK.clearQueue()
			}
		}

	})

	//start clicked
	http.HandleFunc("/start/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "localhost:8081", 301)
		fmt.Println("got start request")
		go b.connect()

	})

	//home clicked
	http.HandleFunc("/home/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "localhost:8081", 301)
		fmt.Println("got home request")
		if b.active {
			go b.SPRK.home()
		}

	})

	//boost clicked
	http.HandleFunc("/boost/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "localhost:8081", 301)
		fmt.Println("got boost request")
		if b.active {
			b.SPRK.boosting = true
			gobot.After(1200*time.Millisecond, func() {
				b.SPRK.boosting = false
			})
		}

	})

	//ajax position request
	http.HandleFunc("/api/position/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("got position request")
		if b.SPRK == nil || !b.active {
			w.Write([]byte("invalid"))
			return
		}
		x := strconv.Itoa(int(b.SPRK.getPosition().X))
		y := strconv.Itoa(int(b.SPRK.getPosition().Y))
		h := strconv.Itoa(int(b.SPRK.heading))
		pos := x + ":" + y + ":" + h + ":"
		if b.collision {
			b.collision = false
			pos += "1"
		} else {
			pos += "0"
		}
		//fmt.Println(pos)
		w.Write([]byte(pos))

	})

	//actually stop the everything when pressing ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for a := range c {
			time.Sleep(200 * time.Millisecond)
			fmt.Println(a)
			os.Exit(0)
		}
	}()

	fmt.Println("starting...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
