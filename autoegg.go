package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/mabo_nanu/autoeggAudio/processAudio"
	"github.com/mzyy94/nscon"
)

func press(input *uint8, duration, wait int) {
	*input++
	time.Sleep(time.Duration(duration) * time.Millisecond)
	*input--
	time.Sleep(time.Duration(wait) * time.Millisecond)
}

func hold(input *uint8) {
	*input++
}

func holdEnd(input *uint8) {
	*input--
}

var xMap map[int]float64 = map[int]float64{
	0:   1,
	45:  0.707107,
	60:  0.5,
	90:  0,
	100: -0.173648,
	120: -0.5,
	135: -0.707107,
	180: -1,
	225: -0.707107,
	270: 0,
	315: 0.707107,
}

var yMap map[int]float64 = map[int]float64{
	0:   0,
	45:  0.707107,
	60:  0.866025,
	90:  1,
	100: 0.984808,
	120: 0.866025,
	135: 0.707107,
	180: 0,
	225: -0.707107,
	270: -1,
	315: -0.707107,
}

func inputLStickAngle(in *nscon.ControllerInput, angle int, wait int) {
	inputLStick(in, xMap[angle%360], yMap[angle%360], wait)
}

func resetLStick(in *nscon.ControllerInput, wait int) {
	inputLStick(in, 0, 0, wait)
}

func inputLStick(in *nscon.ControllerInput, x, y float64, wait int) {
	(*in).Stick.Left.X = x
	(*in).Stick.Left.Y = y
	time.Sleep(time.Duration(wait) * time.Millisecond)
}

var debug bool

func main() {
	f := flag.String("o", "None", "Log file name")
	l := flag.Int("l", 5, "Limit count")
	t := flag.Float64("t", 1.0, "Threshold")
	d := flag.Bool("d", false, "Debug")
	r := flag.Bool("r", true, "Save when shiny found")
	s := flag.Bool("s", true, "Screenshot when shiny found")
	flag.Parse()

	debug = *d

	if *f != "None" {
		logfile, err := os.OpenFile(*f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logfile)
		if debug {
			print("Output log file=")
			println(*f)
		}
	}
	log.Printf("----Start----\nLimit:%v Threshold:%v Debug:%v Report:%v\n", *l, *t, *d, *r)

	processAudio.Initialize(*t)
	defer processAudio.Terminate()

	con := nscon.NewController("/dev/hidg0")
	con.LogLevel = 0
	in := &con.Input
	con.Connect()
	defer con.Close()

	do(in, *l, *r, *s)

	log.Println("-----End-----")
}

func do(in *nscon.ControllerInput, lim int, rep bool, sshot bool) {
	runCount := 0
	noEggCount := 0
	eggInBox := 0
	eggInBag := 5
	shinyCount := 0
	shinyPos := []int{}
	hatchedCount := 0

	pre(in)
	homePosition(in)

	if talk(in) {
		eggInBox++
	}
	for shinyCount < lim {
		if eggInBag == 0 && eggInBox == 5 {
			release(in, shinyPos)
			shinyPos = nil
			eggInBox = 0
			eggInBag = 5
			runCount = 0
			if runCount > 0 {
				if talk(in) {
					eggInBox++
					noEggCount = 0
					runCount = 0
				} else {
					noEggCount++
				}
			}
		}
		stopRunCh := make(chan int)
		inLastStCh := make(chan int)
		endRunCh := make(chan int)
		go run(in, stopRunCh, inLastStCh, endRunCh)
		for eggInBag > 0 {
			time.Sleep(1000 * time.Millisecond)
			var hatched int
			var sp []int
			var lastChk bool
			select {
			case <-inLastStCh:
				<-endRunCh
				hatched, sp = hatch(in, &in.Button.B, sshot)
				lastChk = true
			default:
				hatched, sp = hatch(in, &in.Button.A, sshot)
			}
			if hatched > 0 {
				eggInBag -= hatched
				hatchedCount += hatched
				runCount = 0
				if len(sp) > 0 {
					log.Printf("Shiny found. hatched=%d\n", hatchedCount)
					shinyPos = append(shinyPos, sp...)
					shinyCount++
					if rep {
						report(in)
					}
				}
				if hatched != 5 {
					log.Printf("Error hatched=%d\n", hatched)
					press(&in.Button.Capture, 1000, 10000)
					return
				}
				homePosition(in)
			}
			if lastChk {
				break
			}
		}
		<-endRunCh
		close(stopRunCh)
		runCount++
		if eggInBox < 5 {
			if talk(in) {
				eggInBox++
				noEggCount = 0
			} else {
				noEggCount++
			}
		}
		if runCount > 10 {
			log.Println("Error runCount over 10")
			break
		}
		if noEggCount > 10 {
			log.Println("Error noEggCount over 10")
			break
		}
	}

	log.Printf("Finished! hatched=%d\n", hatchedCount)
	press(&in.Button.Home, 1000, 0)
	press(&in.Button.A, 100, 0)
	return
}

func pre(in *nscon.ControllerInput) {
	for i := 0; i < 4; i++ {
		press(&in.Button.B, 100, 900)
	}
}

// 預かり屋の隣まで移動
func homePosition(in *nscon.ControllerInput) {
	if debug {
		println("homeposition")
	}
	press(&in.Button.X, 100, 1000)
	press(&in.Button.Plus, 100, 2800)
	inputLStick(in, 0.18, 0.18, 200)
	resetLStick(in, 100)
	press(&in.Button.A, 100, 500)
	press(&in.Button.A, 100, 2500)
	inputLStickAngle(in, 120, 1100)
	inputLStickAngle(in, 45, 1100)
	resetLStick(in, 50)
}

