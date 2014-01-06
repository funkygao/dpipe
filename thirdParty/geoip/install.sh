#!/bin/sh

# install geoip c api
git clone https://github.com/maxmind/geoip-api-c.git
cd geoip-api-c/
./bootstrap
./configure
make
make install

# install geoip go lib
go get github.com/abh/geoip
