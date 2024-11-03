package internal

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

func HijackHttpConnection(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	var conn net.Conn
	var bufrw *bufio.ReadWriter

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Web server does not support hijacking", http.StatusInternalServerError)
		return conn, bufrw, errors.New("web server does not support hijacking")
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return conn, bufrw, err
	}

	return conn, bufrw, nil
}
