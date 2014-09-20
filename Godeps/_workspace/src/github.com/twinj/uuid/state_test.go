package uuid

/****************
 * Date: 14/02/14
 * Time: 9:08 PM
 ***************/

import (
	"bytes"
	"testing"
	"time"
)

var state_bytes = []byte{
	0xAA, 0xCF, 0xEE, 0x12,
	0xD4, 0x00,
	0x27, 0x23,
	0x00,
	0xD3,
	0x23, 0x12, 0x4A, 0x11, 0x89, 0xFF,
}

func TestUUID_StateSeed(t *testing.T) {
	if state.past < Timestamp((1391463463*10000000)+(100*10)+gregorianToUNIXOffset) {
		t.Errorf("Expected a value greater than 02/03/2014 @ 9:37pm in UTC but got %d", state.past)
	}
	if state.node == nil {
		t.Errorf("Expected a non nil node")
	}
	if state.sequence <= 0 {
		t.Errorf("Expected a value greater than but got %d", state.sequence)
	}
}

func TestUUIDState_read(t *testing.T) {
	s := new(State)
	s.past = Timestamp((1391463463*10000000)+(100*10)+gregorianToUNIXOffset)
	s.node = state_bytes

	now := Timestamp((1391463463 * 10000000) + (100 * 10))
	s.read(now+(100*10), make([]byte, length))

	if s.sequence != 1 {
		t.Error("The sequence should increment when the time is" +
					"older than the state past time and the node" +
					"id are not the same.", s.sequence)
	}
	s.read(now, state_bytes)

	if s.sequence == 1 {
		t.Error("The sequence should be randomly generated when" +
					" the nodes are equal.", s.sequence)
	}

	s = new(State)
	s.past = Timestamp((1391463463*10000000)+(100*10)+gregorianToUNIXOffset)
	s.node = state_bytes
	s.randomSequence = true
	s.read(now, make([]byte, length))

	if s.sequence == 0 {
		t.Error("The sequence should be randomly generated when" +
					" the randomSequence flag is set.", s.sequence)
	}

	if s.past != now {
		t.Error("The past time should equal the time passed in" +
				" the method.")
	}

	if !bytes.Equal(s.node, make([]byte, length)) {
		t.Error("The node id should equal the node passed in" +
				" the method.")
	}
}

func TestUUIDState_init(t *testing.T) {


}

// Tests that the schedule is run approx every ten seconds
// takes 90 seconds to complete on my machine at 90000000 UUIDs
func TestUUIDState_saveSchedule(t *testing.T) {
	if V1Save {
		count := 0
		now := time.Now()
		NewV1() // prime scheduler
		state.next = timestamp()
		for i := 0; i < 10000000; i++ {
			stamp := timestamp()
			if stamp >= state.next {
				count++
			}
			NewV1()
		}
		d := time.Since(now)
		tenSec := int(d.Seconds()) / int(SaveSchedule) + 1
		if count != tenSec {
			t.Errorf("Should be as many saves as ten second increments but got: %d instead of %d", count, tenSec)
		}   // TODO fix extra save
	}
}

func TestUUID_encode(t *testing.T) {

}

func TestUUID_decode(t *testing.T) {

}


