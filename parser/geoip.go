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

type GeoPoint struct {
	Lat float32 `json:"lat"`
	Lon float32 `json:"lon"`
}

func loadGeoDb(geodbfile string) (err error) {
	geo, err = geoip.Open(geodbfile)
	return
}

func geoEnabled() bool {
	return geo != nil
}

func ipToGeo(ip string) (this GeoPoint) {
	if rec := geo.GetRecord(ip); rec != nil {
		this = GeoPoint{Lat: rec.Latitude, Lon: rec.Longitude}
	}

	return
}

func ipToCountry(ip string) string {
    country, _ := geo.GetCountry(ip)
    return country
}
