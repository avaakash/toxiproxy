package toxics

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/Shopify/toxiproxy/v2/stream"
)

type HttpResponseHeaderToxic struct {
	HeaderKey   string `json:"header_key"`
	HeaderValue string `json:"header_value"`
}

func (t *HttpResponseHeaderToxic) ModifyResponseHeader(resp *http.Response) {
	resp.Header.Set(t.HeaderKey, t.HeaderValue)
}

func (t *HttpResponseHeaderToxic) Pipe(stub *ToxicStub) {
	buffer := bytes.NewBuffer(make([]byte, 0, 32*1024))
	writer := stream.NewChanWriter(stub.Output)
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)
	for {
		tee := io.TeeReader(reader, buffer)
		resp, err := http.ReadResponse(bufio.NewReader(tee), nil)

		if err == stream.ErrInterrupted {
			buffer.WriteTo(writer)
			return
		} else if err == io.EOF {
			stub.Close()
			return
		}
		if err != nil {
			buffer.WriteTo(writer)
		} else {
			t.ModifyResponseHeader(resp)
			resp.Write(writer)
		}
		buffer.Reset()
	}
}

func init() {
	Register("http_response_header", new(HttpResponseHeaderToxic))
}
