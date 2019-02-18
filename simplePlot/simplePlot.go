package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero"
	"gobot.io/x/gobot/platforms/sphero/ollie"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

type sprkbot struct {
	*sprkplus.SPRKPlusDriver
	heading uint16
}

func (sprk *sprkbot) move(dir int) {
	speed := 120
	sprk.heading = uint16(dir)
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

func savePlot(movement *plot.Plot, pts plotter.XYs) {
	plotutil.AddScatters(movement, "X", pts)
	if err := movement.Save(8*vg.Inch, 8*vg.Inch, "movement.png"); err != nil {
		panic(err)
	}
	fmt.Println("saved plot!")
}

func DefaultDataStreamingConfig() sphero.DataStreamingConfig {
	return sphero.DataStreamingConfig{
		N:     100,
		M:     1,
		Mask:  4294967295,
		Pcnt:  0,
		Mask2: 4294967295,
	}
}

func main() {

	//direction which the bot is currently facing
	//heading := uint16(0)

	bleName := "SK-3C50"

	if len(os.Args) > 1 {
		bleName = os.Args[1]
	}

	//create plot
	movement, err := plot.New()
	if err != nil {
		panic(err)
	}

	movement.Title.Text = "XY movement"
	movement.X.Label.Text = "X"
	movement.Y.Label.Text = "Y"

	pts := make(plotter.XYs, 1024)
	ptsI := 0

	defer savePlot(movement, pts)

	bleAdaptor := ble.NewClientAdaptor(bleName)
	var sprk sprkbot
	sprk = newDriver(bleAdaptor)

	work := func() {

		sprk.SetDataStreamingConfig(DefaultDataStreamingConfig())

		fmt.Println("starting, move around with wasd")
		sprk.SetRGB(0, 0, 10)

		//events
		sprk.On("collision", func(data interface{}) {
			fmt.Printf("collision detected = %+v \n", data)
			sprk.SetRGB(0, 20, 0)
			gobot.After(1*time.Second, func() {
				sprk.SetRGB(0, 0, 10)
			})

		})

		sprk.On("sensordata", func(data interface{}) {
			// cont := data.(ollie.DataStreamingPacket)
			// fmt.Printf("got sensorData event: %+v\n", cont)
			// ptsX[ptsI].X = float64(ptsI)
			// ptsX[ptsI].Y = float64(cont.VeloX)
			// ptsY[ptsI].X = float64(ptsI)
			// ptsY[ptsI].Y = float64(cont.VeloY)
			// ptsZ[ptsI].X = float64(ptsI)
			// ptsZ[ptsI].Y = float64(cont.AccelOne)
			// ptsI++

		})

		gobot.Every(200*time.Millisecond, func() {
			sprk.GetLocatorData(func(p ollie.Point2D) {
				fmt.Printf("got locator data: x=%d y=%d\n", p.X, p.Y)
				//pts := make(plotter.XYs, 128)
				pts[ptsI].X = float64(p.X)
				pts[ptsI].Y = float64(p.Y)
				ptsI++
			})
		})

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
			case "e":
				sprk.Roll(255, sprk.heading)
			case "c":
				stepSize := 25

				for i := 0; i <= 360*2; i += stepSize {
					sprk.Roll(90, uint16(i%360))
					sprk.heading = uint16(i % 360)
					time.Sleep(100 * 2 * time.Millisecond)
					fmt.Printf("%d\n", i)
				}
				sprk.Stop()
			case "q":
				savePlot(movement, pts)
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
