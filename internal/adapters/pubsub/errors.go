package pubsub

import "errors"

// ErrPubSubClosed is returned when attempting to publish or subscribe
// after the pubsub broker has been closed.
var ErrPubSubClosed = errors.New("pubsub: broker is closed")
