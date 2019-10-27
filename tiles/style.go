package tiles

import (
	"fmt"
	"io"
	"text/template"
)

const LocalHost = "http://localhost:8008"
const Eart7hHost = "http://ear7h.net:8008"

var Host = LocalHost

func Capabilities() string {
	return fmt.Sprintf("%s/capabilities/ne.json", Host)
}

func Consumer(layer string) string {
	return fmt.Sprintf("%s/maps/ne/user_geo/%s", Host, layer)
}

func WriteUserStyle(w io.Writer, user string) error {
	return userStyleJsonTmpl.Execute(w, map[string]string{
		"Url":      Capabilities(),
		"UserName": user,
	})
}

func WriteEntryStyle(w io.Writer, user, entry string) error {
	return entryStyleJsonTmpl.Execute(w, map[string]string{
		"Url":      Capabilities(),
		"UserName": user,
		"EntryName":    entry,
	})

}

var userStyleJsonTmpl = template.Must(template.New("").Parse(userStyleJsonTmplStr))
var entryStyleJsonTmpl = template.Must(template.New("").Parse(entryStyleJsonTmplStr))

const userStyleJsonTmplStr = `{
  "version": 8,
  "name": "ne",
  "center": [
    -76.275329586789,
    39.153492567373
  ],
  "zoom": 5,
  "sources": {
    "ne": {
      "type": "vector",
      "url": "{{ .Url }}"
    }
  },
  "layers": [
    {
      "id": "entry_geotag_points",
      "source": "ne",
      "source-layer": "entry_geotag_points",
      "type": "circle",
      "layout": {
        "visibility": "visible"
      },
      "paint": {
        "circle-radius": 3,
        "circle-color": "#1dd091"
      },
	  "filter": [
		"==", ["get", "user_name"], "{{ .UserName }}"
	  ]
   },
    {
      "id": "coastline",
      "source": "ne",
      "source-layer": "coastline",
      "type": "line",
      "layout": {
        "visibility": "visible"
      },
      "paint": {
        "line-color": "#8a7b28"
      }
    }
  ]
}`

const entryStyleJsonTmplStr = `{
  "version": 8,
  "name": "ne",
  "center": [
    -76.275329586789,
    39.153492567373
  ],
  "zoom": 5,
  "sources": {
    "ne": {
      "type": "vector",
      "url": "{{ .Url }}"
    }
  },
  "layers": [
    {
      "id": "entry_geotag_points",
      "source": "ne",
      "source-layer": "entry_geotag_points",
      "type": "circle",
      "layout": {
        "visibility": "visible"
      },
      "paint": {
        "circle-radius": 3,
        "circle-color": "#1dd091"
      },
	  "filter": [
		"all",
		["==", ["get", "user_name"], "{{ .UserName }}"],
		["==", ["get", "entry_name"], "{{ .EntryName }}"]
	  ]
   },
    {
      "id": "coastline",
      "source": "ne",
      "source-layer": "coastline",
      "type": "line",
      "layout": {
        "visibility": "visible"
      },
      "paint": {
        "line-color": "#8a7b28"
      }
    }
  ]
}`
