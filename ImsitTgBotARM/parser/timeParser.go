package parser

import (
	"time"
)

func NowTime() (t string, d time.Weekday, week string, err error) {
	// Получаем текущее время в UTC
	now := time.Now().UTC()

	//
	// Добавляем смещение для Московской временной зоны (UTC +3)
	loc := time.FixedZone("Moscow Time", 3*60*60) // смещение 3 часа (в секундах)
	now = now.In(loc)

	moscowTime := now.Format("15:04") // Часы и минуты в формате "HH:MM"
	d = now.Weekday()                 // День недели
	_, w := now.ISOWeek()             // Номер недели

	// Проверка времени по интервалам
	if moscowTime >= "08:00" && moscowTime < "09:30" {
		t = "08:00 — 09:30"
	} else if moscowTime >= "09:30" && moscowTime < "11:10" {
		t = "09:40 — 11:10"
	} else if moscowTime >= "11:10" && moscowTime < "13:00" {
		t = "11:30 — 13:00"
	} else if moscowTime >= "13:00" && moscowTime < "14:40" {
		t = "13:10 — 14:40"
	} else if moscowTime >= "14:40" && moscowTime < "16:20" {
		t = "14:50 — 16:20"
	} else if moscowTime >= "16:20" && moscowTime < "18:00" {
		t = "16:30 — 18:00"
	} else if moscowTime >= "18:00" && moscowTime < "19:40" {
		t = "18:10 - 19:40"
	}

	if w%2 == 1 { // чет нечет неделя
		week = "2week"
	} else {
		week = "1week"
	}

	return t, d, week, err
}
