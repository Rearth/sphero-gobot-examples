package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

// This program enables you to move the sphero around by entering WASD. Use the name of the sphero as the first command-line parameter

type sprkbot struct {
	*sprkplus.SPRKPlusDriver
}

func (sprk *sprkbot) move(dir int) {
	speed := 100
	sprk.Roll(uint8(speed), uint16(dir))
	fmt.Printf("rolling to: %d\n", uint16(dir))
	time.Sleep(500 * time.Millisecond)
	sprk.Roll(1, uint16(dir))
	time.Sleep(150 * time.Millisecond)
}

// NewDriver creates a Driver for a Sphero SPRK+
func newDriver(a ble.BLEConnector) sprkbot {
	d := sprkplus.NewDriver(a)

	return sprkbot{
		SPRKPlusDriver: d,
	}
}

func main() {

	bleName := "SK-3C50"

	if len(os.Args) > 1 {
		bleName = os.Args[1]
	}

	bleAdaptor := ble.NewClientAdaptor(bleName)
	var sprk sprkbot
	sprk = newDriver(bleAdaptor)

	work := func() {

		fmt.Println("starting, move around with wasd")
		sprk.SetRGB(0, 0, 10)

		//events
		sprk.On("collision", func(data interface{}) {
			fmt.Printf("collision detected = %+v \n", data)
		})

		sprk.SetRotationRate(255)

		//keyboard control
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			switch scanner.Text() {
			case "w":
				sprk.move(0)
			case "a":
				sprk.move(270)
			case "s":
				sprk.move(180)
			case "d":
				sprk.move(90)
			}

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
