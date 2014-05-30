#!/usr/bin/env python

import re
from subprocess import check_output

def all():
    ifaces = check_output('ip -o addr', shell=True)
    return re.findall('inet ([0-9\.]+)\/\d+', ifaces)

# vim: tabstop=4 expandtab shiftwidth=4 softtabstop=4

