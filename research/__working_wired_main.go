package main

import (
	"log"
	"math/bits"
	"time"

	"remotepi/pkg/rpio"

	"github.com/morphar/powernap"
)

// RC-5 frame format:
// Header (3 bits)
// Device address (5 bits)
// Device command (6 bits)
// End silence (50 bit)

// RC-5x frame format:
// 1101 0000 (2 bit space) 0011 0000 0001
// Header (3 bits)
// Device address (5 bits)
// Space silence (2 bit)
// Device command (6 bits)
// Device command (6 bits)
// End silence (42 bit)

// Since the repetition of the 36 kHz carrier is 27.778 μs and the duty factor is 25%,
// the carrier pulse duration is 6.944 μs.
// Each bit of the RC-5 code word contains 32 carrier pulses, and an equal duration of silence,
// so the bit time is 64×27.778 μs = 1.778 ms,
// and the 14 symbols (bits) of a complete RC-5 code word take 24.889 ms to transmit.
// The code word is repeated every 113.778 ms (4096 / 36 kHz) as long as a key remains pressed.

const frequency = 36000 // Carrier frequenzy is 36kHz
const dutyCycle = 0.25  // Duty cycle is 25%
const dutyPulseFrequncy = frequency / dutyCycle
const halfBitPulseRepetitions = 32 // Each half-bit repeats the pulse 32 times
const idleBits = 50

const fullBitFrequency = 562.429
const halfBitFrequency = fullBitFrequency / 2 // 281,2145

// var fullBitDuration = time.Duration(float64(time.Second) / fullBitFrequency) // 1778.002µs
// var halfBitDuration = fullBitDuration / 2 // 889.001µs

// Single bit (dual half-bit) durations
const fullBitDuration = 1778 * time.Microsecond
const halfBitDuration = 889 * time.Microsecond
const dutyCycleDuration = time.Second / frequency         // 27.777µs
const dutyPulseDuration = time.Second / dutyPulseFrequncy // 6.944µs
const idleDuration = idleBits * fullBitDuration

// const carrierFrequency = time.Second / frequency // 27.7778 microseconds

/*
const frequency = 36000 // Carrier frequenzy is 36kHz
const dutyCycle = 0.25  // Duty cycle is 25%

const fullBitFrequency = 562.429
const halfBitFrequency = fullBitFrequency / 2 // 281,2145

var fullBitDuration = time.Duration(int64(float64(time.Second) / 562.429))

// const fullBitDuration = time.Duration(float64(time.Second) / fullBitFrequency)
const halfBitDuration = fullBitDuration / 2
*/

// Idea
// Use a select to change the duty cycle
// On each select, use the next bit in the command

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// How to build the signal sender:
// - build the signal to send
// - start a timer that ticks for every duty cycle * 4
//    - this is used for changing the high/low signal
//    - current signal is stored in a var defined in the next ticker
//    - maybe have a counter to know when to to send the 1/4 signal and when to be silent
// - start a timer that ticks for every half bit
//    - this should control where we are in the signal and change it when needed

func addHigh(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	plan.Schedule(start, pin.Low)
	start += halfBitDuration

	for i := 0; i < halfBitPulseRepetitions; i++ {
		plan.Schedule(start, pin.High)
		// plan.Schedule(start+dutyPulseDuration, pin.Low)
		start += dutyCycleDuration
	}
}

func addLow(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	for i := 0; i < halfBitPulseRepetitions; i++ {
		plan.Schedule(start, pin.High)
		// plan.Schedule(start+dutyPulseDuration, pin.Low)
		start += dutyCycleDuration
	}

	plan.Schedule(start, pin.Low)
}

func sendSignal(pin rpio.Pin, bin uint) {
	pin.Output() // Output mode

	plan := powernap.NewPlan()

	len := bits.Len(bin)
	n := time.Duration(0)
	for i := len - 1; i >= 0; i, n = i-1, n+1 {
		if bin&(1<<i) == 0 {
			addLow(plan, n*fullBitDuration, pin)
		} else {
			addHigh(plan, n*fullBitDuration, pin)
		}
	}

	plan.Schedule(n*fullBitDuration, pin.Low)
	plan.Schedule((n*fullBitDuration)+idleDuration, func() { /* idle */ })

	plan.StartBlocking()
	// plan.StartBlocking()
	// spew.Dump(res.ExecTimes)

	///////////////////////////////////////////////////////////////

	// for i, et := range res.ExecTimes {
	// 	expected := res.Start.Add(planExpected[i])
	// 	diff := math.Abs(float64(et.Sub(expected)))
	// 	diffRatio := diff / float64(planExpected[i])

	// 	fmt.Println(expected)
	// 	fmt.Println(et)
	// 	fmt.Println(diff)
	// 	fmt.Println(planExpected[i])

	// 	if i > 0 {
	// 		fmt.Println(et.Sub(res.ExecTimes[i-1]))
	// 	}

	// 	if diffRatio > 0.01 {
	// 		fmt.Printf("expected an error rate of less than 1%%, got: %0.2f%%\n", diffRatio*100)
	// 	}
	// 	fmt.Println("")
	// }
}

func main() {
	err := rpio.Open()
	panicOnErr(err)
	defer rpio.Close()

	pin := rpio.Pin(17)
	defer pin.Low()

	standBy := 0b11010000001100
	sendSignal(pin, uint(standBy))
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
