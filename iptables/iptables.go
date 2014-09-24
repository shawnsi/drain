package iptables

import (
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type iptablesFetcherInterface interface {
	List() []string
}

type IptablesFetcher struct{}

func (fetcher IptablesFetcher) List() []string {
	lines := make([]string, 0)

	out, err := Command("-L").Output()
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		lines = append(lines, string(line))
	}

	return lines
}

func Command(arg ...string) *exec.Cmd {
	if os.Getenv("DEBUG") == "1" {
		command := []string{"iptables"}
		command = append(command, arg...)
		log.Print(strings.Join(command, " "))
	}

	return exec.Command("/sbin/iptables", arg...)
}

func Drains(fetcher iptablesFetcherInterface) (map[string][]net.IP, error) {
	drains := make(map[string][]net.IP)
	in_drain := false
	port := ""

	for _, line := range fetcher.List() {
		if strings.HasPrefix(line, "Chain") {
			re := regexp.MustCompile("DRAIN_([0-9]+)")
			if match := re.FindStringSubmatch(line); match != nil {
				in_drain = true
				port = match[1]
				drains[port] = make([]net.IP, 0)
			} else {
				in_drain = false
			}
		}

		if in_drain && strings.HasPrefix(line, "RETURN") {
			fields := strings.Fields(line)
			IPs, err := net.LookupIP(fields[3])
			if err != nil {
				return drains, err
			}
			drains[port] = append(drains[port], IPs...)
		}
	}

	return drains, nil
}
