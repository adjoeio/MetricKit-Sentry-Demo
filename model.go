package main

import (
	"errors"
	"strings"
)

type SentryCrashStackTree struct {
	Threads       []Thread
	CrashedThread Thread
}

type Thread struct {
	StackFrames []StackFrame
	ID          int
}

type StackFrame struct {
	Binary Binary

	IOSAddress         int64
	SentryImageAddress int64

	InApp bool
}

type Binary struct {
	UUID string
	Name string
}

func (t *SentryCrashStackTree) FromMXCrashDiagnostics(mxCrashDiagnostics *MXCallStackTree) error {
	// According to the Sentry Apple/iOS SDK, the CallStackPerThread flag is always true
	// for crashes. Since we currently don't handle other types of MetricKit Diagnostics,
	// we omit other CrashDiagnostics here and log a warning
	// See: https://github.com/getsentry/sentry-cocoa/blob/0ffc3c62/Sources/Sentry/SentryMetricKitIntegration.m#L209-L214

	if !mxCrashDiagnostics.CallStackPerThread {
		return errors.New("not implemented for multiple callstacks per thread")
	}
	for _, mxCallStacks := range mxCrashDiagnostics.CallStacks {
		if len(mxCallStacks.CallStackRootFrames) > 1 {
			return errors.New("not implemented for multiple callstacks per thread")
		}
	}

	if t == nil {
		*t = SentryCrashStackTree{}
	}

	crashedThreadIdx := -1
	t.Threads = make([]Thread, 0, len(mxCrashDiagnostics.CallStacks))
	for idx, mxCallStacks := range mxCrashDiagnostics.CallStacks {
		stackFrames := unnestCallStack(mxCallStacks.CallStackRootFrames)

		t.Threads = append(t.Threads, Thread{
			StackFrames: stackFrames,
			ID:          idx,
		})

		if mxCallStacks.ThreadAttributed && crashedThreadIdx == -1 {
			crashedThreadIdx = idx
		}
	}

	// If we don't have a thread which is attributed to the crash, we just take the first one
	if crashedThreadIdx < 0 {
		crashedThreadIdx = 0
	}

	t.CrashedThread = t.Threads[crashedThreadIdx]
	return nil
}

func unnestCallStack(callStack []MXCallStackFrame) []StackFrame {
	if len(callStack) == 0 {
		return nil
	}

	frame := callStack[0]

	ret := unnestCallStack(frame.SubFrames)
	ret = append(ret,
		StackFrame{
			Binary: Binary{
				UUID: frame.BinaryUUID,
				Name: frame.BinaryName,
			},
			IOSAddress:         frame.Address,
			SentryImageAddress: frame.Address - frame.OffsetIntoBinaryTextSegment,
			InApp:              strings.Contains(frame.BinaryName, "MonetizeSDK"),
		},
	)

	return ret
}

func (t *SentryCrashStackTree) Images() []Image {
	images := make(map[Image]struct{})
	for _, thread := range t.Threads {
		for _, frame := range thread.StackFrames {
			image := Image{
				DebugID:   frame.Binary.UUID,
				Type:      "macho",
				ImageAddr: Hex(frame.SentryImageAddress),
			}
			images[image] = struct{}{}
		}
	}

	ret := make([]Image, 0, len(images))
	for image := range images {
		ret = append(ret, image)
	}

	return ret
}
