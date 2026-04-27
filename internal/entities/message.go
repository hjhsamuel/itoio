package entities

type ClientAttributes int

const (
	ClientAttributesNone ClientAttributes = iota // default
	ClientAttributeP2P                           // p2p
)

type P2PAttributes int

const (
	P2PDefault      P2PAttributes = iota // default
	P2PShareScreen                       // share screen
	P2PFile                              // upload and download file
	P2PBeControlled                      // be controlled remotely
)
