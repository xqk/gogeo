package gogeo

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed testdata/100000_full.json
var jsonProvince []byte

//go:embed testdata/230000_full.json
var jsonCity []byte

func TestNewGeoMap(t *testing.T) {
	_, err := NewGeoMap("testdata/100000_full.json", "name")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewGeoMapFromBytes(t *testing.T) {
	_, err := NewGeoMapFromBytes(jsonProvince, "name")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewGeoMapFormat(t *testing.T) {
	ff := func(ks []string) string {
		return ks[0] + "-" + ks[1]
	}

	_, err := NewGeoMapFormat("testdata/230000_full.json", []string{"name", "level"}, ff)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewGeoMapFormatFromBytes(t *testing.T) {
	ff := func(ks []string) string {
		return ks[0] + "-" + ks[1]
	}

	_, err := NewGeoMapFormatFromBytes(jsonCity, []string{"name", "level"}, ff)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGeoMap_FindLoc(t *testing.T) {
	m, err := NewGeoMap("testdata/100000_full.json", "name")
	if err != nil {
		t.Fatal(err)
	}

	if m.FindLoc(39.905241, 116.397682) != "北京市" {
		t.Error("位置不匹配")
	}

	if m.FindLoc(53.467057, 108.156513) != StrNotFound {
		t.Error("位置不匹配")
	}
}

func TestGeoMap_FindLocFormat(t *testing.T) {
	ff := func(ks []string) string {
		return ks[0] + "-" + ks[1]
	}

	m, err := NewGeoMapFormat("testdata/230000_full.json", []string{"name", "level"}, ff)
	if err != nil {
		t.Fatal(err)
	}

	if m.FindLoc(45.648633, 127.966759) != "哈尔滨市-city" {
		t.Error("位置不匹配")
	}

	if m.FindLoc(53.467057, 108.156513) != StrNotFound {
		t.Error("位置不匹配")
	}
}

func TestGeoMap_FindDistrictLocFormat(t *testing.T) {
	ff := func(ks []string) string {
		return ks[0] + "-" + ks[1]
	}

	m, err := NewGeoMapFormat("testdata/230100_full.json", []string{"name", "level"}, ff)
	if err != nil {
		t.Fatal(err)
	}

	if m.FindLoc(45.668664, 126.373347) != "道里区-district" {
		t.Error("位置不匹配")
	}

	if m.FindLoc(53.467057, 108.156513) != StrNotFound {
		t.Error("位置不匹配")
	}
}

func TestGeoMap_ContainLoc(t *testing.T) {
	m, err := NewGeoMap("testdata/100000_full.json", "name")
	if err != nil {
		t.Fatal(err)
	}

	if !m.ContainLoc(39.905241, 116.397682) {
		t.Error("位置不匹配")
	}

	if m.ContainLoc(53.467057, 108.156513) {
		t.Error("位置不匹配")
	}
}

func TestNewGroupGeoMapFormatFromRootAdCode(t *testing.T) {
	ff := func(ks []string) string {
		return ks[0] + "-" + ks[1] + "-" + ks[2]
	}

	groupGeoMap, err := NewGroupGeoMapFormatFromRootAdCode("testdata", "100000", []string{"name", "level", "adcode"}, ff)
	if err != nil {
		t.Error(err)
		return
	}

	for adCode, geoMap := range groupGeoMap {
		fmt.Println(adCode, geoMap.AdCodes)
		fmt.Println(geoMap.FindLoc(39.905241, 116.397682))
	}
}

func TestMasterGeo_FindLoc(t *testing.T) {
	ff := func(ks []string) string {
		return ks[0] + "-" + ks[1] + "-" + ks[2]
	}

	masterGeo, err := NewMasterGeo("/Users/xqk/Downloads/map-files/full", "100000", []string{"name", "level", "adcode"}, ff, "-")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("加载完毕")
	locs := masterGeo.FindLoc(45.849582, 127.06789)
	for _, loc := range locs {
		fmt.Println(loc)
	}
}
