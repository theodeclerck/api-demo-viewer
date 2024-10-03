package internal

import (
	"fmt"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
)

func OpenDemo(demoName string, matchID primitive.ObjectID, taskID primitive.ObjectID) {
	f, err := os.Open("files/" + demoName)
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}
	/*defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic("failed to close demo file: ", err)
		}
	}(f)*/

	p := dem.NewParser(f)
	defer func(p dem.Parser) {
		err := p.Close()
		if err != nil {
			log.Panic("failed to close parser: ", err)
		}
	}(p)
	round := 0
	p.RegisterEventHandler(func(e events.RoundStart) {
		round++
		fmt.Printf("Round %d started\n", round)
	})

	fmt.Print(p.GameState().Participants().Playing())

	p.RegisterEventHandler(func(e events.Kill) {
		var hs string
		if e.IsHeadshot {
			hs = " (HS)"
		}
		var wallBang string
		if e.PenetratedObjects > 0 {
			wallBang = " (WB)"
		}
		fmt.Printf("%s is %s", e.Killer, e.Killer.Position())
		fmt.Printf("%s <%v%s%s> %s\n", e.Killer, e.Weapon, hs, wallBang, e.Victim)
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		fmt.Printf("Round %d ended\n", round)
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		err := UpdateTask(taskID, "failed")
		if err != nil {
			return
		}
		log.Panic("failed to parse demo: ", err)
	}

	err = UpdateTask(taskID, "completed")
	if err != nil {
		return
	}

	// TODO: remove file after parsing
	/*err2 := os.Remove("files/" + demoName)
	if err2 != nil {
		print(err2)
		return
	}*/
}
