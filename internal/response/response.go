package response

import (
	"fmt"
	"io"

	"github.com/Dharmik7601/http-protocol-golang/internal/headers"
)

const HTTP_PROTOCOL = "HTTP/1.1"

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeader(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte(fmt.Sprintf("%s 200 OK\r\n", HTTP_PROTOCOL))
	case StatusBadRequest:
		statusLine = []byte(fmt.Sprintf("%s 400 Bad Request\r\n", HTTP_PROTOCOL))
	case StatusInternalServerError:
		statusLine = []byte(fmt.Sprintf("%s 500 Internal Server Error\r\n", HTTP_PROTOCOL))
	default:
		return fmt.Errorf("unrecognized error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(header *headers.Headers) error {
	b := []byte{}
	header.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}
