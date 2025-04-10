package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

var (
	sentryURL  string
	httpClient http.Client
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	viper.AddConfigPath(wd)
	viper.SetConfigType("yaml")
	viper.SetConfigName("env")

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	sentryURL = NewSentryURL().String()
	httpClient = http.Client{
		Timeout: 5 * time.Second,
	}

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	http.HandleFunc("/mx-crash-diagnostic", handleCrashEvent)
	log.Fatal(server.ListenAndServe())
}

func handleCrashEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := IOSMXCrashDiagnosticRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("failed to unmarshall request: %v", err)
		return
	}

	sentryCrashStackTree := SentryCrashStackTree{}
	if err := sentryCrashStackTree.FromMXCrashDiagnostics(&request.CallStackTree); err != nil {
		log.Printf("failed to parse call stack tree: %v", err)
		return
	}

	sentryRequest := SentryEvent{
		Timestamp: int(time.Now().UTC().Unix()),
		Platform:  "cocoa",
		Level:     "fatal",
		Exception: Exception{
			Values: []Values{
				{
					Type:  "MXCrashDiagnostic",
					Value: "MetricKit > MXDiagnostic > Crash in SDK",
					Mechanism: Mechanism{
						Handled: false,
						Meta: Meta{
							Signal: &Signal{
								Number: request.DiagnosticMetaData.Signal,
							},
							MachException: &MachException{
								Code:      request.DiagnosticMetaData.ExceptionCode,
								SubCode:   sentryCrashStackTree.CrashedThread.StackFrames[0].IOSAddress,
								Exception: request.DiagnosticMetaData.ExceptionType,
							},
						},
						Type: "MXCrashDiagnostic",
					},
					ThreadID: sentryCrashStackTree.CrashedThread.ID,
					Stacktrace: Stacktrace{
						Frames: make([]Frame, 0, len(sentryCrashStackTree.CrashedThread.StackFrames)),
					},
				},
			},
		},
		DebugMeta: DebugMeta{
			Images: sentryCrashStackTree.Images(),
		},
		Threads: Threads{
			Values: make([]ThreadValue, 0, len(sentryCrashStackTree.Threads)),
		},
	}

	for _, thread := range sentryCrashStackTree.Threads {
		val := ThreadValue{
			ID:      thread.ID,
			Crashed: sentryCrashStackTree.CrashedThread.ID == thread.ID,
		}

		val.Stacktrace.Frames = make([]Frame, 0, len(thread.StackFrames))
		for _, frame := range thread.StackFrames {
			f := Frame{
				Package:         frame.Binary.Name,
				InApp:           frame.InApp,
				ImageAddr:       Hex(frame.SentryImageAddress),
				InstructionAddr: Hex(frame.IOSAddress),
			}

			val.Stacktrace.Frames = append(val.Stacktrace.Frames, f)
		}

		sentryRequest.Threads.Values = append(sentryRequest.Threads.Values, val)
	}
	for _, frame := range sentryCrashStackTree.CrashedThread.StackFrames {
		f := Frame{
			Package:         frame.Binary.Name,
			InApp:           frame.InApp,
			ImageAddr:       Hex(frame.SentryImageAddress),
			InstructionAddr: Hex(frame.IOSAddress),
		}
		sentryRequest.Exception.Values[0].Stacktrace.Frames = append(sentryRequest.Exception.Values[0].Stacktrace.Frames, f)
	}

	envelopeItemPayload, err := json.Marshal(&sentryRequest)
	if err != nil {
		log.Printf("failed to marshall request: %v", err)
		return
	}

	envelopeHeader := NewEnvelopeHeader()
	envelopeItemHeader := NewItemHeader(len(envelopeItemPayload))

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err = encoder.Encode(&envelopeHeader); err != nil {
		log.Printf("failed to marshall envelope header: %v", err)
		return
	}
	if err = encoder.Encode(&envelopeItemHeader); err != nil {
		log.Printf("failed to marshall envelope header: %v", err)
		return
	}
	buf.Write(envelopeItemPayload)

	req, err := http.NewRequestWithContext(ctx, "POST", sentryURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		log.Printf("failed to build request: %v", err)
		return
	}
	req.Header.Add("Content-Type", "application/x-sentry-envelope")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("failed to send event to Sentry: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to send event to Sentry with status code: %v", resp.StatusCode)
		if body, err := io.ReadAll(resp.Body); err == nil {
			log.Printf("error response: %s", body)
		}
		return
	}

	log.Println("successfully sent event to Sentry:")
	log.Println(buf.String())
	w.WriteHeader(http.StatusOK)
}
