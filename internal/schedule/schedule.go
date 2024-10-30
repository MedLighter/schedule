package schedule

import (
	"fmt"
	"time"
)

type Schedule struct {
	Numerator   Week
	Denominator Week
}

type Week struct {
	Days []Day
}

type Day struct {
	Name    string
	Classes []Class
}

type Class struct {
	Id          int
	Description string
}

type WeekType int8

const (
	WeekTypeNumerator   WeekType = 0
	WeekTypeDenominator WeekType = 1
)

func GetWeekType() WeekType {
	_, week := time.Now().ISOWeek()
	weekType := WeekTypeDenominator
	if week%2 == 0 {
		weekType = WeekTypeNumerator
	}
	return weekType
}

func MapToSchedule(schedule []map[string]interface{}) Schedule {
	weekNum := getWeek(schedule, WeekTypeNumerator)
	weenDenom := getWeek(schedule, WeekTypeDenominator)
	return Schedule{
		Numerator:   weekNum,
		Denominator: weenDenom,
	}
}

func getWeek(schedule []map[string]interface{}, weekType WeekType) Week {
	var key string
	if weekType == WeekTypeNumerator {
		key = "n"
	}
	if weekType == WeekTypeDenominator {
		key = "z"
	}
	var days []Day = []Day{}
	for i := 0; i < 6; i++ {
		name := schedule[i]["name"].(string)
		var classes []Class = []Class{}
		for j := 1; j < 7; j++ {
			s := schedule[i][fmt.Sprintf("%v%d", key, j)]
			if s != nil {
				classes = append(classes, Class{
					Id:          j - 1,
					Description: s.(string),
				})
			}
		}
		days = append(days, Day{
			Name:    name,
			Classes: classes,
		})
	}
	return Week{
		Days: days,
	}
}
