Chromaticity
============

Chromaticity exposes a Hue compatible REST API that has a modular
backend design, allowing a user to control Color Kinetics,
MiLight/LimitlessLED\*, OPC (Open Pixel Control) and Hue\* devices through a
unified api.

\*: Planned

Current State: In progress...
-----------------------------

Basic functionality exists for Color Kinetics and OPC lights.

* Control individual lights and groups of lights
* Set color with the HSV, or Yxy schemes
* Set color temperature in Mirads

Exposes each connected fixture of a controller as as separate light
and automatically creates one group per controller.

No renaming, color effect, alert, color schemes, or other advanced features :)

Auth is currently limited to "foo" user only and discovery (of the api itself)
isn't implemented so don't expect most Hue clients to work at this time.

Usage
-----

```bash
go get github.com/evq/chromaticity
go install github.com/evq/chromaticity
chromaticity
```

A swagger ui can be enabled by uncommenting the relevant block in 
apiserver.go and changing the file path to where you have checked out
the swagger ui.

Chromaticity will read stored (json serialized) light information from
`~/.chromaticity/data.json` Future improvements will write discovered light
information to this file, as well as read configuration from elsewhere in
the chromaticity directory.

Example:
```json
{
  "kinet": {
    "powerSupplies": [
      {
        "Name": "eV-Lights",
        "IP": "192.168.1.47",
        "Mac": "00:0a:c5:ff:00:00",
        "ProtocolVersion": "1",
        "Serial": "20000000",
        "Universe": "0",
        "Manufacturer": "Color Kinetics Incorporated",
        "Type": "PDS-e",
        "FWVersion": "SFT-000066-00",
        "Fixtures": [
          {
            "Serial": "1E000000",
            "Channel": 6,
            "Color": {
              "R": 0,
              "G": 255,
              "B": 0,
              "A": 255
            }
          }
        ]
      }
    ]
  },
  "opc": {
    "servers" : [
      {
        "name": "eV-String-Lights",
        "host": "192.168.1.169",
        "port": "7890",
        "channels": [
          {
            "id": 0,
            "numPixels": 60
          }
        ]
      }
    ]
  }
}
```
See Other
--------

[go-opc](https://github.com/kellydunn/go-opc/): The golang library used to
provide an OPC backend for chromaticity
