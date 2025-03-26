package peers

import (
	"elevator-project/config"
	"elevator-project/network/conn"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"
)

type PeerUpdate struct {
	Peers []int
	New   int
	Lost  []int
}

func Transmitter(port int, id int, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(config.PeerBcastInterval):
		}
		if enable {
			idStr := strconv.Itoa(id)
			conn.WriteTo([]byte(idStr), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte

	lastSeen := make(map[int]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false
		p := PeerUpdate{
			New:  -1,
			Lost: make([]int, 0),
		}

		conn.SetReadDeadline(time.Now().Add(config.PeerBcastInterval))
		n, _, _ := conn.ReadFrom(buf[0:])

		idStr := string(buf[:n])
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		// Adding new connection
		if _, idExists := lastSeen[id]; !idExists {
			p.New = id
			updated = true
		}
		lastSeen[id] = time.Now()

		// Removing dead connections
		for k, v := range lastSeen {
			if time.Since(v) > config.DisconnectTime {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]int, 0, len(lastSeen))

			for k := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Ints(p.Peers)
			sort.Ints(p.Lost)
			peerUpdateCh <- p
		}
	}
}
