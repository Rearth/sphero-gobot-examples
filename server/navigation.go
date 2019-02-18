package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero"
	"gobot.io/x/gobot/platforms/sphero/ollie"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

//updated version of the navigation, used by the server to control the sphero and get current information

//Sprkbot contains relevant data to control the sphero
type Sprkbot struct {
	*sprkplus.SPRKPlusDriver
	heading           uint16
	position          ollie.Point2D
	extraPosition     []ollie.Point2D
	nextPoint         ollie.Point2D
	hasTarget         bool
	speed             uint8
	doneCallback      func()
	collisionCallback func()
	pointQueue        []ollie.Point2D
	boosting          bool
}

func calcHeading(p1, p2 ollie.Point2D) uint16 {
	res := ollie.Point2D{X: p2.X - p1.X, Y: p2.Y - p1.Y}
	angle := math.Atan2(float64(res.X), float64(res.Y)) * 180. / math.Pi

	newHeading := uint16(0)
	if angle >= 0 {
		newHeading = uint16(angle)
	} else {
		newHeading = uint16(360 + angle)
	}

	return newHeading
}

func (sprk *Sprkbot) addToQueue(x, y int) {
	fmt.Println("got new queue point!")
	p := ollie.Point2D{X: int16(x), Y: int16(y)}
	sprk.pointQueue = append(sprk.pointQueue, p)
}

//GoToPoint drives the sphero to the targeted point and inits the nav data
func (sprk *Sprkbot) GoToPoint(x, y int) {
	fmt.Println("got new target point!")
	p := ollie.Point2D{X: int16(x), Y: int16(y)}
	sprk.nextPoint = p
	sprk.hasTarget = true
	sprk.speed = 40
}

func (sprk *Sprkbot) home() {
	sprk.GoToPoint(0, 0)
	sprk.doneCallback = func() {
		time.Sleep(50 * time.Millisecond)
		sprk.Roll(0, 0)
		sprk.heading = 0
		sprk.SetRGB(10, 100, 10)
		time.Sleep(200 * time.Millisecond)
		sprk.SetRGB(0, 10, 0)
		sprk.doneCallback = func() {

		}
	}
}

func (sprk *Sprkbot) update() {
	//called every 75ms

	if len(sprk.pointQueue) > 0 && !sprk.hasTarget {
		sprk.GoToPoint(int(sprk.pointQueue[0].X), int(sprk.pointQueue[0].Y))
	}

	if !sprk.hasTarget {
		return
	}

	//checks if a target has been reached
	if sprk.checkTarget() {
		//slow down
		sprk.Roll(0, sprk.heading)
		sprk.hasTarget = false
		if len(sprk.pointQueue) <= 1 {
			sprk.pointQueue = sprk.pointQueue[:0] //empty slice
		} else {
			sprk.pointQueue = sprk.pointQueue[1:] //remove first point

		}
		go sprk.doneCallback()
		return
	}

	pos := sprk.position
	target := sprk.nextPoint
	if distance(pos, target) > 90 && sprk.speed < 149 {
		sprk.speed = 150
	} else if distance(pos, target) > 70 && sprk.speed > 59 {
		sprk.speed = 60
	} else if distance(pos, target) < 35 && sprk.speed > 39 {
		sprk.speed = 30
	} else if distance(pos, target) < 30 && sprk.speed > 29 {
		sprk.speed = 25
	} else if distance(pos, target) < 20 && sprk.speed > 24 {
		sprk.speed = 15
	} else if distance(pos, target) < 15 && sprk.speed > 24 {
		sprk.speed = 5
	}

	if sprk.boosting {
		sprk.speed += 100
	}

	newHeading := calcHeading(pos, target)
	sprk.heading = newHeading
	sprk.Roll(sprk.speed, newHeading)
	fmt.Printf("rolling to %+v, heading: %d, speed: %d\n", target, newHeading, sprk.speed)

}

