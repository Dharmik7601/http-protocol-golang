package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Dharmik7601/http-protocol-golang/internal/headers"
	"github.com/Dharmik7601/http-protocol-golang/internal/request"
	"github.com/Dharmik7601/http-protocol-golang/internal/response"
	"github.com/Dharmik7601/http-protocol-golang/internal/server"
)

const port = 8888

func responsd400() []byte {
	return []byte(`<html>
		<head>
			<title>400 Bad Request</title>
		</head>
		<body>
			<h1>Bad Request</h1>
			<p>Your request honestly kinda sucked.</p>
		</body>
	</html>`)
}

func responsd500() []byte {
	return []byte(`<html>
		<head>
			<title>500 Internal Server Error</title>
		</head>
		<body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		</body>
	</html>`)
}

func responsd200() []byte {
	return []byte(`<html>
		<head>
			<title>200 OK</title>
		</head>
		<body>
			<h1>Success!</h1>
			<p>Your request was an absolute banger.</p>
		</body>
	</html>`)
}

func toString(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%x02", b)
	}

	return out
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeader(0)
		body := responsd200()
		status := response.StatusOK

		if req.RequestLine.RequestTarget == "/yourproblem" {
			status = response.StatusBadRequest
			body = responsd400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			body = responsd500()
		} else if req.RequestLine.RequestTarget == "/video" {
			f, _ := os.ReadFile("assets/super_heavy.mp4")
			h.Replace("Content-length", fmt.Sprintf("%d", len(f)))
			h.Replace("Content-type", "video/mp4")

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(f)
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = responsd500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOK)

				h.Delete("Content-length")
				h.Set("transfer-encoding", "chunked")
				h.Replace("Content-type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(h)

				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))

				trailer := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailer.Set("X-Content-SHA256", toString(out[:]))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

				w.WriteHeaders(trailer)

				return
			}
		}

		h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
