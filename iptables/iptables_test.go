package iptables

import (
	"net"
	"reflect"
	"strings"
	"testing"
)

type mockIptablesFetcher struct{}

func (fetcher mockIptablesFetcher) List() []string {
	out := `Chain INPUT (policy ACCEPT)
target     prot opt source               destination         
DRAIN_8888  all  --  anywhere             anywhere            
DRAIN_9999  all  --  anywhere             anywhere            

Chain FORWARD (policy ACCEPT)
target     prot opt source               destination         
ACCEPT     all  --  anywhere             anywhere             ctstate RELATED,ESTABLISHED
ACCEPT     all  --  anywhere             anywhere            
ACCEPT     all  --  anywhere             anywhere            

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination         

Chain DRAIN_8888 (1 references)
target     prot opt source               destination         
REJECT     tcp  --  anywhere             anywhere             state NEW tcp dpt:distinct reject-with icmp-port-unreachable

Chain DRAIN_9999 (1 references)
target     prot opt source               destination         
RETURN     all  --  localhost.localdomain  anywhere            
RETURN     all  --  google-public-dns-a.google.com  anywhere            
REJECT     tcp  --  anywhere             anywhere             state NEW tcp dpt:distinct reject-with icmp-port-unreachable`

	return strings.Split(out, "\n")
}

func TestDrains(t *testing.T) {
	expected := make(map[string][]net.IP)
	expected["8888"] = make([]net.IP, 0)

	excludes := make([]net.IP, 0)
	hosts := []string{"localhost", "google-public-dns-a.google.com"}
	for _, host := range hosts {
		IPs, _ := net.LookupIP(host)
		excludes = append(excludes, IPs...)
	}

	expected["9999"] = excludes

	drains, _ := Drains(mockIptablesFetcher{})
	count := len(drains)

	if count != 2 {
		t.Errorf("len(Drains()) => %d, expected => 2", count)
	}

	if !reflect.DeepEqual(drains, expected) {
		t.Errorf("Drains() => %q, expected => %q", drains, expected)
	}
}
