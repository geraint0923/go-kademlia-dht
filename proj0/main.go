package main

import (
	"bufio"
	"net"
	"strconv"
	"strings"
)

type ClientMsg struct {
	SrcClient int
	DstClient int
	Msg       string
}

func NewClienetMsg(src int, msg string) *ClientMsg {
	res := &ClientMsg{
		SrcClient: src,
		DstClient: -1,
		Msg:       msg,
	}
	res.Build()
	return res
}

func (msg *ClientMsg) Build() {
	idx := strings.Index(msg.Msg, ":")
	if idx != -1 {
		cmd := msg.Msg[0:idx]
		cmd = strings.TrimSpace(cmd)
		if len(cmd) > 0 {
			if cmd == "whoami" {
				msg.DstClient = msg.SrcClient
				msg.Msg = "chitter: " + strconv.Itoa(msg.SrcClient) + "\n"
			} else if cmd == "all" {
				msg.Msg = strconv.Itoa(msg.SrcClient) + ": " + msg.Msg[idx+1:]
			} else {
				iv, err := strconv.Atoi(cmd)
				if err == nil {
					msg.DstClient = iv
					msgValue := msg.Msg[idx+1:]
					msg.Msg = strconv.Itoa(msg.SrcClient) + ": " + msgValue
				} else {
					msg.DstClient = -2
				}
			}
		}
	} else {
		msg.Msg = strconv.Itoa(msg.SrcClient) + ": " + msg.Msg
	}
}

const CTLADD = 1
const CTLDEL = 2

type Client struct {
	Opcode int
	Id     int
	Conn   net.Conn
}

func HandleConn(ctlCh chan Client, msgCh chan *ClientMsg, cli Client) {
	cli.Opcode = CTLADD
	ctlCh <- cli
	reader := bufio.NewReader(cli.Conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		msgCh <- NewClienetMsg(cli.Id, line)
	}
	cli.Opcode = CTLDEL
	ctlCh <- cli
}

func HandleMsg(ctlCh chan Client, msgCh chan *ClientMsg) {
	clientMap := make(map[int]*bufio.Writer)
	for {
		select {
		case cli := <-ctlCh:
			if cli.Opcode == CTLADD {
				clientMap[cli.Id] = bufio.NewWriter(cli.Conn)
			} else if cli.Opcode == CTLDEL {
				delete(clientMap, cli.Id)
			}
			break
		case msg := <-msgCh:
			if msg.DstClient == -1 {
				for _, val := range clientMap {
					val.WriteString(msg.Msg)
					val.Flush()
				}
			} else if msg.DstClient > -1 {
				if val, ok := clientMap[msg.DstClient]; ok {
					val.WriteString(msg.Msg)
					val.Flush()
				}
			}
			break
		}
	}
}

func main() {
	clientId := 1
	server, _ := net.Listen("tcp", ":2323")
	ctlChan := make(chan Client)
	msgChan := make(chan *ClientMsg)
	go HandleMsg(ctlChan, msgChan)
	for {
		conn, err := server.Accept()
		if err == nil {
			cli := Client{
				Opcode: 0,
				Id:     clientId,
				Conn:   conn,
			}
			clientId++
			go HandleConn(ctlChan, msgChan, cli)
		}
	}
}
