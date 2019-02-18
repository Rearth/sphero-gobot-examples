package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

func main() {

	//bluetooth adapter
	bleAdaptor := ble.NewClientAdaptor(os.Args[1])
	//the sprk driver
	sprk := sprkplus.NewDriver(bleAdaptor)

	work := func() {

		//move straight ahead
		sprk.Roll(80, 0)

		//stop after 5 seconds
		gobot.After(5*time.Second, func() {
			fmt.Printf("done Moving\n")
			sprk.Stop()
		})

		//periodic call
		gobot.Every(2*time.Second, func() {
			//only 50 brightness instead of 255 to not annoy everyone in the office
			r := uint8(gobot.Rand(50))
			g := uint8(gobot.Rand(50))
			b := uint8(gobot.Rand(50))
			sprk.SetRGB(r, g, b)
		})
	}

	//defining the robot
	robot := gobot.NewRobot("sprkie",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{sprk},
		work,
	)

	robot.Start()
}
