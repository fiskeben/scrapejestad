package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

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

type Position struct {
	Lat float32
	Lng float32
}

func (p Position) String() string {
	return fmt.Sprintf("%f:%f", p.Lat, p.Lng)
}

type Gateway struct {
	Name          string
	Position      Position
	Distance      float32
	RSSI          float32
	LSNR          float32
	RadioSettings RadioSettings
}

func (g Gateway) String() string {
	return fmt.Sprintf("Name=%s Position=%s Distance=%f RSSI=%f LSNR=%f Radiosettings=%s", g.Name, g.Position.String(), g.Distance, g.RSSI, g.LSNR, g.RadioSettings.String())
}

type RadioSettings struct {
	Frequency float32
	Sf        string
	Cr        string
}

func (r RadioSettings) String() string {
	return fmt.Sprintf("Frequency=%f Sf=%s Cr=%s", r.Frequency, r.Sf, r.Cr)
}

func main() {
	r, err := read()
	if err != nil {
		panic(err)
	}
	res, err := parse(r)
	if err != nil {
		panic(err)
	}

	for _, r := range res {
		fmt.Println(r)
		fmt.Println()
	}
}

func read() (io.Reader, error) {
	return os.Open("example.html")
}

func parse(r io.Reader) ([]Reading, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return parseSubtree(doc)
}

func parseSubtree(n *html.Node) ([]Reading, error) {
	if n.Type == html.ElementNode && n.Data == "table" {
		return parseTable(n.FirstChild.NextSibling)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res, err := parseSubtree(c)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, nil
}

func parseTable(t *html.Node) ([]Reading, error) {
	rows := make([]Reading, 0, 10)
	for c := t.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "tr" {
			continue
		}
		nodes := mapRow(c)
		switch len(nodes) {
		case 0:
			continue
		case 5:
			g, err := parseGateway(nodes)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
			row := rows[len(rows)-1]
			row.Gateways = append(row.Gateways, g)
			rows[len(rows)-1] = row
		case 16:
			row, err := parseRow(nodes)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
			rows = append(rows, *row)
		default:
			fmt.Printf("node %v has unexpected number of nodes: %d\n", c, len(nodes))
		}
	}
	return rows, nil
}

func parseRow(n []*html.Node) (*Reading, error) {
	var r Reading

	r.ID = getID(n[0])

	data := strings.TrimSpace(n[1].FirstChild.Data)
	t, err := time.Parse("2006-01-02 15:04:05", data)
	if err != nil {
		return nil, err
	}
	r.Time = t

	data = strings.TrimSpace(n[2].FirstChild.Data)
	v := data[:len(data)-3]
	temp, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	r.Temp = float32(temp)

	data = strings.TrimSpace(n[3].FirstChild.Data)
	v = data[:len(data)-1]
	h, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	r.Humidity = float32(h)

	data = strings.TrimSpace(n[7].FirstChild.Data)
	v = data[:len(data)-1]
	p, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	r.Voltage = float32(p)

	r.Firmware = strings.TrimSpace(n[8].FirstChild.Data)

	parts := strings.Split(strings.TrimSpace(n[9].FirstChild.NextSibling.FirstChild.Data), " ")
	lat, err := strconv.ParseFloat(parts[0], 32)
	if err != nil {
		return nil, err
	}
	lng, err := strconv.ParseFloat(parts[len(parts)-1], 32)
	if err != nil {
		return nil, err
	}
	r.Position = Position{Lat: float32(lat), Lng: float32(lng)}

	data = strings.TrimSpace(n[10].FirstChild.Data)
	fcnt, err := strconv.Atoi(data)
	if err != nil {
		return nil, err
	}
	r.Fcnt = fcnt

	g, err := parseGateway(n[11:])
	if err != nil {
		return nil, err
	}
	r.Gateways = []Gateway{g}

	return &r, nil
}

func parseGateway(n []*html.Node) (Gateway, error) {
	// TODO parse URL to get position
	var g Gateway

	parent := n[0].FirstChild
	if parent.FirstChild != nil {
		g.Name = strings.TrimSpace(parent.FirstChild.Data)
	}

	data := strings.TrimSpace(n[1].FirstChild.Data)
	if len(data) > 2 {
		v := data[:len(data)-2]
		dist, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return g, err
		}
		g.Distance = float32(dist)
	}

	rssi, err := strconv.ParseFloat(strings.TrimSpace(n[2].FirstChild.Data), 32)
	if err != nil {
		return g, err
	}
	g.RSSI = float32(rssi)

	lsnr, err := strconv.ParseFloat(strings.TrimSpace(n[3].FirstChild.Data), 32)
	if err != nil {
		return g, err
	}
	g.LSNR = float32(lsnr)

	parts := strings.Split(strings.TrimSpace(n[4].FirstChild.Data), ",")
	freq, err := strconv.ParseFloat(parts[0][:len(parts[0])-3], 32)
	if err != nil {
		return g, err
	}
	s := RadioSettings{
		Frequency: float32(freq),
		Sf:        strings.TrimSpace(parts[1]),
		Cr:        strings.TrimSpace(parts[2]),
	}
	g.RadioSettings = s

	return g, nil
}

func mapRow(n *html.Node) []*html.Node {
	if n.FirstChild.Data == "th" {
		return make([]*html.Node, 0)
	}
	res := make([]*html.Node, 0, 5)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Data != "td" {
			continue
		}
		res = append(res, c)
	}
	return res
}

func getID(n *html.Node) string {
	c := n.FirstChild
	if id := strings.TrimSpace(c.Data); id != "" {
		return id
	}
	c = c.NextSibling
	if id := strings.TrimSpace(c.Data); id != "" && id != "a" {
		return id
	}
	c = c.FirstChild
	if id := strings.TrimSpace(c.Data); id != "" && id != "a" {
		return id
	}
	return ""
}
