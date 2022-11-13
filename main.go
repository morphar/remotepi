package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/morphar/remotepi/pkg/rc5"
	"github.com/stianeikeland/go-rpio/v4"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// $ cat /proc/asound/card0/pcm*/sub*/status
// closed

// $ cat /proc/asound/card0/pcm*/sub*/status
// state: RUNNING

// The regexp for matching ON state of the audio stream
var reRunning = regexp.MustCompile("state: RUNNING")

func main() {
	// Open pin with the remote control rca connected
	err := rpio.Open()
	exitOnErr(err)
	defer rpio.Close()

	// Currently only supports pin 17
	pin := rpio.Pin(17)
	defer pin.Low()

	// Find all audio card status files (hopefully only 1)
	matches, err := filepath.Glob("/proc/asound/card0/pcm*/sub*/status")
	exitOnErr(err)

	if len(matches) != 1 {
		log.Fatal("For now, this only works with 1 audio card")
	}

	statusFile := matches[0]

	// Create a couple of amplifier rc commands
	// onOff := rc5Command(16, 12, 0)
	// volumeUp := rc5Command(16, 16, 0)
	// volumeDown := rc5Command(16, 17, 0)
	turnOn := rc5.CommandX(16, 12, 1, 0)
	turnOff := rc5.CommandX(16, 12, 2, 0)
	// directVolume := rc5xCommand(16, 111, 10, 0)

	// onOff := 0b11010000001100
	// turnOff := 0b1101000000001100000010
	// sendSignal(pin, uint(onOff), true)

	// Delays before turning on or off the amplifier
	offDeleay := 30 * time.Second
	onDeleay := time.Second

	// Setup the check vars
	var lastOn time.Time
	var stateOn bool

	for {
		src, err := os.ReadFile(statusFile)
		exitOnErr(err)

		if reRunning.Match(src) {
			lastOn = time.Now()
			if !stateOn {
				stateOn = true
				rc5.Send(pin, turnOn, true)
			}
			time.Sleep(offDeleay)
			continue
		}

		if stateOn && time.Since(lastOn) > offDeleay {
			stateOn = false
			rc5.Send(pin, turnOff, true)
		}

		time.Sleep(onDeleay)
	}
}

func exitOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
