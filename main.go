package main

import (
	"math"
	"os"
	"time"

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

var presets = map[string][2][2]float64{
	"delta":                  {{0.5, 4}, {0.5, 4}},
	"theta":                  {{4, 8}, {4, 8}},
	"alpha":                  {{8, 13}, {8, 13}},
	"beta":                   {{13, 30}, {13, 30}},
	"gamma":                  {{30, 50}, {30, 50}},
	"confidence":        {{60, 30}, {30, 60}},
	"relaxing":          {{70, 35}, {35, 70}},
	"higherconsciousness":    {{85, 42.5}, {42.5, 85}},
	"inspiration":            {{90, 45}, {45, 90}},
	"clarity":          {{95, 47.5}, {47.5, 95}},
	"stressrelief":           {{20, 5}, {5, 20}},
	"calm":              {{30, 7.5}, {7.5, 30}},
	"meditation":             {{45, 11.25}, {11.25, 45}},
	"creativity":        {{50, 12.5}, {12.5, 50}},
	"memoryrecall":           {{55, 13.75}, {13.75, 55}},
	"luciddreaming":          {{65, 16.25}, {16.25, 65}},
	"mindfulness":            {{70, 17.5}, {17.5, 70}},
	"productivity":           {{75, 18.75}, {18.75, 75}},
	"motivation":             {{80, 20}, {20, 80}},
	"positiveenergy":         {{85, 21.25}, {21.25, 85}},
	"anxietyrelief":          {{95, 23.75}, {23.75, 95}},
	"innerpeace":             {{100, 25}, {25, 100}},
	"positivity":             {{115, 32.5}, {32.5, 115}},
	"focus":                  {{120, 35}, {35, 120}},
	"energy":                 {{125, 37.5}, {37.5, 125}},
	"relaxation":             {{130, 40}, {40, 130}},
}

func main() {
	if len(os.Args) != 2 {
		panic("Please provide a preset name: delta, theta, alpha, beta, or gamma")
	}

	preset := os.Args[1]
	frequencies, ok := presets[preset]
	if !ok {
		panic("Invalid preset name. Available presets: delta, theta, alpha, beta, gamma")
	}

	duration := 30 * time.Minute
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	leftSine := NewSweepingSineWave(frequencies[0][0], frequencies[0][1], duration, sampleRate, 0)
	rightSine := NewSweepingSineWave(frequencies[1][0], frequencies[1][1], duration, sampleRate, 1)

	mergedStream := beep.Mix(rightSine, leftSine)

	done := make(chan bool)
	speaker.Play(beep.Seq(mergedStream, beep.Callback(func() {
		done <- true
	})))

	<-done
}
