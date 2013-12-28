package pipeline

import (
	"github.com/funkygao/als"
)

// Main Heka pipeline data structure containing raw message data, a Message
// object, and other Heka related message metadata.
type PipelinePack struct {
	// Used for storage of binary blob data that has yet to be decoded into a
	// Message object.
	MsgBytes []byte
	// Main Heka message object.

	Message *als.Message
	// Specific channel on which this pack should be recycled when all
	// processing has completed for a given message.
	RecycleChan chan *PipelinePack
	// Indicates whether or not this pack's Message object has been populated.
	Decoded bool
	// Reference count used to determine when it is safe to recycle a pack for
	// reuse by the system.
	RefCount int32
	// String id of the verified signer of the accompanying Message object, if
	// any.
	Signer string
	// Number of times the current message chain has generated new messages
	// and inserted them into the pipeline.
	MsgLoopCount uint
	// Used internally to stamp diagnostic information onto a packet
	diagnostics *PacketTracking
}
