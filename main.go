package main

import (
	"fmt"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

type SweepingSineWave struct {
	startFreq, endFreq float64
	duration           time.Duration
	sampleRate         beep.SampleRate
	samplesPlayed      int
	channel            int
}

func NewSweepingSineWave(startFreq, endFreq float64, duration time.Duration, sampleRate beep.SampleRate, channel int) *SweepingSineWave {
	return &SweepingSineWave{
		startFreq:     startFreq,
		endFreq:       endFreq,
		duration:      duration,
		sampleRate:    sampleRate,
		samplesPlayed: 0,
		channel:       channel,
	}
}

func (s *SweepingSineWave) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		freq := s.startFreq + (s.endFreq-s.startFreq)*float64(s.samplesPlayed)/(float64(s.sampleRate.N(s.duration)))
		samples[i][s.channel] = math.Sin(2 * math.Pi * freq * float64(s.samplesPlayed) / float64(s.sampleRate))
		s.samplesPlayed++
	}
	return len(samples), true
}

func (s *SweepingSineWave) Err() error {
	return nil
}

var stop chan bool

func main() {
	duration := 30 * time.Minute
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	stop = make(chan bool)

	a := app.New()
	w := a.NewWindow("albino")

	presetDropdown := widget.NewSelect(
		getPresetNames(),
		nil,
	)
	presetDropdown.PlaceHolder = "Select a preset"
	presetDropdown.SetSelected(selectRandomPreset())

	playingImageFile, err := os.Open("./img/playing2.png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open playing image file: %v\n", err)
		return
	}
	defer playingImageFile.Close()

	playingImage, err := png.Decode(playingImageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode playing image file: %v\n", err)
		return
	}

	playingCanvas := canvas.NewImageFromImage(playingImage)
	playingCanvas.SetMinSize(fyne.NewSize(580, 260))
	playingCanvas.Hide()

	playButton := widget.NewButton("  Play  ", func() {
		if presetDropdown.Selected == "" {
			return
		}

		if !playingCanvas.Visible() {
			go playPreset(presetDropdown.Selected, duration, sampleRate)
			playingCanvas.Show()
			playingCanvas.Refresh()
		}
	})

	stopButton := widget.NewButton("  Stop  ", func() {
		if playingCanvas.Visible() {

			speaker.Clear()
			stop <- true
			playingCanvas.Hide()
		}
	})

	buttons := container.NewHBox(layout.NewSpacer(), playButton, stopButton, layout.NewSpacer())

	imageFile, err := os.Open("./img/background.png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open image file: %v\n", err)
		return
	}
	defer imageFile.Close()

	image, err := png.Decode(imageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode image file: %v\n", err)
		return
	}

	background := canvas.NewImageFromImage(image)
	background.FillMode = canvas.ImageFillStretch
	background.Refresh()

	content := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), presetDropdown, layout.NewSpacer()),
		layout.NewSpacer(),
		buttons,
		layout.NewSpacer(),
		layout.NewSpacer(),
		layout.NewSpacer(),
		layout.NewSpacer(),
		container.NewGridWithColumns(1,
			container.NewHBox(layout.NewSpacer(), playingCanvas, layout.NewSpacer()),
		),
		layout.NewSpacer(),
	)

	marginTop := canvas.NewRectangle(color.Transparent)
	marginTop.SetMinSize(fyne.NewSize(0, 35))

	marginContent := container.NewVBox(marginTop, content)

	contentWithBackground := container.New(layout.NewBorderLayout(nil, nil, nil, nil),
		background, marginContent)

	w.SetContent(contentWithBackground)
	w.Resize(fyne.NewSize(600, 400))
	w.ShowAndRun()
}

func getPresetNames() []string {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	return names
}

func playPreset(preset string, duration time.Duration, sampleRate beep.SampleRate) {
	frequencies, ok := presets[preset]
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid preset name. Available presets: %v\n", getPresetNames())
		return
	}

	leftSine := NewSweepingSineWave(frequencies[0][0], frequencies[0][1], duration, sampleRate, 0)
	rightSine := NewSweepingSineWave(frequencies[1][0], frequencies[1][1], duration, sampleRate, 1)

	mergedStream := beep.Mix(rightSine, leftSine)

	done := make(chan bool)
	speaker.Play(beep.Seq(mergedStream, beep.Callback(func() {
		done <- true
	})))

	select {
	case <-done:
	case <-stop:
		speaker.Lock()
		speaker.Unlock()
	}
}

func selectRandomPreset() string {
	presetNames := getPresetNames()
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return presetNames[r.Intn(len(presetNames))]
}
