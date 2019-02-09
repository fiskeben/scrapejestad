package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Row struct {
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

type Position struct {
	Lat float32
	Lng float32
}

type Gateway struct {
	Name          string
	Position      Position
	Distance      float32
	RSSI          float32
	LSNR          float32
	RadioSettings RadioSettings
}

type RadioSettings struct {
	Frequency float32
	Sf        string
	Cr        string
}

func main() {
	r, err := read()
	if err != nil {
		panic(err)
	}
	_, err = parse(r)
	if err != nil {
		panic(err)
	}
}

func read() (io.Reader, error) {
	return os.Open("example.html")
}

func parse(r io.Reader) ([]Row, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return parseSubtree(doc)
}

func parseSubtree(n *html.Node) ([]Row, error) {
	fmt.Printf("checking %s\n", n.Data)
	c := n.FirstChild
	fmt.Printf("child %s\n", c.Data)
	if n.Type == html.ElementNode && n.Data == "table" {
		return parseTable(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		return parseSubtree(c)
	}
	return nil, errors.New("this should not happen")
}

func parseTable(t *html.Node) ([]Row, error) {
	rows := make([]Row, 0, 10)
	for c := t.FirstChild; c != nil; c = c.NextSibling {
		nodes := prepareRow(c)
		switch len(nodes) {
		case 0:
			continue
		case 5:
			g, err := parseGateway(nodes)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
			fmt.Println("hellu")
			row := rows[len(rows)-1]
			row.Gateways = append(row.Gateways, g)
		case 16:
			row, err := parseRow(nodes)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
			rows = append(rows, *row)
			fmt.Printf("row %v\n", row)
		default:
			fmt.Printf("node %v has unexpected number of nodes: %d\n", c, len(nodes))
		}
	}
	return rows, nil
}

func parseRow(n []*html.Node) (*Row, error) {
	var r Row

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

	fmt.Printf("%s %s\n", n[0].Data, n[0].FirstChild.FirstChild.Data)
	g.Name = strings.TrimSpace(n[0].FirstChild.FirstChild.Data)

	data := strings.TrimSpace(n[1].FirstChild.Data)
	v := data[:len(data)-2]
	dist, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return g, err
	}
	g.Distance = float32(dist)

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

func prepareRow(n *html.Node) []*html.Node {
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
