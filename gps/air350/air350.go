package air350

import (
	"machine"
	"time"
)

type Air350 struct {
	uart *machine.UART
	buf  [128]byte
}

func New(uart *machine.UART, tx, rx machine.Pin, baudrate uint32) *Air350 {
	a := Air350{uart: uart}
	a.uart.Configure(machine.UARTConfig{TX: tx, RX: rx, BaudRate: baudrate})
	return &a
}

func (a *Air350) Run(out chan string) {
	i := 0
	for {
		for a.uart.Buffered() > 0 {
			b, _ := a.uart.ReadByte()
			switch b {
			case '\r':
				out <- string(a.buf[:i])
				i = 0
			case '\n':
			default:
				a.buf[i] = b
				i++
				if i == len(a.buf) {
					i = 0
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
