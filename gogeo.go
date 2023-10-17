package gogeo

import (
	"errors"
	"fmt"
	geo "github.com/kellydunn/golang-geo"
	geojson "github.com/paulmach/go.geojson"
	"io/ioutil"
	"strconv"
	"strings"
)

const StrNotFound = "NotFound"

type GeoMap struct {
	GMap    map[string][]*geo.Polygon
	AdCodes []string
}

type MasterGeo struct {
	GroupGMap  map[string]*GeoMap
	RootAdCode string
	Keys       []string
	Delimiter  string
}

func NewMasterGeo(dirPath, rootAdCode string, keys []string, ff func(ks []string) string, delimiter string) (*MasterGeo, error) {
	groupGeoMap, err := NewGroupGeoMapFormatFromRootAdCode(dirPath, rootAdCode, keys, ff)
	if err != nil {
		return nil, err
	}

	if delimiter == "" {
		delimiter = "-"
	}

	return &MasterGeo{GroupGMap: groupGeoMap, RootAdCode: rootAdCode, Keys: keys, Delimiter: delimiter}, nil
}

func (m *MasterGeo) FindLoc(lat, lng float64) []map[string]string {
	oneLevelGeoMap := m.GroupGMap[m.RootAdCode]
	if oneLevelGeoMap == nil {
		return nil
	}
	locs := make([]map[string]string, 0)
	oneLevelLocInfo := m.getLocInfo(oneLevelGeoMap, lat, lng)
	if oneLevelLocInfo == nil {
		return locs
	}
	locs = append(locs, oneLevelLocInfo)

	if oneLevelLocInfo["adcode"] != "" {
		twoLevelGeoMap := m.GroupGMap[oneLevelLocInfo["adcode"]]
		if twoLevelGeoMap == nil {
			return locs
		}
		twoLevelLocInfo := m.getLocInfo(twoLevelGeoMap, lat, lng)
		if twoLevelLocInfo == nil {
			return locs
		}
		locs = append(locs, twoLevelLocInfo)
		if twoLevelLocInfo["adcode"] != "" {
			threeLevelGeoMap := m.GroupGMap[twoLevelLocInfo["adcode"]]
			if threeLevelGeoMap == nil {
				return locs
			}
			threeLevelLocInfo := m.getLocInfo(threeLevelGeoMap, lat, lng)
			if threeLevelLocInfo == nil {
				return locs
			}
			locs = append(locs, threeLevelLocInfo)
		}
	}

	return locs
}

func (m *MasterGeo) getLocInfo(geoMap *GeoMap, lat, lng float64) map[string]string {
	locInfo := make(map[string]string)
	loc := geoMap.FindLoc(lat, lng)
	if loc == StrNotFound {
		return nil
	}

	locItems := strings.Split(loc, m.Delimiter)
	if len(locItems) != len(m.Keys) {
		return nil
	}

	for i, key := range m.Keys {
		locInfo[key] = locItems[i]
	}

	return locInfo
}

func getPolyMap(data []byte, keys []string, ff func(ks []string) string) (map[string][]*geo.Polygon, []string, error) {
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, nil, err
	}

	polysMap := make(map[string][]*geo.Polygon, 0)
	adCodes := make([]string, 0)
	for _, v := range fc.Features {
		pKeys := make([]string, len(keys))
		var adCode string
		for i, key := range keys {
			if _, ok := v.Properties[key]; !ok {
				return nil, nil, errors.New(fmt.Sprintf("file has no key:%v in some features", key))
			}
			if key == "adcode" {
				adCode = strconv.FormatInt(int64(v.Properties["adcode"].(float64)), 10)
				pKeys[i] = adCode
			} else {
				pKeys[i] = v.Properties[key].(string)
			}
		}
		if adCode != "" {
			adCodes = append(adCodes, adCode)
		}
		key := ff(pKeys)

		geometry := v.Geometry
		mps := make([][][][]float64, 0)
		if geometry.Type == "MultiPolygon" {
			mps = geometry.MultiPolygon

		} else if geometry.Type == "Polygon" {
			mps = append(mps, geometry.Polygon)
		}

		for _, polygon := range mps {
			tmpPointList := make([]*geo.Point, 0)
			for _, points := range polygon {
				for _, point := range points {
					tmpPoint := geo.NewPoint(point[1], point[0])
					tmpPointList = append(tmpPointList, tmpPoint)
				}
			}
			polysMap[key] = append(polysMap[key], geo.NewPolygon(tmpPointList))
		}
	}
	return polysMap, adCodes, nil
}

