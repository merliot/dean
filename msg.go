package dean

import (
	"fmt"
	"strings"
)

type message struct {

	// Tags is a dotted-list of tags, i.e. xxx.yyy.zzz.  Empty Tags means
	// the message has reached its destination.  A Tag is added to the
	// message as the message ingresses a tagged connection.  Each ingress
	// into a tagged connection adds another tag.  The tags are dotted
	// together, with the left-most tag being the tag of the last
	// connection.  Likewise, on egress of tagged a connection, the
	// left-most tag is stripped from Tags.
	//
	// A tag in Tags should pass ValidId().

	Tags string

	// Path identifies the message type.  Paths are defined by the
	// application.  There are a few reserved Paths:
	//
	//     "ping", "pong", "get/state", "state", "online", "offline"

	Path string

	// Payload is the message payload

	Payload []byte
}

// Bytes returns the message payload
func (m *message) Bytes() []byte {
	return []byte(m.Payload)
}

func (m *message) String() string {
	return fmt.Sprintf("[%s:%s] %s", m.Tags, m.Path, string(m.Payload))
}

func (m *message) PathEqual(m2 *message) bool {
	return m.Path == m2.Path
}

// popTag pops the first tag off the message tag chain and returns the tag
func (m *message) popTag() string {
	if m.Tags == "" {
		return ""
	}
	tags := strings.SplitN(m.Tags, ".", 2)
	switch len(tags) {
	case 1:
		m.Tags = ""
	case 2:
		m.Tags = tags[1]
	}
	return tags[0]
}
