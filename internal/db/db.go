package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"myxxa/scheduler/internal/schedule"
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	items             map[string]Item
}

type Item struct {
	Value      interface{}
	Created    time.Time
	Expiration int64
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {

	// инициализируем карту(map) в паре ключ(string)/значение(Item)
	items := make(map[string]Item)

	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	// Если интервал очистки больше 0, запускаем GC (удаление устаревших элементов)
	if cleanupInterval > 0 {
		cache.StartGC() // данный метод рассматривается ниже
	}

	return &cache
}

func (c *Cache) Get(key string) (interface{}, bool) {

	c.RLock()

	defer c.RUnlock()

	item, found := c.items[key]

	if !found {
		return nil, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}

	return item.Value, true
}

func GetDaySchedule(day int, weekType schedule.WeekType, cache *Cache) string {
	var text string = "\n"
	req, complited := cache.GetSchedule()
	fmt.Print(req)
	if !complited {
		log.Fatal(errors.New("Ключ не найден").Error())
	}
	// weekType := schedule.GetWeekType()

	var updateText = func(txt *string, classes []schedule.Class) {
		for class := range classes {
			if classes[class].Description == "" {
				continue
			}
			*txt += getTextNumberClass(class)
			*txt += fmt.Sprintf("\n%v\n\n", classes[class].Description)
		}
	}

	switch weekType {
	case schedule.WeekTypeNumerator:
		updateText(&text, req.Numerator.Days[day].Classes)
		break
	case schedule.WeekTypeDenominator:
		updateText(&text, req.Denominator.Days[day].Classes)
		break
	}
	return text
}

func getTextNumberClass(id int) string {
	switch id {
	case 0:
		return "*-* 1 пара 08:30 - 10:00"

	case 1:
		return "*-* 2 пара 10:20 - 11:50"

	case 2:
		return "*-* 3 пара 12:10 - 13:40"

	case 3:
		return "*-* 4 пара 14:00 - 15:30"

	case 4:
		return "*-* 5 пара 15:50 - 17:20"

	case 5:
		return "*-* 6 пара 17:40 - 19:10"
	default:
		return ""
	}
}

func (c *Cache) GetSchedule() (schedule.Schedule, bool) {
	response, completed := c.GetUnmarshal("schedule")
	if !completed {
		return schedule.Schedule{}, completed
	}
	res := schedule.MapToSchedule(response)
	return res, true
}

func (c *Cache) GetUnmarshal(key string) ([]map[string]interface{}, bool) {
	response, completed := c.Get(key)
	if !completed {
		return nil, completed
	}
	var responseUnmarsh []map[string]interface{}
	err := json.Unmarshal(response.([]byte), &responseUnmarsh)
	if err != nil {
		log.Fatal(err)
	}
	return responseUnmarsh, true
}

func (c *Cache) SetMarshal(key string, value interface{}, duration time.Duration) {
	marsh, err := json.Marshal(value)
	if err != nil {
		log.Fatal(err)
	}
	c.Set(key, marsh, time.Minute*30)
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()
	defer c.Unlock()

	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}
}

func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {

	for {
		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

// expiredKeys возвращает список "просроченных" ключей
func (c *Cache) expiredKeys() (keys []string) {

	c.RLock()

	defer c.RUnlock()

	for k, i := range c.items {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}
	return
}

// clearItems удаляет ключи из переданного списка, в нашем случае "просроченные"
func (c *Cache) clearItems(keys []string) {

	c.Lock()

	defer c.Unlock()

	for _, k := range keys {
		delete(c.items, k)
	}
}

func (c *Cache) Delete(key string) error {

	c.Lock()

	defer c.Unlock()

	if _, found := c.items[key]; !found {
		return errors.New("Key not found")
	}

	delete(c.items, key)

	return nil
}
