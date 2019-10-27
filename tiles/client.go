package tiles

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/geojson"

	"github.com/go-spatial/proj"
)

func AddEntryGeotagPoint(latLngPt geom.Point, user, entry, asset string) error {
	pts, err := proj.Convert(proj.EPSG3857, latLngPt[:])
	if err != nil {
		return err
	}
	pt := geom.Point{pts[0], pts[1]}

	feat := geojson.Feature{
		Geometry: geojson.Geometry{Geometry:pt},
		Properties: map[string]interface{}{
			"SRID": 3857,
			"user_name":  user,
			"entry_name": entry,
			"asset_name": asset,
		},
	}

	buf := &bytes.Buffer{}

	err = json.NewEncoder(buf).Encode(feat)
	if err != nil {
		return err
	}

	_, err = http.Post(Consumer("entry_geotag_points"), "application/json", buf)
	return err
}
