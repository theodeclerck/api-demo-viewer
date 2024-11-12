package internal

import (
	"api-demo-viewer/common"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	demoinfo "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"sort"
	"time"
)

const (
	flashEffectLifetime  int32 = 10
	heEffectLifetime     int32 = 10
	defuseEffectLifetime int32 = 45
	bombEffectLifetime   int32 = 60
	killfeedLifetime     int   = 10
	c4timer              int   = 40
)

type Game struct {
	MatchId primitive.ObjectID
	MapName string
	// MapPZero             common.Point
	HalfStarts           []int
	RoundStarts          []int
	Effects              map[int][]common.Effect
	FrameRate            int
	FrameTime            time.Duration
	States               []common.OverviewState
	SmokeEffectLifetime  int32
	KillFeed             map[int][]common.Kill
	Shots                map[int][]common.Shoot
	currentPhase         common.Phase
	latestTimerEventTime time.Duration
	currentFrame         int
}

func NewGame(demoName string, matchID primitive.ObjectID, taskID primitive.ObjectID) error {
	f, err := os.Open("files/" + demoName)
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}

	parser := dem.NewParser(f)
	defer func(p dem.Parser) {
		err := p.Close()
		if err != nil {
			log.Panic("failed to close parser: ", err)
		}
	}(parser)

	header, err := parser.ParseHeader()
	if err != nil {
		log.Panic("failed to parse header: ", err)
	}

	game := &Game{
		MatchId:      matchID,
		HalfStarts:   make([]int, 0),
		RoundStarts:  make([]int, 0),
		Effects:      make(map[int][]common.Effect),
		KillFeed:     make(map[int][]common.Kill),
		Shots:        make(map[int][]common.Shoot),
		currentPhase: common.PhaseRegular,
	}

	game.MapName = header.MapName

	registerEventHandlers(parser, game)

	game.States = parseGameStates(parser, game)

	if err := CreateGame(game); err != nil {
		if err := UpdateTask(taskID, "failed"); err != nil {
			log.Printf("Failed to update task status to failed: %v", err)
		}
		log.Panic("failed to create game: ", err)
	}

	if err := UpdateTask(taskID, "completed"); err != nil {
		log.Printf("Failed to update task status to completed: %v", err)
	}

	return nil
}

