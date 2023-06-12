package database

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
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

		if contains(s.Called, name) {
			continue
		}

		if !silent && s.Command != nil {
			color.Yellowf("RUNNING: %s\n", name)
		}

		startTime := time.Now()

		err := s.Invoke(seeder, parameters)
		if err != nil {
			log.Println("error", err)
			return err
		}

		if !silent && s.Command != nil {
			runTime := time.Since(startTime).Milliseconds()
			color.Printf("%s %d ms DONE\n", name, runTime)
		}
		s.Called = append(s.Called, name)
	}
	return nil
}
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (s *Seeder) CallWith(class interface{}, parameters []interface{}) error {
	return s.Call(class, false, parameters)
}

func (s *Seeder) CallSilent(class interface{}, parameters []interface{}) error {
	return s.Call(class, true, parameters)
}

func (s *Seeder) CallOnce(class interface{}, silent bool, parameters []interface{}) error {
	classType := reflect.TypeOf(class)
	classTypeName := classType.String()
	classPointerTypeName := "*" + classTypeName

	for _, called := range s.Called {
		if called == classTypeName || called == classPointerTypeName {
			return nil
		}
	}

	return s.Call(class, silent, parameters)
}

func (s *Seeder) Resolve(class interface{}) interface{} {
	instanceType := reflect.TypeOf(class)

	var instance interface{}
	var instanceValue reflect.Value
	if s.Container != nil {
		resolvedInstance, err := s.Container.Make(instanceType.String())
		if err != nil {
			// Handle the error if necessary
			return nil
		}

		instanceValue = reflect.ValueOf(resolvedInstance)
		instance = instanceValue.Interface()
		// Set the container and command on the instance (assuming it has the necessary methods)
		setContainerMethod := instanceValue.MethodByName("SetContainer")
		if setContainerMethod.IsValid() {
			setContainerMethod.Call([]reflect.Value{reflect.ValueOf(s.Container)})
		}
	} else {
		// Create a new instance of the class using reflection
		instanceValue = reflect.New(instanceType)
		instance = instanceValue.Interface()
	}

	if s.Command != nil {
		setCommandMethod := instanceValue.MethodByName("SetCommand")
		if setCommandMethod.IsValid() {
			setCommandMethod.Call([]reflect.Value{reflect.ValueOf(s.Command)})
		}
	}
	return instance
}

func (s *Seeder) SetContainer(container foundation.Container) {
	s.Container = container
}

func (s *Seeder) SetCommand(command console.Context) {
	s.Command = command
}

func (s *Seeder) Invoke(seeder interface{}, parameters []interface{}) error {
	runMethod := reflect.ValueOf(seeder).MethodByName("Run")
	// if s.Container == nil {
	// 	log.Println("running", runMethod)
	// 	return fmt.Errorf("container not set")
	// }

	if !runMethod.IsValid() {
		return fmt.Errorf("method [Run] missing from %T", seeder)
	}

	callback := func() error {
		// Invoke the Run method if it exists
		if runMethod.IsValid() {
			returnValue := runMethod.Call([]reflect.Value{reflect.ValueOf(s.Command)})
			if len(returnValue) > 0 && !returnValue[0].IsNil() {
				return returnValue[0].Interface().(error)
			}
			return nil
		}
		return fmt.Errorf("method [Run] missing from %T", seeder)
	}

	return callback()
}