func NewGeoMap(file, key string) (*GeoMap, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	g, adCodes, e := getPolyMap(data, []string{key}, func(ks []string) string { return ks[0] })
	if e != nil {
		return nil, e
	}
	return &GeoMap{GMap: g, AdCodes: adCodes}, nil
}

func NewGeoMapFromBytes(data []byte, key string) (*GeoMap, error) {
	g, adCodes, e := getPolyMap(data, []string{key}, func(ks []string) string { return ks[0] })
	if e != nil {
		return nil, e
	}
	return &GeoMap{GMap: g, AdCodes: adCodes}, nil
}

func NewGeoMapFormat(file string, keys []string, ff func(ks []string) string) (*GeoMap, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	g, adCodes, e := getPolyMap(data, keys, ff)
	if e != nil {
		return nil, e
	}
	return &GeoMap{GMap: g, AdCodes: adCodes}, nil
}

func NewGeoMapFormatFromBytes(data []byte, keys []string, ff func(ks []string) string) (*GeoMap, error) {
	g, adCodes, e := getPolyMap(data, keys, ff)
	if e != nil {
		return nil, e
	}
	return &GeoMap{GMap: g, AdCodes: adCodes}, nil
}

func NewGroupGeoMapFormatFromRootAdCode(dirPath, rootAdCode string, keys []string, ff func(ks []string) string) (map[string]*GeoMap, error) {
	var err error
	groupGeoMap := make(map[string]*GeoMap)

	groupGeoMap, err = aloneGroupGeoMapFormatFromRootAdCode(groupGeoMap, dirPath, rootAdCode, keys, ff)
	if err != nil {
		return nil, err
	}

	return groupGeoMap, nil
}

func aloneGroupGeoMapFormatFromRootAdCode(groupGeoMap map[string]*GeoMap, dirPath, rootAdCode string, keys []string, ff func(ks []string) string) (map[string]*GeoMap, error) {
	rootAdCodeFile := fmt.Sprintf("%s/%s_full.json", dirPath, rootAdCode)
	rootMap, err := NewGeoMapFormat(rootAdCodeFile, keys, ff)
	if err != nil {
		return nil, err
	}
	groupGeoMap[rootAdCode] = rootMap

	for _, adCode := range rootMap.AdCodes {
		newGroupGeoMap, err1 := aloneGroupGeoMapFormatFromRootAdCode(groupGeoMap, dirPath, adCode, keys, ff)
		if err1 != nil {
			continue
		}
		groupGeoMap = newGroupGeoMap
	}
	return groupGeoMap, nil
}

func (g *GeoMap) FindPoint(tmpPoint *geo.Point) string {
	for name, polys := range g.GMap {
		for _, poly := range polys {
			if poly.Contains(tmpPoint) {
				return name
			}
		}
	}
	return StrNotFound
}

func (g *GeoMap) ContainPoint(tmpPoint *geo.Point) bool {
	for _, polys := range g.GMap {
		for _, poly := range polys {
			if poly.Contains(tmpPoint) {
				return true
			}
		}
	}
	return false
}

func (g *GeoMap) FindLoc(lat, lng float64) string {
	tmpPoint := geo.NewPoint(lat, lng)
	return g.FindPoint(tmpPoint)
}

func (g *GeoMap) ContainLoc(lat, lng float64) bool {
	tmpPoint := geo.NewPoint(lat, lng)
	return g.ContainPoint(tmpPoint)
}
