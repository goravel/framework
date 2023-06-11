package database

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/foundation"
)


type Seeder struct {
	Container foundation.Container
	Command   console.Context
	Called    []string
}

func (s *Seeder) Call(class interface{}, silent bool, parameters []interface{}) error {
	classes, ok := class.([]interface{})
	if !ok {
		classes = []interface{}{class}
	}

	for _, class := range classes {
		seeder := s.Resolve(class)

		name := fmt.Sprintf("%T", seeder)

		if !silent && s.Command != nil {
			fmt.Printf("%s <fg=yellow;options=bold>RUNNING</>\n", name)
		}

		startTime := time.Now()
		log.Println(seeder)
		err := s.Invoke(seeder, parameters)
		if err != nil {
			return err
		}

		if !silent && s.Command != nil {
			runTime := time.Since(startTime).Milliseconds()
			fmt.Printf("%s <fg=gray>%d ms</> <fg=green;options=bold>DONE</>\n\n", name, runTime)
		}

		s.Called = append(s.Called, name)
	}

	return nil
}

func (s *Seeder) CallWith(class interface{}, parameters []interface{}) error {
	return s.Call(class, false, parameters)
}

func (s *Seeder) CallSilent(class interface{}, parameters []interface{}) error {
	return s.Call(class, true, parameters)
}

func (s *Seeder) CallOnce(class interface{}, silent bool, parameters []interface{}) error {
	for _, called := range s.Called {
		if called == fmt.Sprintf("%T", class) {
			return nil
		}
	}

	return s.Call(class, silent, parameters)
}

func (s *Seeder) Resolve(class interface{}) database.Seeder {
	if s.Container != nil {
		instance, err := s.Container.Make(class)

		if err != nil {
			// Handle the error if necessary
			return nil
		}

		// Check if the resolved instance implements the Seeder interface
		seeder, ok := instance.(database.Seeder)
		if !ok {
			log.Println("database.Seeder", instance, ok)
			// Handle the case where the resolved instance does not implement the Seeder interface
			return nil
		}
		log.Println("database.Seeder", instance)
		// Set the container and command on the seeder instance
		seeder.SetContainer(s.Container)
		seeder.SetCommand(s.Command)

		return seeder
	}

	// Handle the case where the container is nil
	log.Println("Container is nil")
	return nil
}

// func (s *Seeder) Resolve(class interface{}) database.Seeder {
// 	var instance database.Seeder

// 	if s.Container != nil {
// 		resolvedInstance, _ := s.Container.Make(class)
// 		instance, _ = resolvedInstance.(database.Seeder)
// 		if instance == nil {
// 			// Handle the case where the resolved instance does not implement the database.Seeder
// 			return nil
// 		}
// 		instance.SetContainer(s.Container)
// 	} else {
// 		instance, _ = class.(database.Seeder)
// 	}

// 	if s.Command != nil {
// 		instance.SetCommand(s.Command)
// 	}

// 	return instance
// }

func (s *Seeder) SetContainer(container foundation.Container) {
	s.Container = container
}

func (s *Seeder) SetCommand(command console.Context) {
	s.Command = command
}

func (s *Seeder) Invoke(seeder database.Seeder, parameters []interface{}) error {
	if s.Container == nil {
		return fmt.Errorf("container not set")
	}

	if !s.methodExists("Run", seeder) {
		return fmt.Errorf("method [Run] missing from %T", s)
	}

	callback := func() error {
		if s.methodExists("Run", seeder) {
			log.Println("Run method")
			return seeder.Run(s.Command)
		}
		return fmt.Errorf("method [Run] missing from %T", s)
	}

	return callback()
}

func (s *Seeder) methodExists(name string, seeder database.Seeder) bool {
	_, found := reflect.TypeOf(seeder).MethodByName(name)
	return found
}
