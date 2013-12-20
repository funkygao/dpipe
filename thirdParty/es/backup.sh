#!/bin/sh
#===============================
# daily backup ES index
#
# easy_install esclient
#
# to restore:
# esimport --url http://localhost:9200 --file kibana-int.bz2
#
#===============================

date=`date +%F`
esdump --url http://localhost:9200/ --indexes kibana-int --bzip2 --file kibana-int-$date.bz2
