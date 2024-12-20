package internal

import (
	"io"

	"github.com/epicchainlabs/epicchain-sdk-go/object"
	oid "github.com/epicchainlabs/epicchain-sdk-go/object/id"
)

// HeaderWriter is an interface of target component
// to write object header.
type HeaderWriter interface {
	// WriteHeader writes object header w/ payload part.
	// The payload of the object may be incomplete.
	//
	// Must be called exactly once. Control remains with the caller.
	// Missing a call or re-calling can lead to undefined behavior
	// that depends on the implementation.
	//
	// Must not be called after Close call.
	WriteHeader(*object.Object) error
}

// Target is an interface of the object writer.
type Target interface {
	HeaderWriter
	// Write writes object payload chunk.
	//
	// Can be called multiple times.
	//
	// Must not be called after Close call.
	io.Writer

	// Close is used to finish object writing.
	//
	// Close must return ID of the object
	// that has been written.
	//
	// Must be called no more than once. Control remains with the caller.
	// Re-calling can lead to undefined behavior
	// that depends on the implementation.
	Close() (oid.ID, error)
}
