#!/usr/bin/env python

import re
import time
from subprocess import check_call, check_output, CalledProcessError, Popen

import psutil

iptables_comment = 'DRAIN'


def iptables_rule(action, port):
    rule = '%s INPUT -m state --state NEW -j REJECT -p tcp --dport %s \
            -m comment --comment %s' % (action, port, iptables_comment)
    check_call('iptables %s' % rule, shell=True)


def iptables_running():
    p = Popen(['lsmod | grep -q ^ip_tables'], shell=True).wait()
    return (p == 0)


def established(port):
    return filter(
        lambda c: c.laddr[1] == int(port) and c.status != 'LISTEN',
        psutil.net_connections(),
    )

def monitor(port):
    draining = True

    while draining:
        draining = len(established(port))
        yield draining
        time.sleep(1)

def running():
    try:
        rules = check_output('iptables -L INPUT -n | tail -n +3 | grep -F "/* DRAIN */"', shell=True)
        return re.findall('dpt:(\d+)', rules)
    except CalledProcessError:
        return []


def start(port):
    iptables_rule('-A', port)


def stop(port):
    iptables_rule('-D', port)

# vim: tabstop=4 expandtab shiftwidth=4 softtabstop=4
