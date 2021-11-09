package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"
	"time"
)

var validNum = regexp.MustCompile(`[^0-9]`)

type Unique struct {
	sync.Mutex
	numbers map[string]bool
}

func (u *Unique) add(name string) {
	u.Lock()
	defer u.Unlock()
	u.numbers[name] = true
}

var u = Unique{numbers: map[string]bool{}}

// need a mutex to make counting thread safe
type Counter struct { 
	sync.Mutex
	counters map[string]int
}

func (c *Counter) inc(name string) {
	c.Lock()
	defer c.Unlock()
	c.counters[name]++
}

func (c *Counter) dec(name string) {
	c.Lock()
	defer c.Unlock()
	c.counters[name]--
}

func (c *Counter) reset(name string) {
	c.Lock()
	defer c.Unlock()
	c.counters[name] = 0
}

func main() {
	addr := "127.0.0.1:4000"
	serve(addr)
}

func serve(addr string) {
	server, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	defer server.Close()
	fmt.Println("listening on ", addr)

	// open a log file
	file, err := os.Create("numbers.log")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// running report
	c := Counter{
		counters: map[string]int{
			"clients": 0,
			"unique": 0,
			"duplicate": 0,
			"total": 0,
		},
	}
	go func() {
		for {
			fmt.Printf("Recieved %v unique numbers, %v duplicates. Unique Total: %v\n", c.counters["unique"], c.counters["duplicate"], c.counters["total"])
			c.reset("unique")
			c.reset("duplicate")
			time.Sleep(time.Second * 10)
		}
	}()

	// accept connections
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("no longer accepting connections")
			return
		}
		// pass log file writer to handler
		go handle(conn, file, server, &c)
	}
}

func handle(conn net.Conn, file *os.File, server net.Listener, c *Counter) {
	// close if too many clients
	if c.counters["clients"] >= 5 {
		conn.Close()
		return
	}
	c.inc("clients")

	// read line
	buf, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		conn.Close()
		c.dec("clients")
		return
	}

	// if client sends terminate, close server
	if string(buf[:len(buf)-2]) == "terminate" {
		server.Close()
		return
	}

	// validate its a 9 digits and nothing else
	// len(buf)-2 because of return characters at the end
	if len(buf)-2 != 9 {
		conn.Close()
		c.dec("clients")
		return
	}

	// regex finds any non 0-9
	if validNum.Match(buf[:len(buf)-2]) {
		conn.Close()
		c.dec("clients")
		return
	}

	// search slice and make sure number doesn't already exist
	// abusing system memory for brevity and because I'm not sure if reading a file is thread safe by default
	_, ok := u.numbers[string(buf[:len(buf)-2])] 
	if !ok {
		u.add(string(buf[:len(buf)-2]))
		_, err = file.Write(buf)
		if err != nil {
			conn.Close()
			c.dec("clients")
			return
		}
		c.inc("total")
		c.inc("unique")
	} else {
		c.inc("duplicate")
	}

	// handle next line of input
	handle(conn, file, server, c)
}
