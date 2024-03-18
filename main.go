package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Command line flags.
var (
	listenAddr string
	bufSize    uint
)

// A single client's benchmarking data.
type bench struct {
	bytesPerSec int
	remoteAddr  string
	connected   bool
}

// Bubbletea model.
type model struct {
	benches []bench
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc": // Quit.
			return m, tea.Quit
		case "h": // Do not display disconnected clients.
			m.benches = delDeadBenches(m.benches)
			return m, nil
		}
	case bench: // Add a new bench, or update if already exists.
		found := -1
		for i, v := range m.benches {
			if msg.remoteAddr == v.remoteAddr {
				found = i // Already exits, save its index.
			}
		}
		if found >= 0 { // Exists, replace.
			m.benches[found] = msg
		} else { // Such a client not found, append.
			m.benches = append(m.benches, msg)
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	var status strings.Builder
	// Status bar.
	status.WriteString("`q` - quit, `h` - hide disconnected.\n")
	status.WriteString(fmt.Sprintf("[ Listening on %s | Chunk size %d ]\n",
		listenAddr, bufSize))
	// Clients...
	for _, v := range m.benches {
		if !v.connected {
			status.WriteString("X ")
		}
		rateMiBPerSec := float64(v.bytesPerSec) / 1024.0 / 1024.0
		status.WriteString(fmt.Sprintf("->%s %d_B/s %.2f_MiB/s\n",
			v.remoteAddr, v.bytesPerSec, rateMiBPerSec))
	}
	// Help.

	return status.String()
}

// Go over a slice of clients (benchmarks) and delete the disconnected ones.
func delDeadBenches(mixed []bench) []bench {
	var active []bench
	for _, v := range mixed {
		if v.connected {
			active = append(active, v)
		}
	}
	return active
}

func handleIncoming(ln net.Listener, tui *tea.Program) {
	for {
		conn, err := ln.Accept() // Wait for a new incoming connection.
		if err != nil {
			panic(err)
		}
		ch := make(chan int, 10)
		go connFlood(conn, ch)                    // Send data into it.
		go floodBench(ch, tui, conn.RemoteAddr()) // Benchmark it.
	}
}

// Continuously send data to the incoming connection, and report the amount
// sent to the benchmarking function.
func connFlood(conn net.Conn, bytesSent chan int) {
	data := make([]byte, bufSize)
	for i := range data {
		data[i] = 'A'
	}
	for {
		n, err := conn.Write(data)
		if err != nil {
			close(bytesSent)
			break
		}
		bytesSent <- n
	}
}

// Calculate data flow rate and report it to the user interface.
// Data flow rate is the ratio: (total bytes sent)/(total time taken).
func floodBench(bytesSent chan int, tui *tea.Program, remoteAddr net.Addr) {
	startTime := time.Now()
	totalBytesSent := 0
	dataFlowRate := 0.0
	for n := range bytesSent {
		totalBytesSent += n
		elapsedTime := time.Since(startTime).Seconds()
		dataFlowRate = float64(totalBytesSent) / elapsedTime
		tui.Send(bench{
			bytesPerSec: int(dataFlowRate),
			remoteAddr:  remoteAddr.String(),
			connected:   true,
		})
	}
	tui.Send(bench{
		bytesPerSec: int(dataFlowRate),
		remoteAddr:  remoteAddr.String(),
		connected:   false,
	})
}

func init() {
	flag.StringVar(&listenAddr, "listen", ":8080", "listen address")
	flag.UintVar(&bufSize, "buf", 1024, "message chunk size")
}

func main() {
	flag.Parse()

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	tui := tea.NewProgram(model{
		benches: []bench{},
	})

	go handleIncoming(ln, tui)           // Listen incoming requests.
	if _, err := tui.Run(); err != nil { // Start user interface (TUI).
		panic(err)
	}
}
