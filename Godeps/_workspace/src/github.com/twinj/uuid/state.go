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
	// Print save log
	Report = false

	// Save every x seconds
	SaveSchedule uint64 = 10

	V1SaveState = SaveStateOS{}
)

const (
	// If true uuid V1 will save state in a temp dir
	V1Save = true
)

func init() {
	if V1Save {
		gob.Register(stateEntity{})
	}
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

	// save state interface
	SaveState
}

// Changes the state with current data
// Compares the current found node to the last node stored,
// If they are the same or randomSequence is already set due
// to an earlier read issue then the sequence is randomly generated
// else if there is an issue with the time the sequence is incremeted
func (o *State) read(pNow Timestamp, pNode net.HardwareAddr) {
	if bytes.Equal(pNode, o.node) || o.randomSequence {
		o.sequence = uint16(seed.Int())&0x3FFF
	} else if pNow < o.past {
		o.sequence ++
	}
	o.past = pNow
	o.node = pNode
}

func (o *State) persist() {
	if V1Save {
		if !V1SaveState.Setup() {
			V1SaveState.Init(o)
		}
		o.Save(o)
	}
}

// Initialises the UUID state when the package is first loaded
// it first attempts to decode the file state into State
// if this file does not exist it will create the file and do a flush
// of the random state which gets loaded at package runtime
// second it will attempt to resolve the current hardware address nodeId
// thirdly it will check the state of the clock
func (o *State) init() {
	if V1Save {
		o.randomSequence = false
	}
	intfcs, err := net.Interfaces()
	if err != nil {
		log.Println("UUID.State.init: address error:", err)
		return
	}
	a := getHardwareAddress(intfcs)
	if a == nil {
		log.Println("UUID.State.init: address error:", err)
		return
	}
	if bytes.Equal(a, state.node) {
		state.sequence ++
	}
	state.node = a
	state.randomNode = false
}

// ***********************************************  StateEntity

// StateEntity acts as a marshaller struct for the state
type stateEntity struct {
	Past       Timestamp
	Node       []byte
	Sequence   uint16
}

type SaveState interface {
	// Init is run if Setup() is false
	// Init should setup the system to save the state
	Init(*State)

	// Save saves the state and is called only if const V1Save and
	// Setup() is true
	Save(*State)

	// Should return whether Saving has been initialised.
	Setup() bool
}

type SaveStateOS struct {
	cache     *os.File
	saveState uint64
	setup     bool
}

func (o *SaveStateOS) Setup() bool {
	return o.setup
}

// Saves the current state of the generator
// If the scheduled file save is reached then the file is synced
func (o *SaveStateOS) Save(pState *State) {
	if pState.past >= pState.next {
		err := o.open()
		defer o.cache.Close()
		if err != nil {
			log.Println("UUID.State.save:", err)
			return
		}
		// do the save
		o.encode(pState)
		schedule := Timestamp(SaveSchedule)
		// default: schedule next save for 10 seconds from now
		pState.next = pState.past+schedule*ticksPerSecond
		if Report {
			log.Printf("UUID STATE: SAVED %d", pState.past)
		}
	}
}

func (o *SaveStateOS) Init(pState *State) {
	pState.SaveState = o
	err := o.open()
	defer o.cache.Close()
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("'%s' created\n", "UUID.SaveState")
			var err error
			o.cache, err = os.Create(os.TempDir()+"/state.unique")
			if err != nil {
				log.Println("UUID.State.init: SaveState error:", err)
				goto pastInit
			}
			o.encode(pState)
		} else {
			log.Println("UUID.State.init: SaveState error:", err)
			goto pastInit
		}
	}
	err = o.decode(pState)
	if err != nil {
		goto pastInit
	}
	pState.randomSequence = false
pastInit:
	if timestamp() <= pState.past {
		pState.sequence ++
	}
	pState.next = state.past
	o.setup = true
}

func (o *SaveStateOS) reset() {
	o.cache.Seek(0, 0)
}

func (o *SaveStateOS) open() error {
	var err error
	o.cache, err = os.OpenFile(os.TempDir()+"/state.unique", os.O_RDWR, os.ModeExclusive)
	return err
}

// Encodes State generator data into a saved file
func (o *SaveStateOS) encode(pState *State) {
	// ensure reader state is ready for use
	o.reset()
	enc := gob.NewEncoder(o.cache)
	// Wrap private State data into the StateEntity
	err := enc.Encode(&stateEntity{pState.past, pState.node, pState.sequence})
	if err != nil {
		log.Panic("UUID.encode error:", err)
	}
}

// Decodes StateEntity data into the main State
func (o *SaveStateOS) decode(pState *State) error {
	o.reset()
	dec := gob.NewDecoder(o.cache)
	entity := stateEntity{}
	err := dec.Decode(&entity)
	if err != nil {
		log.Println("UUID.decode error:", err)
		return err
	}
	pState.past = entity.Past
	pState.node = entity.Node
	pState.sequence = entity.Sequence
	return nil
}
