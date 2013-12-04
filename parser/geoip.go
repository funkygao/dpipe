/*
map ip to geo point so that elasticsearch can support geo based points.
The declaration looks as follows:
{
    "pin" : {
        "properties" : {
            "location" : {
                "type" : "geo_point"
            }
        }
    }
}

PUT
{
    "pin" : {
        "location" : {
            "lat" : 41.12,
            "lon" : -71.34
        }
    }
}
*/
package parser

import (
	"github.com/abh/geoip"
)

var geo *geoip.GeoIP

type geoPoint struct {
	Lat float32 `json:"lat"`
	Lon float32 `json:"lon"`
}

func loadGeoDb(geodbfile string) {
	var err error
	geo, err = geoip.Open(geodbfile)
	if err != nil {
		panic(err)
	}
}

func ipToGeo(ip string) (this geoPoint) {
	if rec := geo.GetRecord(ip); rec != nil {
		this = geoPoint{Lat: rec.Latitude, Lon: rec.Longitude}
	}

	return
}
