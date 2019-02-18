package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

func cirlce() {

	//direction which the bot is currently facing
	//heading := uint16(0)

	bleName := "SK-3C50"

	if len(os.Args) > 1 {
		bleName = os.Args[1]
	}

	bleAdaptor := ble.NewClientAdaptor(bleName)
	sprk := sprkplus.NewDriver(bleAdaptor)

	work := func() {

		fmt.Println("starting")

		stepSize := 16

		for i := 0; i <= 360*2; i += stepSize {
			sprk.Roll(255, uint16(i%360))
			time.Sleep(63 * 4 * time.Millisecond)
			fmt.Printf("%d\n", i)
		}

		sprk.Stop()
		fmt.Println("done")

	}

	robot := gobot.NewRobot("sprkie",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{sprk},
		work,
	)

	robot.Start()
}
