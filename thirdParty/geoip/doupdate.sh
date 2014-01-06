#!/bin/sh
#==================================================================
# cp {this file} to /home/funkygao/gopkg/src/github.com/abh/geoip/db
#==================================================================
cd /home/funkygao/gopkg/src/github.com/abh/geoip/db

wget -N http://geolite.maxmind.com/download/geoip/database/GeoLiteCountry/GeoIP.dat.gz
wget -N http://geolite.maxmind.com/download/geoip/database/GeoIPv6.dat.gz
wget -N http://geolite.maxmind.com/download/geoip/database/GeoLiteCity.dat.gz
wget -N http://geolite.maxmind.com/download/geoip/database/GeoLiteCityv6-beta/GeoLiteCityv6.dat.gz
wget -N http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNum.dat.gz
wget -N http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNumv6.dat.gz
gzip -df *.gz

cp -f GeoIP.dat /opt/local/share/GeoIP/
cp -f GeoLiteCity.dat /opt/local/share/GeoIP/
