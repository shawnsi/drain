#!/usr/bin/env python

"""TCP Drain.

Usage:
  drain [options] monitor <port>
  drain [options] start <port>...
  drain [options] stop <port>...
  drain [options] status

Options:
  -h --help     Show this screen
  -d --debug    Print debug information

Commands:
  start       Stop new TCP connections and drain existing
  stop        Open all TCP connections
  status      Show active drains

"""
from __future__ import print_function
import sys

from docopt import docopt
import drain

CLEAR_LINE = '\r\033[K'

def monitor(port):
    for count in drain.monitor(port):
        sys.stdout.write(CLEAR_LINE + '%d connections remaining...' % count)
        sys.stdout.flush()

    print()

if __name__ == '__main__':
    args = docopt(__doc__, version='0.0.1')

    if args['monitor']:
        monitor(args['<port>'][0])

    if args['start']:
        for port in args['<port>']:
            print('Draining port %s' % port)
            drain.start(port)
            monitor(port)


    if args['stop']:
        for port in args['<port>']:
            drain.stop(port)

    if args['status']:
        if not drain.iptables_running():
            print('iptables is not running!')

        for port in drain.running():
            print(port)

# vim: tabstop=8 expandtab shiftwidth=4 softtabstop=4
