package hue

import "time"

type Light struct {
	Name    string
	UID     string `json:"uniqueid"`
	ID      string
	Type    string
	ModelID string `json:"modelid"`
	State   struct {
		lightSettings
		Reachable bool
		alert     string
		colorMode string `json:"colormode"`
	}
}

type lightSettings struct {
	On     bool
	Bri    uint8
	Hue    uint16
	Sat    uint8
	XY     []float32
	CT     uint16
	Effect string
}

type Scene struct {
	ID             string
	Name           string
	Lights         []string
	TransitionTime int16 `json:"transitiontime"`
	lightSettings
}

type Group struct {
	ID     string
	Name   string
	Type   string
	Lights []string
	Action lightSettings
}

type GroupAction struct {
	lightSettings
	Scene string
}

type Schedule struct {
	ID          string
	Name        string
	Description string
	Command     struct {
		Address string
		Body    interface{}
		Method  string
	}
	Status    string
	Created   time.Time
	LocalTime time.Time `json:"localtime"`
}