func getSpeedFromCurve(x float64) uint8 {
	v := -38.06184 + 3.936315*x - 0.07822995*math.Pow(x, 2) + 0.0006000874*math.Pow(x, 3)
	return uint8(v)

}

func (sprk *Sprkbot) checkTarget() bool {
	for _, p := range sprk.extraPosition {
		if distance(p, sprk.nextPoint) < 1.7 {
			fmt.Println("point done")
			return true
		}
	}

	return false
}

func (sprk *Sprkbot) getPosition() ollie.Point2D {
	return sprk.position
}

// NewDriver creates a Driver for a Sphero SPRK+
func newDriver(a ble.BLEConnector) Sprkbot {
	d := sprkplus.NewDriver(a)

	return Sprkbot{
		SPRKPlusDriver: d,
	}
}

//distance return the distance between 2 points (euclidic)
func distance(p1 ollie.Point2D, p2 ollie.Point2D) float64 {
	return math.Sqrt(math.Pow(float64(p1.X-p2.X), 2) + math.Pow(float64(p1.Y-p2.Y), 2))
}

func (sprk *Sprkbot) clearQueue() {
	sprk.pointQueue = sprk.pointQueue[:0] //empty slice
}

//angle returns the signed angle between 2 vectors -> rotate vector 1 by x degree counter-clockwise to get vector 2
func angle(p1, p2 ollie.Point2D) float64 {
	fmt.Printf("calculating angle: %+v, %+v", p1, p2)
	r := (math.Atan2(float64(p2.Y), float64(p2.X)) - math.Atan2(float64(p1.Y), float64(p1.X))) * 180. / math.Pi
	if r > 180 {
		return r - 360
	} else if r < -180 {
		return r + 360
	}
	return r
}

//Create starts the whole process
func Create() (*gobot.Robot, *Sprkbot) {

	bleName := "SK-3C50"

	if len(os.Args) > 1 {
		bleName = os.Args[1]
	}

	bleAdaptor := ble.NewClientAdaptor(bleName)
	var sprk Sprkbot
	sprk = newDriver(bleAdaptor)

	work := func() {
		sprk.heading = 0

		fmt.Println("starting robot work")
		sprk.SetRGB(0, 0, 10)

		//send only 1 packet
		sprk.SetDataStreamingConfig(sphero.DataStreamingConfig{
			N:     100,
			M:     1,
			Mask:  4294967295,
			Pcnt:  1,
			Mask2: 4294967295,
		})

		//events
		sprk.On("collision", func(data interface{}) {
			fmt.Printf("collision detected = %+v \n", data)
			sprk.SetRGB(0, 100, 100)
			sprk.collisionCallback()
			gobot.After(500*time.Millisecond, func() {
				sprk.SetRGB(0, 10, 0)
			})

		})

		// gobot.Every(75*time.Millisecond, func() {
		// 	sprk.update()
		// })

		gobot.Every(130*time.Millisecond, func() {
			sprk.GetLocatorData(func(p ollie.Point2D) {

				//interpolate points
				sprk.extraPosition = sprk.extraPosition[:0]
				cX := float64(p.X - sprk.position.X)
				cY := float64(p.Y - sprk.position.Y)
				for i := 0; i <= 10; i++ {
					x := float64(sprk.position.X) + cX*(float64(i)/10.)
					y := float64(sprk.position.Y) + cY*(float64(i)/10.)
					p := ollie.Point2D{X: int16(x), Y: int16(y)}
					sprk.extraPosition = append(sprk.extraPosition, p)
				}

				if distance(sprk.position, p) < 1 && sprk.hasTarget {
					sprk.speed += 9
				}

				sprk.position = p
				sprk.update()

			})
		})
		fmt.Println("robot work done")
	}

	sprk.doneCallback = func() {

	}

	robot := gobot.NewRobot("sprkie",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{sprk},
		work,
	)

	return robot, &sprk

}
