package scrapejestad

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func mktime(d string) int64 {
	t, err := time.Parse("2006-01-02 15:04:05", d)
	if err != nil {
		panic(err)
	}
	return t.Unix()
}

func mkdate(d string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", d)
	if err != nil {
		panic(err)
	}
	return t
}

func Test_parsing(t *testing.T) {
	r, err := os.Open("testdata/example.html")
	if err != nil {
		t.Fatalf("failed to open testdata: %v", err)
	}
	res, err := parse(r)
	if err != nil {
		t.Fatalf("error parsing testdata: %v", err)
	}

	results := []Reading{
		{
			SensorID: "242",
			Time: mktime("2019-12-05 21:19:33"),
			Date: mkdate("2019-12-05 21:19:33"),
			Temp: 6.875,
			Humidity: 107.25,
			Light: 0,
			PM25: 0,
			PM10: 0,
			Voltage: 3.37,
			Firmware: "v2",
			Position: Position{
				Lat: 60.430900,
				Lng: 5.2325101,
			}, Fcnt: 28357,
			Gateways: []Gateway{
				{
					Name: "florvaag-1",
					Position: Position{Lat: 60.431778, Lng: 5.231865},
					Distance: 0.104,
					RSSI: -47,
					LSNR: 9.5,
					RadioSettings: RadioSettings{Frequency: 868.5, Sf: "SF9BW125", Cr: "4/5CR"},
				},
				{
					Name: "eui-00f142122877fa05",
					Position: Position{Lat: 60.41283, Lng: 5.327483},
					Distance: 5.587,
					RSSI: -117,
					LSNR: -1,
					RadioSettings: RadioSettings{Frequency: 868.5, Sf: "SF9BW125", Cr: "4/5CR"},
				},
			},
		},
		{
			SensorID: "242",
			Time: mktime("2019-12-05 21:02:39"),
			Date: mkdate("2019-12-05 21:02:39"),
			Temp: 6.875,
			Humidity: 107.312,
			Light: 0,
			PM25: 0,
			PM10: 0,
			Voltage: 3.37,
			Firmware: "v2",
			Position: Position{
				Lat: 60.4309,
				Lng: 5.23251,
			},
			Fcnt: 28356,
			Gateways: []Gateway{
				{
					Name: "florvaag-1",
					Position: Position{Lat: 60.431778, Lng: 5.231865},
					Distance: 0.104,
					RSSI: -45,
					LSNR: 12.25,
					RadioSettings: RadioSettings{Frequency: 867.7, Sf: "SF9BW125", Cr: "4/5CR"},
				},
				{
					Name: "mjs-bergen-gateway-5",
					Position: Position{Lat: 60.389248, Lng: 5.285356},
					Distance: 5.465,
					RSSI: -113,
					LSNR: -10,
					RadioSettings: RadioSettings{Frequency: 867.7, Sf: "SF9BW125", Cr: "4/5CR"},
				},
			},
		},
		}

	if len(res) != len(results) {
		t.Errorf("expected %d results, got %d", len(results), len(res))
	}

	for i, r := range results {
		if diff := cmp.Diff(res[i], r); diff != "" {
			t.Errorf("%d not equal: %v", i, diff)
		}
	}
}

func Test_extractPositionPart(t *testing.T) {
	uri := "http://www.openstreetmap.org/?mlat=60.362316&amp;mlon=5.340381"
	lat, err := extractPositionPart(uri, "mlat")
	if err != nil {
		t.Fatalf("error extracting latitiude: %v", err)
	}
	if lat != 60.362316 {
		t.Errorf("expected latitude to be 60.362316, got %f", lat)
	}
	lng, err := extractPositionPart(uri, "mlon")
	if err != nil {
		t.Fatalf("error extracting longitude: %v", err)
	}
	if lng != 5.340381 {
		t.Errorf("expected longitude to be 5.340381, got %f", lng)
	}
}

func Test_handleMissingData(t *testing.T) {
	f, err := os.Open("testdata/missing_data.html")
	if err != nil {
		t.Fatalf("failed to open testdata: %v", err)
	}
	res, err := parse(f)
	if err != nil {
		t.Fatalf("error parsing data: %v", err)
	}
	if len(res) != 3 {
		t.Errorf("expected 3 readings, got %d", len(res))
	}

}
