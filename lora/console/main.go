package main

import (
	"machine"
	"time"
)

// change these to test a different UART or pins if available
var (
	uart = machine.Serial
	tx   = machine.UART_TX_PIN
	rx   = machine.UART_RX_PIN
)

var (
	uart1 = machine.UART1
	tx1   = machine.UART1_TX_PIN
	rx1   = machine.UART1_RX_PIN
)

func handle(input []byte) {
	uart1.Write(input)
	uart1.Write([]byte("\r\n"))
	for i := 0; i < 200; i++ {
		for uart1.Buffered() > 0 {
			data, _ := uart1.ReadByte()
			uart.WriteByte(data)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

var buf [242]byte

func exec(cmd, expect string, wait int) {
	uart.Write([]byte(cmd))
	uart.Write([]byte("\r\n"))
	uart1.Write([]byte(cmd))
	uart1.Write([]byte("\r\n"))
	i := 0
	for j := 0; j < wait/100; j++ {
		for uart1.Buffered() > 0 {
			buf[i], _ = uart1.ReadByte()
			uart.WriteByte(buf[i])
			i++
		}
		time.Sleep(100 * time.Millisecond)
	}
	resp := string(buf[:i-2])
	if resp != expect {
		panic("bummer")
	}
}

func main() {
	time.Sleep(2 * time.Second)

	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	uart1.Configure(machine.UARTConfig{TX: tx1, RX: rx1, BaudRate: 9600})

	exec("AT+FDEFAULT=Seeed", "+FDEFAULT: OK", 1000)
	exec("AT+LOG=DEBUG", "+LOG: DEBUG", 1000)
	exec("AT+MODE=TEST", "+MODE: TEST", 1000)
	exec("AT+TEST=RFCFG,902.3,SF7,125,12,15,14,ON,OFF,OFF",
		"+TEST: RFCFG F:902300000, SF7, BW125K, TXPR:12, RXPR:15, POW:14dBm, CRC:ON, IQ:OFF, NET:OFF",
		1000)

	uart.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))

	input := make([]byte, 64)
	i := 0
	for {
		if uart.Buffered() > 0 {
			data, _ := uart.ReadByte()

			switch data {
			case 13:
				// return key
				uart.Write([]byte("\r\n"))
				//				uart.Write([]byte("You typed: "))
				//				uart.Write(input[:i])
				//				uart.Write([]byte("\r\n"))
				handle(input[:i])
				i = 0
			default:
				// just echo the character
				uart.WriteByte(data)
				input[i] = data
				i++
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
