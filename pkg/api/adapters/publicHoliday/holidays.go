package publicHoliday

import (
	"context"
	"fmt"
	"github.com/wlbr/feiertage"
	"time"
)

type DefaultPublicHolidayChecker struct {
	regionName string
}

func (d DefaultPublicHolidayChecker) IsPublicHoliday(_ context.Context, t time.Time) (bool, error) {
	region := mustGetRegion(d.regionName, t.Year())
	for _, v := range region.Feiertage {
		if equalUpToDay(v, t) {
			return true, nil
		}
	}
	return false, nil
}

// NewDefaultRegionHolidayChecker creates a new client for public holidays in that region
// See https://github.com/wlbr/feiertage for a list of valid inputs
func NewDefaultRegionHolidayChecker(regionName string) (*DefaultPublicHolidayChecker, error) {
	res := &DefaultPublicHolidayChecker{regionName: regionName}

	//check that our backend knows the supplied region
	for _, v := range feiertage.GetAllRegions(time.Now().Year(), false) {
		if v.Name == regionName {
			return res, nil
		}
	}
	return nil, fmt.Errorf("region not found ")
}

func getRegion(regionName string, year int) (feiertage.Region, error) {
	for _, v := range feiertage.GetAllRegions(year, false) {
		if v.Name == regionName {
			return v, nil
		}
	}
	return feiertage.Region{}, fmt.Errorf("not found")
}
func mustGetRegion(regionName string, year int) feiertage.Region {
	region, err := getRegion(regionName, year)
	if err != nil {
		panic(err)
	}
	return region
}

func equalUpToDay(t1 feiertage.Feiertag, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}
