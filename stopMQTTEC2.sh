#!/bin/bash
ps axf | grep mosquitto | grep -v grep | awk '{print "sudo kill -9 " $1}' | sh
exit