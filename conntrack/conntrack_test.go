package conntrack

import (
	"net"
	"strings"
	"testing"
)

func BenchmarkLocalIP(b *testing.B) {
	ip := net.ParseIP("127.0.0.1")
	for i := 0; i < b.N; i++ {
		LocalIP(mockNetFetcher{}, ip)
	}
}

type mockNetFetcher struct{}

func (fetcher mockNetFetcher) InterfaceIPs() ([]net.IP, error) {
	ips := make([]net.IP, 0)
	ips = append(ips, net.ParseIP("127.0.0.1"))
	ips = append(ips, net.ParseIP("172.16.24.1"))
	return ips, nil
}

func TestLocalIP(t *testing.T) {
	fetcher := mockNetFetcher{}
	ip := net.ParseIP("127.0.0.1")
	if !LocalIP(fetcher, ip) {
		t.Errorf("LocalIP(\"%s\") => false, expected => true", ip)
	}

	ip = net.ParseIP("172.16.24.1")
	if !LocalIP(fetcher, ip) {
		t.Errorf("LocalIP(\"%s\") => false, expected => true", ip)
	}

	ip = net.ParseIP("8.8.8.8")
	if LocalIP(fetcher, ip) {
		t.Errorf("LocalIP(\"%s\") => true, expected => false", ip)
	}
}

type mockNfModuleFetcher struct{}

func (mockNfModuleFetcher) Exists(name string) bool {
	result := false

	if name == "nf_conntrack" {
		result = true
	}

	return result
}

func (mockNfModuleFetcher) Loaded(name string) bool {
	return true
}

type mockIpModuleFetcher struct{}

func (mockIpModuleFetcher) Exists(name string) bool {
	result := false

	if name == "ip_conntrack" {
		result = true
	}

	return result
}

func (mockIpModuleFetcher) Loaded(name string) bool {
	return true
}

func TestGetModule(t *testing.T) {
	name, path := GetModule(mockNfModuleFetcher{})
	if name != "nf_conntrack" || path != "/proc/net/nf_conntrack" {
		t.Errorf("GetModule() => (\"%s\", \"%s\"), expected => (\"nf_conntrack\", \"/proc/net/nf_conntrack\")", name, path)
	}

	name, path = GetModule(mockIpModuleFetcher{})
	if name != "ip_conntrack" || path != "/proc/net/ip_conntrack" {
		t.Errorf("GetModule() => (\"%s\", \"%s\"), expected => (\"ip_conntrack\", \"/proc/net/ip_conntrack\")", name, path)
	}
}

type mockConnectionFetcher struct{}

func (fetcher mockConnectionFetcher) GetConntrackLines(path string) []string {
	nf_conntrack := `
ipv4     2 udp      17 4 src=1.2.3.4 dst=5.6.7.8 sport=1000 dport=2000 [UNREPLIED] src=5.6.7.8 dst=1.2.3.4 sport=2000 dport=1000 mark=0 zone=0 use=2
ipv4     2 tcp      6 431973 ESTABLISHED src=1.2.3.4 dst=5.6.7.8 sport=3000 dport=4000 src=5.6.7.8 dst=1.2.3.4 sport=4000 dport=3000 [ASSURED] mark=0 zone=0 use=2
ipv4     2 udp      17 4 src=1.2.3.4 dst=5.6.7.8 sport=5000 dport=6000 [UNREPLIED] src=5.6.7.8 dst=1.2.3.4 sport=6000 dport=5000 mark=0 zone=0 use=2
ipv4     2 tcp      6 431973 ESTABLISHED src=10.20.30.40 dst=50.60.70.80 sport=7000 dport=8000 src=50.60.70.80 dst=10.20.30.40 sport=8000 dport=7000 [ASSURED] mark=0 zone=0 use=2
ipv4     2 tcp      6 431973 TIME_WAIT src=10.20.30.40 dst=50.60.70.80 sport=9000 dport=10000 src=50.60.70.80 dst=10.20.30.40 sport=10000 dport=9000 [ASSURED] mark=0 zone=0 use=2
	`
	return strings.Split(strings.TrimSpace(nf_conntrack), "\n")
}

func TestConnections(t *testing.T) {
	connections := Connections(mockConnectionFetcher{})
	count := len(connections)

	if count != 3 {
		t.Errorf("len(Connections()) => %d), expected => 3", count)
	}

	connection := connections[0]
	if !connection.Source.Equal(net.ParseIP("1.2.3.4")) {
		t.Errorf("connection.Source => %s, expected => 1.2.3.4)", connection.Source)
	}
	if connection.SourcePort != "3000" {
		t.Errorf("connection.SourcePort => \"%s\", expected => \"3000\")", connection.SourcePort)
	}
	if !connection.Destination.Equal(net.ParseIP("5.6.7.8")) {
		t.Errorf("connection.Destination => %s, expected => 5.6.7.8)", connection.Destination)
	}
	if connection.DestinationPort != "4000" {
		t.Errorf("connection.DestinationPort => \"%s\", expected => \"4000\")", connection.DestinationPort)
	}
	if connection.Proto != "tcp" {
		t.Errorf("connection.Proto => \"%s\", expected => \"tcp\")", connection.Proto)
	}
	if connection.State != "ESTABLISHED" {
		t.Errorf("connection.State => \"%s\", expected => \"ESTABLISHED\")", connection.State)
	}
}
