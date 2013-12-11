#!/usr/bin/env python

import sys
import os
import fcntl
import time
import subprocess

if len(sys.argv) < 2:
    print >> sys.stderr, "Usage: %s <program>" % sys.argv[0]
    sys.exit(1)

prog_name = os.path.basename(sys.argv[1])
prog_dir = os.path.dirname(sys.argv[1])
prog_argv = sys.argv[1:]

RUN_BASE_DIR = '/var/run'

lockfile = "%s/run-%s.lock" % (RUN_BASE_DIR, prog_name)
lockfd = os.open(lockfile, os.O_WRONLY | os.O_CREAT, 0666)
fcntl.flock(lockfd, fcntl.LOCK_EX | fcntl.LOCK_NB) # throw exception if locked already

logfile = "%s/run-%s.log" % (RUN_BASE_DIR, prog_name)
logfp = open(logfile, "a")

# setup timezone
os.environ["TZ"] = "Asia/Shanghai"
time.tzset()

def now():
    return time.strftime("%Y-%m-%d %T", time.localtime(time.time()))

child_pid = os.fork()
if child_pid != 0:
    # i am parent
    sys.exit(0)

while True:
    print >> logfp, "[%d] %s - Try running" % (os.getpid(), now()), " ".join(prog_argv)
    logfp.flush()

    sp = subprocess.Popen(prog_argv, cwd=prog_dir, close_fds=True, stdout=logfp, stderr=subprocess.STDOUT)
    sp.wait()

    print >> logfp, "[%d] %s - Terminated" % (os.getpid(), now()), " ".join(prog_argv)
    logfp.flush()

    time.sleep(5)
