TODO
====
* Handle errors
* Fix backends to reuse connections
* Use `json:",string"` option for light ids etc?
* Add more debug logging using logrus
* Switch to gopkg.in for versioning?
* Need to add discover devices support for zigbee backend
* Persist groups and discovered lights to disk
* Handle group membership changes due to light removal (id change, less
  members)
* For above consider having globally unique identifier per light and having
  group persistence use that identifier
