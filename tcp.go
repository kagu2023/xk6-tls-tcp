package tcpclient

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"

	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/tcpclient", new(TcpClient))
}

type TcpClient struct{}

func (client *TcpClient) Connect(insecureSkipVerify bool, addr string, enableReadCRLF bool) (*TcpConnection, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return nil, err
	}
	return NewTcpConnection(conn, enableReadCRLF), nil
}

type TcpConnection struct {
	rawConnection *tls.Conn
	reader        *bufio.Reader
	scanner       *bufio.Scanner
	writer        *bufio.Writer
}

func NewTcpConnection(rawConn *tls.Conn, enableReadCRLF bool) *TcpConnection {
	var scanner *bufio.Scanner = nil
	if enableReadCRLF {
		splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			// we are at EOF and there is no data.
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}

			// Search for a CRLF.
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				if i > 0 && data[i-1] == '\r' {
					return i + 1, data[0 : i-1], nil
				}
			}

			// If we're at EOF, we have a final, non-terminated line. Return it.
			if atEOF {
				return len(data), data, nil
			}

			// Request more data.
			return 0, nil, nil
		}

		scanner = bufio.NewScanner(rawConn)
		scanner.Split(splitFunc)
	}

	return &TcpConnection{
		rawConnection: rawConn,
		reader:        bufio.NewReader(rawConn),
		scanner:       scanner,
		writer:        bufio.NewWriter(rawConn),
	}
}

func (conn *TcpConnection) Close() error {
	return conn.rawConnection.Close()
}

func (conn *TcpConnection) Write(data []byte) error {
	_, err := conn.rawConnection.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (conn *TcpConnection) Read(data []byte) (int, error) {
	cnt, err := conn.rawConnection.Read(data)
	if err != nil {
		return cnt, err
	}
	return cnt, nil
}

func (conn *TcpConnection) WriteStringCRLine(str string) (int, error) {
	cnt, err := conn.writer.WriteString(str)
	if err != nil {
		return cnt, err
	}

	cntNewLine, err := conn.writer.WriteString("\r\n")
	cnt += cntNewLine
	if err != nil {
		return cnt, err
	}

	err = conn.writer.Flush()
	if err != nil {
		return cnt, err
	}

	return cnt, nil
}

func (conn *TcpConnection) ReadCRLine() (string, error) {
	if conn.scanner == nil {
		return "", fmt.Errorf("ReadCRLine is not enabled")
	}

	if conn.scanner.Scan() {
		return conn.scanner.Text(), nil
	}

	err := conn.scanner.Err()
	if err != nil {
		return "", err
	}

	return conn.scanner.Text(), err
}

func (conn *TcpConnection) ReadLine() (string, error) {
	isPrefix := true
	result := ""
	for {
		if !isPrefix {
			break
		}

		bytes, curIsPrefix, err := conn.reader.ReadLine()
		if err != nil {
			return result, err
		}

		result += string(bytes)
		isPrefix = curIsPrefix
	}

	return result, nil
}
