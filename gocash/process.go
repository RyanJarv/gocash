package gocash

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// NewProcess returns *Process instance
func NewProcess(svr *Server, cache *Cache) *Process {
	p := &Process{
		sess:     svr.Sessions,
		cache:    cache,
		maxProcs: make(chan int, 4),
	}
	return p
}

// Process takes lines from *Server and runs against *Cache
type Process struct {
	sess     chan *net.Conn
	cache    *Cache
	maxProcs chan int
}

// Start processing input from the server
func (p *Process) Start() {
	p.maxProcs = make(chan int, 4)
	for {
		p.maxProcs <- 1
		sess := <-p.sess
		go p.handle(sess)
	}
}

func (p *Process) handle(conn *net.Conn) {
	scan := bufio.NewScanner(*conn)
	for scan.Scan() {
		cmd, args, err := p.parse(scan.Text())
		if err != nil {
			fmt.Fprintln(*conn, "ERR", err)
			return
		}
		ret, err := p.cache.Call(cmd, args...)
		if err != nil {
			fmt.Fprintln(*conn, "ERR", err)
			continue
		}
		fmt.Fprintln(*conn, ret)
	}
	<-p.maxProcs
}

func (p *Process) parse(line string) (string, []string, error) {
	scan := bufio.NewScanner(strings.NewReader(line))
	scan.Split(bufio.ScanWords)

	scan.Scan()
	cmd := scan.Text()

	var args []string
	for scan.Scan() {
		args = append(args, scan.Text())
	}

	if scan.Err() != nil {
		return "", nil, scan.Err()
	}

	return cmd, args, nil
}
