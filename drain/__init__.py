#!/usr/bin/env python

import re
import socket
import time
from subprocess import check_call, check_output, CalledProcessError, Popen

import psutil

CHAIN_PREFIX = 'DRAIN'


def iptables(rule):
    return check_call('iptables %s' % rule, shell=True)


def iptables_chain(port):
    return '%s_%s' % (CHAIN_PREFIX, port)


def iptables_running():
    p = Popen(['lsmod | grep -q ^ip_tables'], shell=True).wait()
    return (p == 0)


def established(port):
    return filter(
        lambda c: c.laddr[1] == int(port) and c.status != 'LISTEN',
        psutil.net_connections(),
    )


def monitor(port, excludes=None):
    draining = True

    excludes = map(
        socket.gethostbyname,
        excludes
    )

    while draining:
        connections = filter(
            lambda c: c.raddr[0] not in excludes,
            established(port)
        )
        draining = len(connections)
        yield draining
        time.sleep(1)


def running():
    try:
        rules = check_output('iptables -L | grep -E "Chain DRAIN_[0-9]+"', shell=True)
        return re.findall('DRAIN_(\d+)', rules)
    except CalledProcessError:
        return []


def start(port, excludes=None):
    # Create the chain
    chain = iptables_chain(port)
    iptables('-N %s' % chain)

    # Add source exclusions via RETURN action
    for exclude in excludes or []:
        iptables('-A %s -s %s -j RETURN' % (chain, exclude))

    # Add TCP drain
    rule = '-A %s -m state --state NEW -j REJECT -p tcp --dport %s' % (chain, port)
    iptables(rule)

    # Add jump to INPUT chain
    iptables('-A INPUT -j %s' % chain)


def stop(port):
    # Drop the chain
    chain = iptables_chain(port)
    iptables('-D INPUT -j %s' % chain)
    iptables('-F %s' % chain)
    iptables('-X %s' % chain)

# vim: tabstop=4 expandtab shiftwidth=4 softtabstop=4
