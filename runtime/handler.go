package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
)

type responseStreamChunk struct {
	Result proto.Message `json:"result,omitempty"`
	Error  string        `json:"error,omitempty"`
}

// ForwardResponseStream forwards the stream from gRPC server to REST client.
func ForwardResponseStream(w http.ResponseWriter, recv func() (proto.Message, error)) {
	w.WriteHeader(http.StatusOK)
	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			buf, merr := json.Marshal(responseStreamChunk{Error: err.Error()})
			if merr != nil {
				glog.Error("Failed to marshal an error: %v", merr)
				return
			}
			if _, werr := fmt.Fprintln(w, buf); werr != nil {
				glog.Error("Failed to notify error to client: %v", werr)
				return
			}
			return
		}
		buf, err := json.Marshal(responseStreamChunk{Result: resp})
		if err != nil {
			glog.Error("Failed to marshal response chunk: %v", err)
			return
		}
		if _, err = fmt.Fprintln(w, buf); err != nil {
			glog.Error("Failed to send response chunk: %v", err)
			return
		}
	}
}

// ForwardResponseStream forwards the message from gRPC server to REST client.
func ForwardResponseMessage(w http.ResponseWriter, resp proto.Message) {
	buf, err := json.Marshal(resp)
	if err != nil {
		glog.Errorf("Marshal error: %v", err)
		HTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(buf); err != nil {
		glog.Errorf("Failed to write response: %v", err)
	}
}