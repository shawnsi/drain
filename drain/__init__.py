#!/usr/bin/env python

import re
from subprocess import check_call, check_output, CalledProcessError, Popen

iptables_comment = 'DRAIN'


def iptables_rule(action, port):
    rule = '%s INPUT -m state --state NEW -j REJECT -p tcp --dport %s \
            -m comment --comment %s' % (action, port, iptables_comment)
    check_call('iptables %s' % rule, shell=True)


def iptables_running():
    p = Popen(['lsmod | grep -q ^ip_tables'], shell=True).wait()
    return (p == 0)


def monitor(port):
    try:
        p = Popen('watch -n1 "netstat -anp | grep ESTABLISHED | grep :%s"' % port, shell=True).wait()
    except KeyboardInterrupt:
        pass


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
