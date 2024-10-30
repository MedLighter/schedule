package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"myxxa/scheduler/internal/db"
	"net/http"
	"time"
)

type RequestBody struct {
	Nrec     int    `json:"Nrec"`
	WeekType int8   `json:"WeekType"`
	WeekDays string `json:"WeekDays"`
}

func InterceptorSchedule(cache *db.Cache) {
	checkScheduleCache(cache)
}

func checkScheduleCache(cache *db.Cache) {
	key := fmt.Sprintf("schedule")
	_, completed := cache.GetUnmarshal(key)
	if completed {
		return
	}
	schedule, err := RequestScheduleOnWeek()
	if err != nil {
		log.Fatal(err)
	}
	cache.SetMarshal(key, schedule, time.Hour*12)
}

func RequestScheduleOnWeek() ([]map[string]interface{}, error) {
	url := "https://abiturient-api.vlsu.ru/api/student/GetGroupSchedule"

	body := RequestBody{
		Nrec:     281474976724434,
		WeekType: 0,
		WeekDays: "1,2,3,4,5,6",
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Ошибка при маршализации:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://student.vlsu.ru")
	req.Header.Set("Referer", "https://student.vlsu.ru/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var responseBody []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			fmt.Println("Ошибка при декодировании ответа:", err)
			return nil, err
		}
		return responseBody, nil
	} else {
		fmt.Printf("Ошибка: %sn", resp.Status)
		return nil, nil
	}
}
