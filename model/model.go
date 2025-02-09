// Package model provides types for deserializing
// the response from the DZSA API.
package model

import (
	"net"
	"strconv"
)

// QueryResponse is the response from the DZSA API
// https://dayzsalauncher.com/api/v1/query/:ip:port
// endpoint.
type QueryResponse struct {
	Result Result `json:"result"`
	Status int    `json:"status"`
}

// Endpoint represents the endpoint of a dayz server.
type Endpoint struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// String returns the string representation of an endpoint.
func (e Endpoint) String() string {
	return net.JoinHostPort(e.IP, strconv.Itoa(e.Port))
}

// Mods represents a mod of a dayz server.
type Mods struct {
	Name            string `json:"name"`
	SteamWorkshopID int    `json:"steamWorkshopId"`
}

// Result represents the result of a dayz server.
type Result struct {
	BattlEye         bool     `json:"battlEye"`
	Endpoint         Endpoint `json:"endpoint"`
	Environment      string   `json:"environment"`
	FirstPersonOnly  bool     `json:"firstPersonOnly"`
	Folder           string   `json:"folder"`
	Game             string   `json:"game"`
	GamePort         int      `json:"gamePort"`
	Map              string   `json:"map"`
	MaxPlayers       int      `json:"maxPlayers"`
	Mission          string   `json:"mission"`
	Mods             []Mods   `json:"mods"`
	Name             string   `json:"name"`
	NameOverride     bool     `json:"nameOverride"`
	Password         bool     `json:"password"`
	Players          int      `json:"players"`
	Profile          bool     `json:"profile"`
	Shard            string   `json:"shard"`
	Sponsor          bool     `json:"sponsor"`
	Time             string   `json:"time"`
	TimeAcceleration int      `json:"timeAcceleration"`
	Vac              bool     `json:"vac"`
	Version          string   `json:"version"`
}
