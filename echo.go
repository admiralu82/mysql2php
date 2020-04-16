package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// set GOARCH=amd64
// set GOOS=linux

var port int = 4000
var scriptURL = "http://localhost/echo.php"

func ReadCFG() {

	type cfg struct {
		Port      int
		ScriptURL string
	}
	c := cfg{4000, "http://localhost/echo.php"}

	if TestFile("echo.cfg") {
		buf, _ := ioutil.ReadFile("echo.cfg")
		json.Unmarshal(buf, &c)
	} else {
		buf, _ := json.MarshalIndent(c, " ", " ")
		ioutil.WriteFile("echo.cfg", buf, os.ModePerm)
	}

	port = c.Port
	scriptURL = c.ScriptURL

	fmt.Println("Start config: Port=", port, "(tunnel)", port+1, "(SQL TCP)", " scriptUrl=", scriptURL)
}

func TestFile(file string) bool {
	fi, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func main() {

	ReadCFG()

	rand.Seed(time.Now().UnixNano())

	addrc, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(port))
	addrm, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(port+1))

	lc, errc := net.ListenTCP("tcp4", addrc)
	lm, errm := net.ListenTCP("tcp4", addrm)

	if errc != nil || errm != nil {
		fmt.Println("Error master listening:", errm, errc)
		os.Exit(1)
	}

	for {
		fmt.Println("M: Waiting Master clients...")
		connm, err := lm.Accept()
		if err != nil {
			fmt.Println("M: Error accepting ", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}

		p := Pair{}
		p.Init()
		go p.SendRequestInit()

		lc.SetDeadline(time.Now().Add(5 * time.Second))
		connc, err := lc.Accept()
		if err != nil {
			myerr, ok := err.(net.Error)
			if ok && myerr.Timeout() {
				connm.Close()
				continue
			}
			fmt.Println("Error accepting client: ", err.Error())
			connm.Close()
			continue
		}

		p.AddConn(&connm, &connc)
		if p.ReadAuth() == false {
			fmt.Println("Auth bad.")
			p.Close()
			continue
		}
		go p.Loop()
	}
}

type Pair struct {
	sync.Mutex
	id     byte
	cm, cc *net.Conn

	ctx    context.Context
	cancel context.CancelFunc
}

func (p *Pair) SendRequestInit() {
	// просим клиента подключиться
	strUrl := scriptURL + "?ID=" + strconv.Itoa(int(p.id))
	fmt.Println("AuthScript", strUrl)
	resp, err := http.Get(strUrl)
	if err != nil {
		fmt.Println("AuthScript Error SendRequestInit", err.Error())
	} else {
		buf, _ := ioutil.ReadAll(resp.Body)
		ioutil.WriteFile("echo.log", buf, os.ModePerm)
	}
	p.cancel()
	fmt.Println("AuthScript stoped...")
}

func (p *Pair) ReadAuth() bool {
	idRead := make([]byte, 1)
	l, err := (*p.cc).Read(idRead)
	if l != 1 || err != nil {
		fmt.Println("ReadAuth Error read id client")
		return false
	}
	if l == 1 && idRead[0] == p.id {
		return true
	}
	return false
}

func (p *Pair) Loop() {
	var toM, toC int
	toM = 0
	toC = 0

	go func() {

		buf := make([]byte, 10000)
		for {
			(*p.cc).SetReadDeadline(time.Now().Add(1 * time.Second))
			l, err := (*p.cc).Read(buf)
			if err != nil {
				myerr, ok := err.(net.Error)
				if ok && myerr.Timeout() {
					continue
				}
				if _, ok := <-p.ctx.Done(); ok == false {
					break
				}

				fmt.Println("Read toM error", err.Error())
				break
			}

			l, err = (*p.cm).Write(buf[:l])
			if err != nil {
				fmt.Println("Write toM error", err.Error())
				break
			}
			toM += l
			//fmt.Println("Write toM", l, "bytes")
		}

		p.cancel()
	}()
	go func() {

		buf := make([]byte, 10000)
		for {

			(*p.cm).SetReadDeadline(time.Now().Add(1 * time.Second))
			l, err := (*p.cm).Read(buf)
			if err != nil {
				myerr, ok := err.(net.Error)
				if ok && myerr.Timeout() {
					continue
				}
				if _, ok := <-p.ctx.Done(); ok == false {
					break
				}
				fmt.Println("Read toC error", err.Error())
				break
			}

			l, err = (*p.cc).Write(buf[:l])
			if err != nil {
				fmt.Println("Write toC error", err.Error())
				break
			}
			toC += l

			//fmt.Println("Write toC", l, "bytes")
		}
		p.cancel()
	}()

	<-p.ctx.Done()
	fmt.Println("Done ", p.id, "To Master ", toM, "bytes, To Client ", toC, "bytes;")
	p.Close()
}

func (p *Pair) Close() {
	(*p.cm).Close()
	(*p.cc).Close()
}

func (p *Pair) Init() {
	p.id = byte(rand.Int() % 256)

	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx
	p.cancel = cancel
}

func (p *Pair) AddConn(cm, cc *net.Conn) {
	p.cm = cm
	p.cc = cc
}
