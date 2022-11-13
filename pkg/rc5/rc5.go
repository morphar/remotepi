package rc5

import (
	"math/bits"
	"time"

	"github.com/stianeikeland/go-rpio/v4"

	"github.com/morphar/powernap"
)

// RC-5 frame format:
// 110 10000 001100
// Header (3 bits)
// Device address (5 bits)
// Device command (6 bits)
// End silence (50 bit)

// RC-5x frame format:
// 110 10000 (2 bit space) 001100 000001
// Header (3 bits)
// Device address (5 bits)
// Space silence (2 bit)
// Device command (6 bits)
// Device command (6 bits)
// End silence (42 bit)
// NOTE: if the bit number 2 is 0, then add 64 to command value...???

// Since the repetition of the 36 kHz carrier is 27.778 μs and the duty factor is 25%,
// the carrier pulse duration is 6.944 μs.
// Each bit of the RC-5 code word contains 32 carrier pulses, and an equal duration of silence,
// so the bit time is 64×27.778 μs = 1.778 ms,
// and the 14 symbols (bits) of a complete RC-5 code word take 24.889 ms to transmit.
// The code word is repeated every 113.778 ms (4096 / 36 kHz) as long as a key remains pressed.

// const carrierFrequency = time.Second / frequency // 27.7778 microseconds

const frequency = 36000 // Carrier frequenzy is 36kHz
const dutyCycle = 0.25  // Duty cycle is 25%
const dutyPulseFrequncy = frequency / dutyCycle
const halfBitPulseRepetitions = 32 // Each half-bit repeats the pulse 32 times
const rc5IdleBits = 50
const rc5XIdleBits = 42

// const fullBitFrequency = 562.429
// const halfBitFrequency = fullBitFrequency / 2 // 281,2145

// Single bit (dual half-bit) durations
// var fullBitDuration = time.Duration(float64(time.Second) / fullBitFrequency) // 1778.002µs
const fullBitDuration = 1778 * time.Microsecond

// var halfBitDuration = fullBitDuration / 2 // 889.001µs
const halfBitDuration = 889 * time.Microsecond
const dutyCycleDuration = time.Second / frequency         // 27.777µs
const dutyPulseDuration = time.Second / dutyPulseFrequncy // 6.944µs
const rc5IdleDuration = rc5IdleBits * fullBitDuration
const rc5XIdleDuration = rc5XIdleBits * fullBitDuration

// const rc5xSpaceDuration = fullBitDuration * 2

// How to build the signal sender:
// - build the signal to send
// - start a timer that ticks for every duty cycle * 4
//    - this is used for changing the high/low signal
//    - current signal is stored in a var defined in the next ticker
//    - maybe have a counter to know when to to send the 1/4 signal and when to be silent
// - start a timer that ticks for every half bit
//    - this should control where we are in the signal and change it when needed

func addSilence(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	plan.Schedule(start, pin.Low)
}

func addWiredHigh(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	plan.Schedule(start, pin.Low)
	start += halfBitDuration
	plan.Schedule(start, pin.High)
}

func addWiredLow(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	plan.Schedule(start, pin.High)
	start += halfBitDuration
	plan.Schedule(start, pin.Low)
}

func addIRHigh(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	plan.Schedule(start, pin.Low)
	start += halfBitDuration

	for i := 0; i < halfBitPulseRepetitions; i++ {
		plan.Schedule(start, pin.High)
		plan.Schedule(start+dutyPulseDuration, pin.Low)
		start += dutyCycleDuration
	}
}

func addIRLow(plan *powernap.Plan, start time.Duration, pin rpio.Pin) {
	for i := 0; i < halfBitPulseRepetitions; i++ {
		plan.Schedule(start, pin.High)
		plan.Schedule(start+dutyPulseDuration, pin.Low)
		start += dutyCycleDuration
	}

	plan.Schedule(start, pin.Low)
}

func Send(pin rpio.Pin, bin int, wired bool) {
	pin.Output() // Output mode

	var addLow func(*powernap.Plan, time.Duration, rpio.Pin)
	var addHigh func(*powernap.Plan, time.Duration, rpio.Pin)
	if wired {
		addLow = addWiredLow
		addHigh = addWiredHigh
	} else {
		addLow = addIRLow
		addHigh = addIRHigh
	}

	plan := powernap.NewPlan()

	bitLen := bits.Len(uint(bin))
	n := time.Duration(0)

	// RC-5X NOTE: if the bit number 2 is 0, then add 64 to command value...???

	idleDuration := rc5IdleDuration

	for i := bitLen - 1; i >= 0; i, n = i-1, n+1 {
		if bitLen == 22 && (i == 12 || i == 13) {
			idleDuration = rc5XIdleDuration
			addSilence(plan, n*fullBitDuration, pin)
			continue
		}

		if bin&(1<<i) == 0 {
			addLow(plan, n*fullBitDuration, pin)
		} else {
			addHigh(plan, n*fullBitDuration, pin)
		}
	}

	plan.Schedule(n*fullBitDuration, pin.Low)
	plan.Schedule((n*fullBitDuration)+idleDuration, func() { /* idle */ })

	plan.StartTightBlocking()
	plan.StartTightBlocking()
	plan.StartTightBlocking()
	plan.StartTightBlocking()
}

// RC-5 frame format:
// 110 10000 001100
// Header (3 bits)
// Device address (5 bits)
// Device command (6 bits)
// End silence (50 bit)
func Command(system, command, toggle int) int {
	bin := 0b11

	bin = bin << 1
	if toggle == 1 {
		bin |= 0b001
	}

	bin = bin << 5
	bin |= system

	bin = bin << 6
	bin |= command

	return bin
}

// RC-5x frame format:
// 110 10000 (2 bit space) 001100 000001
// Header (3 bits)
// Device address (5 bits)
// Space silence (2 bit)
// Device command (6 bits)
// Device command (6 bits)
// End silence (42 bit)
// NOTE: if the bit number 2 is 0, then add 64 to command value...???
func CommandX(system, command, ext, toggle int) int {
	bin := 0b11

	if command >= 64 {
		bin = 0b10
		command -= 64
	}

	bin = bin << 1
	if toggle == 1 {
		bin |= 0b001
	}

	bin = bin << 5
	bin |= system

	bin = bin << 2 // Space

	bin = bin << 6
	bin |= command

	bin = bin << 6
	bin |= ext

	return bin
}
