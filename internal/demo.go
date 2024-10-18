package internal

import (
	"fmt"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
)

func OpenDemo(demoName string, matchID primitive.ObjectID, taskID primitive.ObjectID) error {
	f, err := os.Open("files/" + demoName)
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Panic("failed to close demo file: ", err)
		}
	}()

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
		//fmt.Printf("%s is %s", e.Killer, e.Killer.Position())
		fmt.Printf("%s <%v%s%s> %s\n", e.Killer.Name, e.Weapon, hs, wallBang, e.Victim.Name)
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		fmt.Printf("Round %d ended\n", round)
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		err := UpdateTask(taskID, "failed")
		if err != nil {
			return err
		}
		log.Fatalf("failed to parse demo: %v", err)
	}

	if err := UpdateTask(taskID, "completed"); err != nil {
		log.Printf("Failed to update task status to completed: %v", err)
	}

	defer func() {
		err := os.Remove("files/" + demoName)
		if err != nil {
			log.Printf("failed to remove demo file: %v", err)
		}
	}()
	return nil
}
