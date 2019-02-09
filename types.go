package scrapejestad

import (
	"fmt"
	"strings"
	"time"
)

// Reading represents one unique data point.
type Reading struct {
	ID       string
	Time     time.Time
	Temp     float32
	Humidity float32
	Light    float32
	PM25     float32
	PM10     float32
	Voltage  float32
	Firmware string
	Position Position
	Fcnt     int
	Gateways []Gateway
}

// String returns a string representation of a Reading.
func (r Reading) String() string {
	s := fmt.Sprintf(`ID=%s
Time=%s
Temp=%f
Humidity=%f
Light=%f
PM25=%f
PM10=%f
Voltage=%f
Firmware=%s
Position=%s
Fcnt=%d`, r.ID, r.Time.Format(time.RFC3339), r.Temp, r.Humidity, r.Light, r.PM25, r.PM10, r.Voltage, r.Firmware, r.Position.String(), r.Fcnt)
	gateways := make([]string, len(r.Gateways))
	for i, g := range r.Gateways {
		gateways[i] = fmt.Sprintf("  %d %s\n", i, g.String())
	}
	s = fmt.Sprintf("%s\nGateways:\n%s", s, strings.Join(gateways, " "))
	return s
}

// Position is a coordinate with latitude and longitude.
type Position struct {
	Lat float32
	Lng float32
}

// String returns a position as lat:lng.
func (p Position) String() string {
	return fmt.Sprintf("%f:%f", p.Lat, p.Lng)
}

// Gateway holds data about a LoRa:wan gateway a Reading has been sent to.
type Gateway struct {
	Name          string
	Position      Position
	Distance      float32
	RSSI          float32
	LSNR          float32
	RadioSettings RadioSettings
}

// String returns a string representation of a Gateway.
func (g Gateway) String() string {
	return fmt.Sprintf("Name=%s Position=%s Distance=%f RSSI=%f LSNR=%f Radiosettings=%s", g.Name, g.Position.String(), g.Distance, g.RSSI, g.LSNR, g.RadioSettings.String())
}

// RadioSettings holds data about the radio settings used to transmit a Reading.
type RadioSettings struct {
	Frequency float32
	Sf        string
	Cr        string
}

// String returns a string representation of RadioSettings.
func (r RadioSettings) String() string {
	return fmt.Sprintf("Frequency=%f Sf=%s Cr=%s", r.Frequency, r.Sf, r.Cr)
}
