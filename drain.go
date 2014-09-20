package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/shawnsi/drain/conntrack"
	"github.com/shawnsi/drain/iptables"
	"log"
	"os"
	"os/user"
	"time"
)

func Monitor(ports []string) {
	for connections := conntrack.Established(ports); len(connections) > 0; {
		fmt.Printf("%d connections remaining on ports %s...\n", len(connections), ports)
		time.Sleep(1 * time.Second)
		connections = conntrack.Established(ports)
	}
}

func Start(ports []string, excludes []string) error {
	for _, port := range ports {
		chain := fmt.Sprintf("DRAIN_%s", port)

		// Create DRAIN chain
		if out, err := iptables.Command("-N", chain).CombinedOutput(); err != nil {
			return fmt.Errorf("Failed to add chain for port %s!\n%s", port, out)
		}

		// Append RETURN for each excluded hostname
		for _, exclude := range excludes {
			if out, err := iptables.Command("-A", chain, "-s", exclude, "-j", "RETURN").CombinedOutput(); err != nil {
				return fmt.Errorf("Failed to add exclude for host %s!\n%s", exclude, out)
			}
		}

		// Append REJECT for new TCP connections on the port
		if out, err := iptables.Command(
			"-A", chain,
			"-m", "state", "--state", "NEW",
			"-j", "REJECT",
			"-p", "tcp",
			"--dport", port).CombinedOutput(); err != nil {
			return fmt.Errorf("Failed to add TCP REJECT for port %s!\n%s", port, out)
		}

		// Jump to DRAIN chain in INPUT chain
		if out, err := iptables.Command("-A", "INPUT", "-j", chain).CombinedOutput(); err != nil {
			return fmt.Errorf("Failed to add jump to INPUT chain for port %s!\n%s", port, out)
		}
	}

	return nil
}

func Status(ports []string) {
	drains, _ := iptables.Drains(iptables.IptablesFetcher{})
	for port, exclusions := range drains {
		fmt.Printf("%s => %s\n", port, exclusions)
	}
}

func Stop(ports []string) error {
	for _, port := range ports {
		chain := fmt.Sprintf("DRAIN_%s", port)

		// Delete jump to DRAIN chain in INPUT chain
		if out, err := iptables.Command("-D", "INPUT", "-j", chain).CombinedOutput(); err != nil {
			return fmt.Errorf("Failed to delete jump from INPUT chain for port %s!\n%s", port, out)
		}

		// Flush the DRAIN chain
		if out, err := iptables.Command("-F", chain).CombinedOutput(); err != nil {
			return fmt.Errorf("Failed to flush chain for port %s!\n%s", port, out)
		}

		// Delete the DRAIN chain
		if out, err := iptables.Command("-X", chain).CombinedOutput(); err != nil {
			return fmt.Errorf("Failed to delete chain for port %s!\n%s", port, out)
		}
	}

	return nil
}

func main() {
	usage := `TCP Drain.

Usage:
  drain [options] monitor <port>
  drain [options] start [--exclude=<host>...] <port>...
  drain [options] stop <port>...
  drain [options] status

Options:
  -e --exclude=<host>  Exclude a hostname or ip from the drain
  -h --help            Show this screen
  -d --debug           Print debug information
  -v --version	       show version

Commands:
  monitor     Monitor connection counts
  start       Stop new TCP connections and drain existing
  stop        Open all TCP connections
  status      Show active drains`

	if user, err := user.Current(); err != nil {
		log.Fatal(err)
	} else if user.Uid != "0" {
		log.Fatal("You must be root to run drain")
	}

	arguments, _ := docopt.Parse(usage, nil, true, "Drain 0.0.1", false)

	if arguments["--debug"].(bool) {
		os.Setenv("DEBUG", "1")
	}

	if arguments["monitor"].(bool) {
		ports := arguments["<port>"].([]string)
		Monitor(ports)
	}

	if arguments["start"].(bool) {
		ports := arguments["<port>"].([]string)
		excludes := arguments["--exclude"].([]string)

		if err := Start(ports, excludes); err != nil {
			fmt.Print(err)
		} else {
			Monitor(ports)
		}
	}

	if arguments["status"].(bool) {
		ports := arguments["<port>"].([]string)
		Status(ports)
	}

	if arguments["stop"].(bool) {
		ports := arguments["<port>"].([]string)

		if err := Stop(ports); err != nil {
			fmt.Print(err)
		}
	}
}
