package common

import (
	demoinfo "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	"time"
)

const (
	PhaseFreezetime Phase = iota
	PhaseRegular
	PhasePlanted
	PhaseRestart
	PhaseWarmup
	PhaseHalftime
)

type Point struct {
	X float32
	Y float32
	Z float32
}

type Effect struct {
	Position Point
	Type     demoinfo.EquipmentType
	Lifetime int32
	Team     demoinfo.Team
}

type OverviewState struct {
	IngameTick int
	Players    []Player
	Grenades   []GrenadeProjectile
	Infernos   []Inferno
	Bomb       Bomb
	TeamCT     TeamState
	TeamT      TeamState
	Timer      Timer
}

type Player struct {
	Name               string
	ID                 int
	Team               demoinfo.Team
	Position           Point
	ViewDirectionX     float32
	ViewDirectionY     float32
	FlashDuration      time.Duration
	FlashTimeRemaining time.Duration
	Inventory          []demoinfo.EquipmentType
	ActiveWeapon       demoinfo.EquipmentType
	Health             int16
	Armor              int16
	Money              int16
	Kills              int16
	Deaths             int16
	Assists            int16
	IsAlive            bool
	IsDefusing         bool
	HasHelmet          bool
	HasDefuseKit       bool
	HasBomb            bool
}

type GrenadeProjectile struct {
	Position Point
	Type     demoinfo.EquipmentType
}

type Inferno struct {
	ConvexHull2D []Point
}

type Bomb struct {
	Position       Point
	IsBeingCarried bool
}

type TeamState struct {
	ClanName string
	Score    byte
}

type Timer struct {
	TimeRemaining time.Duration
	Phase         Phase
}

type Phase int

type MapInfo struct {
	AlternateOverview string
	HeightThreshold   float64
}

type Kill struct {
	Killer     string
	KillerTeam demoinfo.Team
	Victim     string
	VictimTeam demoinfo.Team
	Weapon     demoinfo.EquipmentType
	HeadShot   bool
	WallBang   bool
}

type Shoot struct {
	Position       Point
	ViewDirectionX float32
}