// 預かり屋の隣から走って戻ってくる
func run(in *nscon.ControllerInput, stopCh, lastStCh, doneCh chan int) {
	if debug {
		println("run")
	}
	defer close(doneCh)
	inputLStickAngle(in, 180, 500)
	press(&in.Button.R, 100, 0)
	for i := 0; i < 16; i++ {
		select {
		case <-stopCh:
			close(lastStCh)
			return
		default:
			inputLStickAngle(in, 180+i*45, 200)
		}
	}
	for i := 0; i < 8; i++ {
		select {
		case <-stopCh:
			close(lastStCh)
			return
		default:
			inputLStickAngle(in, 135+i*45, 200)
		}
	}
	for i := 0; i < 7; i++ {
		select {
		case <-stopCh:
			close(lastStCh)
			return
		default:
			inputLStickAngle(in, 135+i*45, 200)
		}
	}
	close(lastStCh)
	inputLStickAngle(in, 100, 100)
	inputLStickAngle(in, 60, 2100)
	resetLStick(in, 0)
}

// 育て屋に話しかけて、卵が手に入ればtrue、なければfalseを返す
func talk(in *nscon.ControllerInput) bool {
	if debug {
		print("talk")
	}
	press(&in.Button.A, 100, 500)
	press(&in.Button.A, 100, 500)
	processAudio.StartClickDetect()
	press(&in.Button.B, 100, 700)
	press(&in.Button.B, 100, 500)
	if processAudio.StopClickDetect() == 0 {
		if debug {
			println(" got")
		}
		time.Sleep(1350 * time.Millisecond)
		press(&in.Button.B, 100, 1550)
		press(&in.Button.B, 100, 1500)
		press(&in.Button.B, 100, 500)
		return true
	}
	if debug {
		println()
	}
	return false
}

// 卵を孵化し、孵化した卵の数、色違いの位置を返す
func hatch(in *nscon.ControllerInput, chkBtn *uint8, sshot bool) (hatched int, shinyPos []int) {
	if debug {
		print("hatch ")
	}
	hatched = 0
	shinyPos = nil
	for {
		processAudio.StartClickDetect()
		time.Sleep(200 * time.Millisecond)
		press(chkBtn, 100, 500)
		if processAudio.StopClickDetect() == 0 {
			break // 孵化が始まっていなければ処理を抜ける
		}
		time.Sleep(13000 * time.Millisecond)
		processAudio.StartShinyDetect()
		time.Sleep(1700 * time.Millisecond)
		press(&in.Button.B, 100, 3300)
		if processAudio.StopShinyDetect() {
			if shinyPos == nil {
				shinyPos = make([]int, 0)
			}
			shinyPos = append(shinyPos, hatched)
			press(&in.Button.B, 100, 100)
			if sshot {
				press(&in.Button.Capture, 100, 100)
			}
			time.Sleep(3000 * time.Millisecond)
		}
		inputLStickAngle(in, 315, 300)
		resetLStick(in, 400)
		hatched++
	}
	if debug {
		println(hatched)
	}
	return
}

// 孵化ポケモンを逃がす。色違いは隣のBOXへ移動。
func release(in *nscon.ControllerInput, shinyPos []int) {
	if debug {
		println("release")
	}
	// ボックスを開く
	press(&in.Button.X, 100, 1000)
	press(&in.Button.A, 100, 1800)
	press(&in.Button.R, 100, 2200)
	// 色違いでないポケモンを逃がす
	press(&in.Dpad.Down, 100, 100)
	press(&in.Dpad.Left, 100, 100)
	j := 0
	for i := 0; i < 5; i++ {
		if j < len(shinyPos) && i == shinyPos[j] {
			press(&in.Dpad.Down, 100, 100)
			j++
			continue
		}
		press(&in.Button.A, 100, 600)
		press(&in.Dpad.Up, 100, 100)
		press(&in.Dpad.Up, 100, 100)
		press(&in.Button.A, 100, 1000)
		press(&in.Dpad.Up, 100, 100)
		press(&in.Button.A, 100, 1200)
		press(&in.Button.A, 100, 800)
	}
	press(&in.Button.Y, 100, 100)
	press(&in.Button.Y, 100, 100)
	// 色違いを隣のボックスへ移動
	if j > 0 {
		press(&in.Dpad.Up, 100, 100)
		press(&in.Button.A, 100, 100)
		for i := 1; i < j; i++ {
			press(&in.Dpad.Up, 100, 100)
		}
		press(&in.Button.A, 100, 100)
		press(&in.Dpad.Up, 100, 100)
		press(&in.Dpad.Up, 100, 100)
		press(&in.Dpad.Up, 100, 100)
		press(&in.Dpad.Right, 100, 100)
		press(&in.Button.A, 100, 500)
		press(&in.Dpad.Right, 100, 100)
		press(&in.Button.A, 100, 100)
		press(&in.Button.B, 100, 200)
		press(&in.Dpad.Down, 100, 100)
		press(&in.Dpad.Down, 100, 100)
	} else {
		press(&in.Dpad.Right, 100, 100)
		press(&in.Dpad.Up, 100, 100)
	}
	press(&in.Button.A, 100, 100)
	press(&in.Dpad.Up, 100, 100)
	press(&in.Button.A, 100, 400)
	press(&in.Dpad.Down, 100, 100)
	press(&in.Dpad.Left, 100, 100)
	press(&in.Button.A, 100, 800)
	press(&in.Button.B, 100, 2500)
	press(&in.Button.B, 100, 1800)
	press(&in.Button.B, 100, 1400)
}

// レポートを書く
func report(in *nscon.ControllerInput) {
	if debug {
		println("report")
	}
	press(&in.Button.X, 100, 1000)
	press(&in.Button.R, 100, 3000)
	press(&in.Button.A, 100, 5000)
	press(&in.Button.B, 100, 3000)
}
