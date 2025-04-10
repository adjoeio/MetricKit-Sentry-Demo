package main

import (
	"encoding/json"
	"fmt"
)

type SentryEvent struct {
	Platform string `json:"platform"`
	Level    string `json:"level"`

	Exception Exception `json:"exception"`
	Threads   Threads   `json:"threads"`
	DebugMeta DebugMeta `json:"debug_meta"`

	Timestamp int `json:"timestamp"`
}

type DebugMeta struct {
	Images []Image `json:"images"`
}

type Image struct {
	DebugID   string `json:"debug_id"`
	Type      string `json:"type"`
	ImageAddr Hex    `json:"image_addr"`
}

type Stacktrace struct {
	Frames []Frame `json:"frames"`
}

type Frame struct {
	Package         string `json:"package"`
	ImageAddr       Hex    `json:"image_addr"`
	InstructionAddr Hex    `json:"instruction_addr"`
	InApp           bool   `json:"in_app"`
}

type Exception struct {
	Values []Values `json:"values"`
}

type Values struct {
	Type       string     `json:"type"`
	Value      string     `json:"value"`
	Stacktrace Stacktrace `json:"stacktrace"`
	Mechanism  Mechanism  `json:"mechanism"`
	ThreadID   int        `json:"thread_id"`
}

type Threads struct {
	Values []ThreadValue `json:"values"`
}

type ThreadValue struct {
	Stacktrace Stacktrace `json:"stacktrace"`
	ID         int        `json:"id"`
	Crashed    bool       `json:"crashed"`
}

type Mechanism struct {
	Type    string `json:"type"`
	Meta    Meta   `json:"meta"`
	Handled bool   `json:"handled"`
}

type Meta struct {
	Signal        *Signal        `json:"signal,omitempty"`         // Information on the POSIX signal.
	MachException *MachException `json:"mach_exception,omitempty"` // A Mach Exception on Apple systems comprising a code triple and optional descriptions.
	NSError       *NSError       `json:"ns_error,omitempty"`       // An NSError on Apple systems comprising domain and code.
	ErrNo         *ErrNo         `json:"errno,omitempty"`          // Error codes set by Linux system calls and some library functions.
}

type Signal struct {
	Number   int     `json:"number"`              // POSIX signal number
	Code     *int    `json:"code,omitempty"`      // Optional Apple signal code
	Name     *string `json:"name,omitempty"`      // Optional name of the signal based on the signal number.
	CodeName *string `json:"code_name,omitempty"` // Optional name of the signal code.
}

type MachException struct {
	Code      int     `json:"code"`           // Required numeric exception code.
	SubCode   int64   `json:"subcode"`        // Required numeric exception subcode.
	Exception int     `json:"exception"`      // Required numeric exception number.
	Name      *string `json:"name,omitempty"` // Optional name of the exception constant in iOS / macOS.
}

type NSError struct {
	Code   int    `json:"code"`   // Required numeric error code.
	Domain string `json:"domain"` // Required domain of the NSError as string.
}

type ErrNo struct {
	Number int     `json:"number"`         // The error number
	Name   *string `json:"name,omitempty"` // Optional name of the error
}

type Hex int64

func (h Hex) String() string {
	return fmt.Sprintf("0x%016x", int64(h))
}

func (h Hex) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}
