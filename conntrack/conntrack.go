package conntrack

import (
	"bufio"
	"fmt"
	"github.com/shawnsi/drain/iptables"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

type Connection struct {
	Proto           string
	Source          net.IP
	Destination     net.IP
	SourcePort      string
	DestinationPort string
	State           string
	Raw             string
}

type connectionFetcherInterface interface {
	GetConntrackLines(path string) []string
}

type connectionFetcher struct{}

func (fetcher connectionFetcher) GetConntrackLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lines := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines
}

func NewConnection(raw string, module string) (connection Connection, err error) {
	err = nil
	fields := strings.Fields(raw)

	if module == "nf_conntrack" {
		fields = fields[2:]
	}

	proto := fields[0]

	if proto != "tcp" {
		err = fmt.Errorf("Not supported protocol: %s", proto)
		return
	}

	source_s := strings.Split(fields[4], "=")[1]
	destination_s := strings.Split(fields[5], "=")[1]
	source := net.ParseIP(source_s)
	destination := net.ParseIP(destination_s)
	sport := strings.Split(fields[6], "=")[1]
	dport := strings.Split(fields[7], "=")[1]
	state := fields[3]

	connection = Connection{
		Proto:           proto,
		Source:          source,
		Destination:     destination,
		SourcePort:      sport,
		DestinationPort: dport,
		State:           state,
		Raw:             raw,
	}

	return
}

func Connections(fetcher connectionFetcherInterface) (connections []Connection) {
	connections = make([]Connection, 0)

	module, path := GetModule(moduleFetcher{})
	for _, line := range fetcher.GetConntrackLines(path) {
		if connection, err := NewConnection(line, module); err == nil {
			connections = append(connections, connection)

			if os.Getenv("DEBUG") == "1" {
				log.Printf("Found connection: %+v\n", connection)
			}
		}
	}
	return
}

func Established(ports []string) []Connection {
	connections := make([]Connection, 0)
	drains, _ := iptables.Drains(iptables.IptablesFetcher{})

	for _, connection := range LocalConnections() {
		for _, port := range ports {
			if connection.DestinationPort == port {
				if connection.State == "ESTABLISHED" || connection.State == "TIME_WAIT" {
					excludes := drains[port]
					excluded := false

					if len(excludes) > 0 {
						for _, IP := range excludes {
							if IP.Equal(connection.Source) {
								excluded = true
							}
						}
					}

					if excluded == false {
						connections = append(connections, connection)

						if os.Getenv("DEBUG") == "1" {
							log.Printf("Found established connection: %+v\n", connection)
						}
					}
				}
			}
		}
	}

	return connections
}

func LocalConnections() []Connection {
	connections := make([]Connection, 0)
	for _, connection := range Connections(connectionFetcher{}) {
		if LocalIP(netFetcher{}, connection.Destination) {
			connections = append(connections, connection)

			if os.Getenv("DEBUG") == "2" {
				log.Printf("Found local connection: %+v\n", connection)
			}
		}
	}
	return connections
}

type netFetcherInterface interface {
	InterfaceIPs() ([]net.IP, error)
}

type netFetcher struct{}

func (fetcher netFetcher) InterfaceIPs() ([]net.IP, error) {
	ips := make([]net.IP, 0)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, addr := range addrs {
		ip, _, _ := net.ParseCIDR(addr.String())
		ips = append(ips, ip)
	}

	return ips, nil
}

func LocalIP(fetcher netFetcherInterface, ip net.IP) bool {
	iface_ips, err := fetcher.InterfaceIPs()
	if err != nil {
		log.Fatal(err)
	}

	for _, iface_ip := range iface_ips {
		if iface_ip.Equal(ip) {
			return true
		}
	}

	return false
}

type moduleFetcherInterface interface {
	Exists(name string) bool
	Loaded(name string) bool
}

type moduleFetcher struct{}

func (moduleFetcher) Exists(name string) bool {
	if err := exec.Command("/sbin/modinfo", name).Run(); err != nil {
		return false
	}
	return true
}

func (moduleFetcher) Loaded(name string) bool {
	loaded := false
	out, _ := exec.Command("/sbin/lsmod").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, name) {
			loaded = true
		}
	}
	return loaded
}

func GetModule(fetcher moduleFetcherInterface) (name string, path string) {
	if fetcher.Exists("ip_conntrack") {
		name, path = "ip_conntrack", "/proc/net/ip_conntrack"
	}

	if fetcher.Exists("nf_conntrack") {
		name, path = "nf_conntrack", "/proc/net/nf_conntrack"
	}

	if name == "" {
		log.Fatal("Cannot find any conntrack module.")
	}

	if !fetcher.Loaded(name) {
		log.Fatalf("%s module not loaded", name)
	}

	return
}
