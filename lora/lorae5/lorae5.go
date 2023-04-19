package lorae5

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"machine"
	"time"
)

type LoraE5 struct {
	uart *machine.UART
	buf  [1024]byte
}

func New(uart *machine.UART, tx, rx machine.Pin, baudrate uint32) *LoraE5 {
	l := LoraE5{uart: uart}
	l.uart.Configure(machine.UARTConfig{TX: tx, RX: rx, BaudRate: baudrate})
	return &l
}

func (l *LoraE5) response(wait int) []byte {
	i := 0
	// TODO: use bufio buffer to WriteByte to
	for j := 0; j < wait/100; j++ {
		for l.uart.Buffered() > 0 {
			b, _ := l.uart.ReadByte()
			if i < len(l.buf) {
				l.buf[i] = b
				//print(string(l.buf[i]))
				i++
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return l.buf[:i]
}

func (l *LoraE5) exec(cmd, expect []byte, wait int) error {
	//println(string(cmd))
	l.uart.Write(cmd)
	l.uart.Write([]byte("\r\n"))
	resp := l.response(wait)
	//println(string(resp))
	reader := bytes.NewReader(resp)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		//println("SCAN", scanner.Text())
		if bytes.HasPrefix(scanner.Bytes(), expect) {
			return nil
		}
	}
	return errors.New("Expected " + string(expect))
}

func (l *LoraE5) Tx(msg []byte, wait int) error {
	var cmd []byte

	msgHex := make([]byte, hex.EncodedLen(len(msg)))
	hex.Encode(msgHex, msg)

	cmd = append(cmd, []byte("AT+TEST=TXLRPKT,\"")...)
	cmd = append(cmd, msgHex...)
	cmd = append(cmd, []byte("\"")...)

	return l.exec(cmd, []byte("+TEST: TX DONE"), wait)
}

func (l *LoraE5) Rx(wait int) ([]byte, error) {
	var cmd = []byte("AT+TEST=RXLRPKT\r\n")
	var expect = []byte("+TEST: RX ")
	//println(string(cmd))
	l.uart.Write(cmd)
	resp := l.response(wait)
	//println(string(resp))
	reader := bytes.NewReader(resp)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		//println("SCAN", scanner.Text())
		scan := scanner.Bytes()
		if bytes.HasPrefix(scan, expect) {
			pktHex := scan[len(expect)+1 : len(scan)-1]
			pkt := make([]byte, hex.DecodedLen(len(pktHex)))
			hex.Decode(pkt, pktHex)
			//println(string(pkt))
			return pkt, nil
		}
	}
	return nil, errors.New("No Rx packet")
}

func (l *LoraE5) RxPoll(out chan []byte, wait int) {
	for {
		pkt, err := l.Rx(wait)
		if err == nil {
			out <- pkt
		}
	}
}

type command struct {
	cmd    []byte
	expect []byte
	wait   int
}

var cmds = map[string]command{
	"reset": {
		cmd:    []byte("AT+FDEFAULT=Seeed"),
		expect: []byte("+FDEFAULT: OK"),
		wait:   1000,
	},
	"debug": {
		cmd:    []byte("AT+LOG=DEBUG"),
		expect: []byte("+LOG: DEBUG"),
		wait:   1000,
	},
	"test": {
		cmd:    []byte("AT+MODE=TEST"),
		expect: []byte("+MODE: TEST"),
		wait:   1000,
	},
	"rfcfg": {
		cmd:    []byte("AT+TEST=RFCFG,902.3,SF7,125,12,15,14,ON,OFF,OFF"),
		expect: []byte("+TEST: RFCFG F:902300000, SF7, BW125K, TXPR:12, RXPR:15, POW:14dBm, CRC:ON, IQ:OFF, NET:OFF"),
		wait:   1000,
	},
}

func (l *LoraE5) Init() error {
	for _, cmd := range []string{"reset" /* "debug",*/, "test", "rfcfg"} {
		command := cmds[cmd]
		err := l.exec(command.cmd, command.expect, command.wait)
		if err != nil {
			return err
		}
	}
	return nil
}