func parseGameStates(parser dem.Parser, game *Game) []common.OverviewState {
	playbackFrames := parser.Header().PlaybackFrames
	header := parser.Header()
	game.FrameRate = int(header.FrameRate())
	game.FrameTime = header.FrameTime()
	game.SmokeEffectLifetime = int32(18 * game.FrameRate)

	states := make([]common.OverviewState, 0, playbackFrames)

	for ok, err := parser.ParseNextFrame(); ok && err == nil; ok, err = parser.ParseNextFrame() {
		game.currentFrame = parser.CurrentFrame()
		if game.currentFrame == 0 {
			continue
		}
		gameState := parser.GameState()
		players := make([]common.Player, 0, 10)

		for _, p := range gameState.Participants().Playing() {
			var hasBomb bool
			inventory := make([]demoinfo.EquipmentType, 0)
			for _, weapon := range p.Weapons() {
				if weapon.Type == demoinfo.EqBomb {
					hasBomb = true
				}
				if isWeaponOrGrenade(weapon.Type) {
					inventory = append(inventory, weapon.Type)
				}
				inventory = append(inventory, weapon.Type)
			}

			sort.Slice(inventory, func(i, j int) bool { return inventory[i] < inventory[j] })

			// TODO: Add info about map vertical

			var activeWeapon demoinfo.EquipmentType
			if p.ActiveWeapon() == nil {
				activeWeapon = demoinfo.EqUnknown
			} else {
				activeWeapon = p.ActiveWeapon().Type
			}

			player := common.Player{
				Name:               p.Name,
				ID:                 p.UserID,
				Team:               p.Team,
				Position:           common.Point{X: float32(p.Position().X), Y: float32(p.Position().Y), Z: float32(p.Position().Z)},
				ViewDirectionX:     p.ViewDirectionX(),
				ViewDirectionY:     p.ViewDirectionY(),
				FlashDuration:      p.FlashDurationTime(),
				FlashTimeRemaining: p.FlashDurationTimeRemaining(),
				Inventory:          inventory,
				ActiveWeapon:       activeWeapon,
				Health:             int16(p.Health()),
				Armor:              int16(p.Armor()),
				Money:              int16(p.Money()),
				Kills:              int16(p.Kills()),
				Deaths:             int16(p.Deaths()),
				Assists:            int16(p.Assists()),
				IsAlive:            p.IsAlive(),
				IsDefusing:         p.IsDefusing,
				HasHelmet:          p.HasHelmet(),
				HasDefuseKit:       p.HasDefuseKit(),
				HasBomb:            hasBomb,
			}
			players = append(players, player)
		}
		sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })

		grenades := make([]common.GrenadeProjectile, 0)
		for _, grenade := range gameState.GrenadeProjectiles() {
			g := common.GrenadeProjectile{
				Position: common.Point{X: float32(grenade.Position().X), Y: float32(grenade.Position().Y), Z: float32(grenade.Position().Z)},
				Type:     grenade.WeaponInstance.Type,
			}
			grenades = append(grenades, g)
		}

		infernos := make([]common.Inferno, 0)
		for _, inferno := range gameState.Infernos() {
			r2points := inferno.Fires().Active().ConvexHull2D()
			points := make([]common.Point, 0)
			for _, p := range r2points {
				points = append(points, common.Point{X: float32(p.X), Y: float32(p.Y)})
			}
			i := common.Inferno{ConvexHull2D: points}
			infernos = append(infernos, i)
		}

		var isBeingCarried bool
		if gameState.Bomb().Carrier != nil {
			isBeingCarried = true
		} else {
			isBeingCarried = false
		}

		bomb := common.Bomb{
			Position:       common.Point{X: float32(gameState.Bomb().Position().X), Y: float32(gameState.Bomb().Position().Y), Z: float32(gameState.Bomb().Position().Z)},
			IsBeingCarried: isBeingCarried,
		}

		cts := common.TeamState{
			Score:    byte(gameState.TeamCounterTerrorists().Score()),
			ClanName: gameState.TeamCounterTerrorists().ClanName(),
		}

		ts := common.TeamState{
			Score:    byte(gameState.TeamTerrorists().Score()),
			ClanName: gameState.TeamTerrorists().ClanName(),
		}

		var timer common.Timer
		if gameState.IsWarmupPeriod() {
			timer = common.Timer{TimeRemaining: 0, Phase: common.PhaseWarmup}
		} else {
			switch game.currentPhase {
			case common.PhaseFreezetime:
				timer = common.Timer{TimeRemaining: 0, Phase: common.PhaseFreezetime} // TODO: Add freezetime duration

			case common.PhaseRegular:
				timer = common.Timer{TimeRemaining: 0, Phase: common.PhaseRegular} // TODO: Add round time

			case common.PhasePlanted:
				timer = common.Timer{TimeRemaining: time.Duration(c4timer)*time.Second - (parser.CurrentTime() - game.latestTimerEventTime), Phase: common.PhasePlanted}

			case common.PhaseRestart:
				timer = common.Timer{TimeRemaining: 0, Phase: common.PhaseRestart} // TODO: Add restart duration

			case common.PhaseHalftime:
				timer = common.Timer{TimeRemaining: 0, Phase: common.PhaseHalftime} // TODO: Add halftime duration

			default:
				panic("unhandled default case")
			}
		}

		state := common.OverviewState{
			IngameTick: parser.GameState().IngameTick(),
			Players:    players,
			Grenades:   grenades,
			Infernos:   infernos,
			Bomb:       bomb,
			TeamCT:     cts,
			TeamT:      ts,
			Timer:      timer,
		}
		states = append(states, state)
	}
	return states
}

func registerEventHandlers(parser dem.Parser, game *Game) {

}

func isWeaponOrGrenade(e demoinfo.EquipmentType) bool {
	return e.Class() == demoinfo.EqClassSMG ||
		e.Class() == demoinfo.EqClassHeavy ||
		e.Class() == demoinfo.EqClassRifle ||
		e.Class() == demoinfo.EqClassPistols ||
		e.Class() == demoinfo.EqClassGrenade
}
