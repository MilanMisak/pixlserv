package uuid
/****************
 * Date: 14/02/14
 * Time: 7:43 PM
 ***************/

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"os"
	seed "math/rand"
)

var (
	Report bool
)

func init() {
	gob.Register(stateEntity{})
}

// **************************************************** State

// Holds package information about the current
// state of the UUID generator
type State struct {

	// A flag which informs whether to
	// randomly create a node id
	randomNode         bool

	// A flag which informs whether to
	// randomly create the sequence
	randomSequence     bool

	// the last time UUID was saved
	past               Timestamp

	// the next time the state will be saved
	next Timestamp

	// the last node which saved a UUID
	node               []byte

	// An iterated value to help ensure different
	// values across the same domain
	sequence           uint16

	// file to save state
	check *os.File
}

// Changes the state with current data
// Compares the current found node to the last node stored,
// If they are the same or randomSequence is already set due
// to an earlier read issue then the sequence is randomly generated
// else if there is an issue with the time the sequence is incremeted
func (o *State) read(pNow Timestamp, pNode net.HardwareAddr) {
	if bytes.Equal(pNode, o.node) || o.randomSequence {
		o.sequence = uint16(seed.Int()) & 0x3FFF
	} else if pNow < o.past {
		o.sequence ++
	}
	o.past = pNow
	o.node = pNode
}

// Saves the current state of the generator
// If the scheduled file save is reached then the file is synced
func (o *State) save() {
	if o.past >= o.next {
		var err error
		o.check, err = os.OpenFile("uuid" + "/" + "state.unique", os.O_RDWR, os.ModeExclusive)
		defer o.check.Close()
		if err != nil {
			log.Println("UUID.State.save:", err)
			return
		}
		// do the save
		o.encode()
		// schedule next save for 10 seconds from now
		o.next = o.past + 10*ticksPerSecond
		if Report {
			log.Printf("UUID STATE: SAVED %d", o.past)
		}
	}
}

// Initialises the UUID state when the package is first loaded
// it first attempts to decode the file state into State
// if this file does not exist it will create the file and do a flush
// of the random state which gets loaded at packarge runtime
// second it will attempt to resolve the current hardware address nodeId
// thirdly it will check the state of the clock
func (o *State) init() {
	var err error
	o.check, err = os.OpenFile("uuid" + "/" + "state.unique", os.O_RDWR, os.ModeExclusive)
	defer o.check.Close()
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("'%s' %s created\n", "uuid/", "UUIDState")
			os.Mkdir("uuid" + "/", os.ModeDir | 0755)
			o.check, err = os.Create("uuid" + "/" + "state.unique")
			if err != nil {
				log.Println("UUID.State.init: file error:", err)
				goto nodeId
			}
			o.encode()
		} else {
			log.Println("UUID.State.init: file error:", err)
			goto nodeId
		}
	}
	err = o.decode()
	if err != nil {
		goto nodeId
	}
	o.randomSequence = false
nodeId:
{
	intfcs, err := net.Interfaces()
	if err != nil {
		log.Println("UUID.State.init: address error:", err)
		goto pastInit
	}
	a := getHardwareAddress(intfcs)
	if a == nil {
		log.Println("UUID.State.init: address error:", err)
		goto pastInit
	}
	if bytes.Equal(a, o.node) {
		o.sequence ++
	}
	o.node = a
	o.randomNode = false
}
pastInit:
	if timestamp() <= o.past {
		o.sequence ++
	}
	o.next = state.past
}

// ***********************************************  StateEntity

// StateEntity acts as a marshaller struct for the state
type stateEntity struct {
	Past       Timestamp
	Node       []byte
	Sequence   uint16
}

// Encodes State generator data into a saved file
func (o *State) encode() {
	// ensure at beginning of file - cause overwrite of old state
	o.check.Seek(0, 0)
	enc := gob.NewEncoder(o.check)
	// Wrap private State data into the StateEntity
	entity := stateEntity{state.past, state.node, state.sequence}
	err := enc.Encode(&entity)
	if err != nil {
		log.Panic("UUID.encode error:", err)
	}
}

// Decodes StateEntity data into the main State
func (o *State) decode() error {
	o.check.Seek(0, 0)
	dec := gob.NewDecoder(o.check)
	entity := stateEntity{}
	err := dec.Decode(&entity)
	if err != nil {
		log.Println("UUID.decode error:", err)
		return err
	}
	o.past = entity.Past
	o.node = entity.Node
	o.sequence = entity.Sequence
	return nil
}
