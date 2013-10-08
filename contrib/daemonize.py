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

lockfilename = "/var/run/run-%s.lock" % prog_name
lockfd = os.open(lockfilename, os.O_WRONLY | os.O_CREAT, 0666)
fcntl.flock(lockfd, fcntl.LOCK_EX | fcntl.LOCK_NB)

logfilename = "/var/run/run-%s.log" % prog_name
logfp = open(logfilename, "a")

child_pid = os.fork()
if child_pid != 0:
	sys.exit(0)

prog_argv = sys.argv[1:]

while True:
	print >> logfp, "Try running", " ".join(prog_argv)
	sp = subprocess.Popen(prog_argv, cwd=prog_dir, close_fds=True, stdout=logfp, stderr=subprocess.STDOUT)
	sp.wait()
	time.sleep(3)

