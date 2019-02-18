package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero/ollie"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

func main() {

	//missing events: Freefall/Landing

	bleName := "SK-3C50"

	if len(os.Args) > 1 {
		bleName = os.Args[1]
	}

	bleAdaptor := ble.NewClientAdaptor(bleName)
	sprk := sprkplus.NewDriver(bleAdaptor)

	work := func() {
		fmt.Println("starting work")
		// sprk.On("collision", func(data interface{}) {
		// 	fmt.Print("got collision event")
		// 	fmt.Println(data)
		// })

		// sprk.On("sensordata", func(data interface{}) {
		// 	cont := data.(sphero.DataStreamingPacket)
		// 	fmt.Print("got sensorData event")
		// 	fmt.Println(cont)
		// 	fmt.Printf("falling:=%d\n", cont.AccelOne)
		// })

		fmt.Println(sprk.Events())

		sprk.SetRGB(100, 0, 0)

		// sprk.SetDataStreamingConfig(sphero.DefaultDataStreamingConfig())

		for i := 0; ; i++ {
			fmt.Println("requesting power state!")
			sprk.GetPowerState(func(p ollie.PowerStatePacket) {
				fmt.Printf("got power state: %+v\n", p)
			})
			time.Sleep(500 * time.Millisecond)
		}

	}

	robot := gobot.NewRobot("sprkie",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{sprk},
		work,
	)

	sprk.AddEvent("freefall")
	sprk.AddEvent("landing")

	robot.Start()
	// bleAdaptor.Subscribe("22bb746f2ba675542d6f726568705327", func(data []byte, e error) {
	// 	fmt.Println(data)
	// })

}
