package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero"
	"gobot.io/x/gobot/platforms/sphero/ollie"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

type sprkbot struct {
	*sprkplus.SPRKPlusDriver
	heading       uint16
	position      ollie.Point2D
	extraPosition []ollie.Point2D
	navPoints     navHistory
	pointsDone    []bool
	curPoint      int
	updatePos     bool
	desiredSpeed  uint8
}

type navHistory struct {
	Points []ollie.Point2D `xml:"points"`
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

func (sprk *sprkbot) navigatorUpdate() {
	//called every 50 ms, checks for crossed points
	//fmt.Println("nav update")
	points := sprk.navPoints.Points
	i := sprk.curPoint
	newPoint := false

	if len(sprk.extraPosition) < 5 || len(points) < 3 {
		return
	}

	//select all extrapolated points
	for _, pos := range sprk.extraPosition {
		//select current and near nav points
		for j := i; j < i+4 && j < len(points); j++ {
			if distance(pos, points[j]) < 3 {
				fmt.Printf("point done: %d, %+v\n", j, points[j])
				sprk.pointsDone[j] = true
				newPoint = true
				sprk.SetRGB(10, 50, 10)
				sprk.SetRGB(0, 10, 0)
				i++
			}
		}
	}

	if newPoint {
		sprk.updatePos = true
	}

	sprk.curPoint = i

	if i >= len(points) {
		return
	}

	sprk.desiredSpeed = 40

	//angle between position->curPoint->nextPoint, to calculate speeds
	// lastPoint := sprk.position
	// curPoint := points[i]
	// nextPoint := points[i]
	// if i < len(points)-1 {
	// 	nextPoint = points[i+1]
	// }
	// v1 := ollie.Point2D{X: curPoint.X - lastPoint.X, Y: curPoint.Y - lastPoint.Y}
	// v2 := ollie.Point2D{X: nextPoint.X - curPoint.X, Y: nextPoint.Y - curPoint.Y}

	// a := angle(v1, v2)
	// d := distance(lastPoint, nextPoint)
	// //angle 180 -> 10 speed
	// //angle 0 -> 70 speed
	// speed := 20 + 50*(1-a/180)
	// if sprk.desiredSpeed > 40 && d < 10 {
	// 	speed *= d / 5
	// }
	// sprk.desiredSpeed = uint8(speed)
	// fmt.Printf("angle: %+v speed: %+v\n", a, sprk.desiredSpeed)

}

func (sprk *sprkbot) getNextAngle() float64 {
	//angle between position->curPoint->nextPoint, to calculate speeds
	points := sprk.navPoints.Points
	lastPoint := sprk.position
	i := sprk.curPoint
	curPoint := points[i]
	nextPoint := points[i]
	if i < len(points)-1 {
		nextPoint = points[i+1]
	}
	v1 := ollie.Point2D{X: curPoint.X - lastPoint.X, Y: curPoint.Y - lastPoint.Y}
	v2 := ollie.Point2D{X: nextPoint.X - curPoint.X, Y: nextPoint.Y - curPoint.Y}

	a := angle(v1, v2)
	return a
}

func (sprk *sprkbot) startNav() {
	println("started navigation")
	points := sprk.navPoints.Points

	for i, point := range points {
		sprk.curPoint = i
		if sprk.pointsDone[i] {
			continue
		}
		//stand still
		// sprk.Roll(0, sprk.heading)
		// time.Sleep(200 * time.Millisecond)
		// //turn in right direction
		// h := calcHeading(sprk.position, point)
		// sprk.Roll(0, h)
		// sprk.heading = h
		// if math.Abs(float64(h-sprk.heading)) > 4 {
		// 	time.Sleep(2000 * time.Millisecond)
		// } else {
		// 	time.Sleep(100 * time.Millisecond)
		// }
		// sprk.heading = h
		sprk.rollTo(point)
		fmt.Println("point done!")
	}

	fmt.Println("navigation done")
	sprk.Roll(0, 0)
	sprk.SetRGB(20, 20, 200)
	gobot.After(500*time.Millisecond, func() {
		sprk.SetRGB(0, 10, 0)
	})
}

func (sprk *sprkbot) rollTo(point ollie.Point2D) {
	//todo get to new points in right angle to go to next one
	rolling := false
	fmt.Println("started new rollTo")
	sprk.desiredSpeed = 40

	for index := 0; ; index++ {
		if sprk.updatePos {
			sprk.updatePos = false
			return
		}
		pos := sprk.position
		dist := distance(point, pos)

		newHeading := calcHeading(pos, point)
		if math.Abs(float64(newHeading-sprk.heading)) < 1 && rolling {
			//direction already right
			//fmt.Println("no change needed")
			time.Sleep(60 * time.Millisecond)
			sprk.heading = newHeading
			speed := sprk.desiredSpeed
			if dist < 8 && speed > 40 && sprk.getNextAngle() > 10 {
				speed = 20
				sprk.Roll(speed, sprk.heading)
				sprk.desiredSpeed = speed
			}
			if distance(sprk.position, sprk.extraPosition[0]) > 0.5 {
				continue
			}
		}

		if math.Abs(float64(newHeading-sprk.heading)) > 5 {
			//correction heading first before accelerating too fast
			sprk.Roll(0, newHeading)
			time.Sleep(300 * time.Millisecond)
		}

		sprk.heading = newHeading
		speed := sprk.desiredSpeed
		if dist < 5 && speed > 40 && sprk.getNextAngle() > 10 {
			speed = 20
		}
		sprk.Roll(speed, sprk.heading)
		sprk.desiredSpeed = speed
		fmt.Printf("targeting point: %v, own Point: %v, dist: %f\n", point, pos, dist)
		rolling = true
		time.Sleep(80 * time.Millisecond)
	}
}

// NewDriver creates a Driver for a Sphero SPRK+
func newDriver(a ble.BLEConnector) sprkbot {
	d := sprkplus.NewDriver(a)

	return sprkbot{
		SPRKPlusDriver: d,
	}
}

//distance return the distance between 2 points (euclidic)
func distance(p1 ollie.Point2D, p2 ollie.Point2D) float64 {
	return math.Sqrt(math.Pow(float64(p1.X-p2.X), 2) + math.Pow(float64(p1.Y-p2.Y), 2))
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

func main() {

	//direction which the bot is currently facing
	//heading := uint16(0)

	bleName := "SK-3C50"
	pointData := navHistory{}

	if len(os.Args) > 1 {
		bleName = os.Args[1]
	}

	ptsI := 0

	bleAdaptor := ble.NewClientAdaptor(bleName)
	var sprk sprkbot
	sprk = newDriver(bleAdaptor)

	work := func() {
		sprk.heading = 0

		sprk.SetDataStreamingConfig(sphero.DataStreamingConfig{
			N:     100,
			M:     1,
			Mask:  4294967295,
			Pcnt:  1,
			Mask2: 4294967295,
		})

		fmt.Println("starting, move around with wasd")
		sprk.SetRGB(0, 0, 10)

		//events
		sprk.On("collision", func(data interface{}) {
			fmt.Printf("collision detected = %+v \n", data)
			// sprk.SetRGB(0, 20, 0)
			// gobot.After(1*time.Second, func() {
			// 	sprk.SetRGB(0, 0, 10)
			// })

		})

		gobot.Every(150*time.Millisecond, func() {
			sprk.GetLocatorData(func(p ollie.Point2D) {
				pointData.Points = append(pointData.Points, p)
				ptsI++

				//extrapolate data - 8 steps from last to new position
				sprk.extraPosition = sprk.extraPosition[:0]
				cX := float64(p.X - sprk.position.X)
				cY := float64(p.Y - sprk.position.Y)
				for i := 0; i <= 10; i++ {
					x := float64(sprk.position.X) + cX*(float64(i)/10.)
					y := float64(sprk.position.Y) + cY*(float64(i)/10.)
					p := ollie.Point2D{X: int16(x), Y: int16(y)}
					sprk.extraPosition = append(sprk.extraPosition, p)
				}
				if distance(sprk.position, p) < 1 {
					sprk.desiredSpeed += 10
				}
				sprk.position = p
				sprk.navigatorUpdate()
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
			case "t":
				sprk.rollTo(ollie.Point2D{X: 0, Y: 0})
				sprk.Roll(0, 0)
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

				f, _ := os.Create("nav.xml")
				defer f.Close()

				enc := xml.NewEncoder(f)
				enc.Indent("  ", "    ")
				if err := enc.Encode(pointData); err != nil {
					fmt.Printf("error: %v\n", err)
					return
				}

			case "r":
				dataRaw, err := ioutil.ReadFile("nav.xml")
				if err != nil {
					return
				}

				navData := navHistory{}
				e := xml.Unmarshal([]byte(dataRaw), &navData)
				if e != nil {
					fmt.Printf("error: %v\n", err)
					return
				}

				p := ollie.Point2D{X: 0, Y: 0}
				navData.Points = append(navData.Points, p)

				sprk.navPoints = navData

				sprk.pointsDone = make([]bool, len(navData.Points))

				fmt.Println(navData)
				//start nav checker
				fmt.Printf("#of Points: %d \n", len(navData.Points))
				go func() {
					sprk.navigatorUpdate()
					time.Sleep(50 * time.Millisecond)
				}()
				sprk.startNav()

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
