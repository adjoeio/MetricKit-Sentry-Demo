package main

type IOSMXCrashDiagnosticRequest struct {
	CallStackTree      MXCallStackTree    `json:"callStackTree"`
	DiagnosticMetaData DiagnosticMetaData `json:"diagnosticMetaData"`
}

type MXCallStackTree struct {
	CallStacks         []MXCallStack `json:"callStacks"`
	CallStackPerThread bool          `json:"callStackPerThread"`
}

type MXCallStack struct {
	CallStackRootFrames []MXCallStackFrame `json:"callStackRootFrames"`
	ThreadAttributed    bool               `json:"threadAttributed"`
}

type MXCallStackFrame struct {
	BinaryUUID                  string             `json:"binaryUUID"`
	BinaryName                  string             `json:"binaryName"`
	SubFrames                   []MXCallStackFrame `json:"subFrames"`
	OffsetIntoBinaryTextSegment int64              `json:"offsetIntoBinaryTextSegment"`
	Address                     int64              `json:"address"`
}

type DiagnosticMetaData struct {
	ExceptionType int `json:"exceptionType"`
	ExceptionCode int `json:"exceptionCode"`
	Signal        int `json:"signal"`
}
