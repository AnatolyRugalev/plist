package plist

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"
)

var decodeTests = []struct {
	out interface{}
	in  string
}{
	{"foo", fooRef},
	{"UTF-8 ☼", utf8Ref},
	// {0, zeroRef},
	// {1, oneRef},
	{uint64(1), oneRef},
	// {-1, minOneRef},
	{1.2, realRef},
	{false, falseRef},
	{true, trueRef},
	// {[]interface{}{"a", "b", "c", 4, true}, arrRef},
	{time.Date(1900, 01, 01, 12, 00, 00, 0, time.UTC), time1900Ref},
	// {[]byte(fooRef), dataRef},
	// {map[string]interface{}{
	// 	"foo":  "bar",
	// 	"bool": true},
	// 	dictRef},
	// {struct {
	// 	Foo  string `plist:"foo"`
	// 	Bool bool   `plist:"bool"`
	// }{"bar", true},
	// 	dictRef},
}

func TestDecode(t *testing.T) {
	for _, tt := range decodeTests {
		var out interface{}
		if err := Unmarshal([]byte(tt.in), &out); err != nil {
			t.Error(err)
			continue
		}
		if out != tt.out {
			t.Errorf("Unmarshal(%v) = \n%v, want %v", tt.in, out, tt.out)
		}
	}
}

func TestDecodeReal(t *testing.T) {
	buf := bytes.NewReader([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><real>1.5</real></plist>
`))
	// test float64
	var data float64
	if err := NewDecoder(buf).Decode(&data); err != nil {
		t.Fatal(err)
	}
	if data != 1.5 {
		t.Error("Expected", 1.5, "got", data)
	}
	buf.Seek(0, 0)
	// test float32
	var data32 float32
	if err := NewDecoder(buf).Decode(&data32); err != nil {
		t.Fatal(err)
	}
	if data32 != 1.5 {
		t.Error("Expected", 1.5, "got", data32)
	}
	buf.Seek(0, 0)
	//test error
	var dataErr string
	if err := NewDecoder(buf).Decode(&dataErr); err == nil {
		t.Fatal("Expected UnmarshalTypeError got nil")
	}
}

func TestDecodeArray(t *testing.T) {
	buf := bytes.NewReader([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0">
<array>
	<string>object</string>
</array>
</plist>
`))
	var data []string
	if err := NewDecoder(buf).Decode(&data); err != nil {
		t.Fatal(err)
	}
	if data[0] != "object" {
		t.Error("Expected", "object", "got", data[0])
	}
	buf.Seek(0, 0)
	//test err
	var errdata []bool
	err := NewDecoder(buf).Decode(&errdata)
	if err == nil {
		t.Fatal("Expected UnmarshalTypeError got nil")
	}
}

func TestDecodeBoolean(t *testing.T) {
	buf := bytes.NewReader([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0">
<array>
<true/>
<false/>
<true></true>
<false></false>
</array>
</plist>
`))
	var data []bool
	if err := NewDecoder(buf).Decode(&data); err != nil {
		t.Fatal(err)
	}
	if data[0] != true {
		t.Error("Expected", true, "got", data[0])
	}
	if data[1] != false {
		t.Error("Expected", false, "got", data[1])
	}
	if data[2] != true {
		t.Error("Expected", true, "got", data[2])
	}
	if data[3] != false {
		t.Error("Expected", false, "got", data[3])
	}
}

func TestDecodeDict(t *testing.T) {
	type sparseBundleHeader struct {
		InfoDictionaryVersion string `plist:"CFBundleInfoDictionaryVersion"`
		BandSize              uint64 `plist:"band-size"`
		BackingStoreVersion   int    `plist:"bundle-backingstore-version"`
		DiskImageBundleType   string `plist:"diskimage-bundle-type"`
		Size                  uint64 `plist:"size"`
	}
	buf := bytes.NewReader([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>CFBundleInfoDictionaryVersion</key>
		<string>6.0</string>
		<key>band-size</key>
		<integer>8388608</integer>
		<key>bundle-backingstore-version</key>
		<integer>1</integer>
		<key>diskimage-bundle-type</key>
		<string>com.apple.diskimage.sparsebundle</string>
		<key>size</key>
		<integer>4398046511104</integer>
	</dict>
</plist>`))
	var data sparseBundleHeader
	decoder := NewDecoder(buf)
	err := decoder.Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if data.InfoDictionaryVersion != "6.0" {
		t.Errorf("Expected %v, got %v", "6.0", data.InfoDictionaryVersion)
	}
	if data.BandSize != 8388608 {
		t.Errorf("Expected %v, got %v", 8388608, data.BandSize)
	}

	// Output: {6.0 8388608 1 com.apple.diskimage.sparsebundle 4398046511104}
	buf.Seek(0, 0)
	var mapData map[string]interface{}
	if err := NewDecoder(buf).Decode(&mapData); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeData(t *testing.T) {
	buf := bytes.NewReader([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>TestData</key>
		<data>Zm9vYmFy</data>
	</dict>
</plist>`))
	var data struct {
		TestData []byte
	}
	if err := NewDecoder(buf).Decode(&data); err != nil {
		t.Fatal(err)
	}
	testData := []byte("foobar")
	str64 := base64.StdEncoding.EncodeToString(testData)
	if string(data.TestData) != str64 {
		t.Errorf("Expected %v, got %v", str64, string(data.TestData))
	}
}

func TestDecodeTime(t *testing.T) {
	buf := bytes.NewReader([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>TestDate</key>
		<date>2015-09-05T21:55:30Z</date>
	</dict>
</plist>`))
	var date struct {
		TestDate time.Time
	}
	if err := NewDecoder(buf).Decode(&date); err != nil {
		t.Fatal(err)
	}
	if date.TestDate.Year() != 2015 {
		t.Error("Expected", 2015, "got", date.TestDate.Year())
	}
}
