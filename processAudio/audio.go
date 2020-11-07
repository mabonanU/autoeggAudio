package processAudio

import (
	"github.com/gordonklaus/portaudio"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

const (
	rate   float64 = 44100
	sample int     = 4096
)

var stream *portaudio.Stream
var windowFunc [sample]float64 //窓関数
var iClick int                 //会話等でのボタンクリック音の周波数(2718Hz)のindex
var tClick float64             //クリック音と判定する2718Hz閾値
var iBicycle1 int              //自転車チャージ完了音の周波数(2391Hz)のindex
var iBicycle2 int              //自転車チャージ完了音の周波数(2423Hz)のindex
var iBicycle3 int              //自転車チャージ完了音の周波数(2466Hz)のindex
var tBicycle float64           //クリック音ではなく自転車チャージ完了音と判定する閾値(iBicycle1~3の合計)
var prevClick bool             //直前のフレームでクリックと判定されていたか
var cCount int                 //クリック判定された回数
var doClickDetect bool         //クリック判定を行うかどうかのフラグ
var iShiny1 int                //色違いが産まれたときの音(8080Hz)のindex
var iShiny2 int                //色違いが産まれたときの音(12107Hz)のindex
var tShiny1 float64            //色違いが産まれたと判定する8080Hzの閾値
var tShiny2 float64            //色違いが産まれたと判定する12107Hzの閾値
var sCount int                 //閾値を連続して超えた回数をカウント
var tShCount int               //閾値を連続で超えた回数の閾値
var sDetected bool             //色違い判定されたか
var doShinyDetect bool         //色違い判定を行うかどうかのフラグ

// Initialize audio
func Initialize(threshold float64) {
	chk(portaudio.Initialize())

	var err error
	stream, err = portaudio.OpenDefaultStream(1, 0, rate, sample, processAudio)
	chk(err)

	iClick = 2718*sample/int(rate) + 1
	iBicycle1 = 2390*sample/int(rate) + 1
	iBicycle2 = 2422*sample/int(rate) + 1
	iBicycle3 = 2465*sample/int(rate) + 1
	iShiny1 = 8080*sample/int(rate) + 1
	iShiny2 = 12107*sample/int(rate) + 1

	copy(windowFunc[:], window.Hamming(sample))

	tClick = 1.0 * threshold
	tBicycle = 0.5 * threshold
	tShiny1 = 0.6 * threshold
	tShiny2 = 0.4 * threshold
	tShCount = 5

	chk(stream.Start())
}

// Terminate stream
func Terminate() {
	chk(stream.Stop())
	stream.Close()
	chk(portaudio.Terminate())
}

func processAudio(in []float32) {
	windowed := make([]float64, sample)
	for i, v := range in {
		windowed[i] = float64(v) * windowFunc[i]
	}
	fftdata := fft.FFTReal(windowed)

	// クリック判定
	if doClickDetect {
		sqd := square(fftdata[iClick])
		if sqd > tClick {
			if !prevClick {
				bccl := (square(fftdata[iBicycle1]) + square(fftdata[iBicycle2]) + square(fftdata[iBicycle3]))
				if bccl < tBicycle {
					cCount++
					prevClick = true
				}
			}
		} else {
			prevClick = false
		}
	}
	// 色違い判定
	if doShinyDetect {
		shin1 := square(fftdata[iShiny1])
		shin2 := square(fftdata[iShiny2])
		if shin1 > tShiny1 && shin2 > tShiny2 {
			sCount++
			if sCount > tShCount {
				sDetected = true
				doShinyDetect = false
			}
		} else {
			sCount = 0
		}
	}
}

// クリック音の検知を開始する
func StartClickDetect() bool {
	if doClickDetect == true {
		return false // すでに検知開始している場合はfalseを返す
	}
	cCount = 0
	prevClick = false
	doClickDetect = true
	return true
}

// 前回StartClickDetect()を呼んでからクリック音を検知した回数を返す。
func StopClickDetect() int {
	doClickDetect = false
	return cCount
}

func StartShinyDetect() {
	sCount = 0
	sDetected = false
	doShinyDetect = true
}

func StopShinyDetect() bool {
	doShinyDetect = false
	return sDetected
}

func square(c complex128) float64 {
	return real(c)*real(c) + imag(c)*imag(c)
}

func sum(arr []float32) float32 {
	var temp float32 = 0
	for _, v := range arr {
		temp += v
	}
	return temp
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
