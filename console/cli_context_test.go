package console

import (
	"testing"
)

func TestAsk(_ *testing.T) {
	/*
			ctx := &CliContext{}
			// single line input text
			question := "How are you feeling today?"
			answer, err := ctx.Ask(question, console.AskOption{
				Default:     "Good",
				Description: "Please enter your feeling",
				Limit:       10,
				Placeholder: "Good",
				Prompt:      ">",
				Validate: func(s string) error {
					if s == "" {
						return fmt.Errorf("please enter your feeling")
					}
					return nil
				},
			})
			if err != nil {
				ctx.Error(err.Error())
				return
			}
			ctx.Info(fmt.Sprintf("You said: %s", answer))

			// multiple lines input text
			question = "tell me about yourself"
			answer, err = ctx.Ask(question, console.AskOption{
				Default:     "I am a software engineer",
				Description: "Please enter your bio",
		        Multiple:    true,
				Lines:       5,
				Placeholder: "Bio",
				Validate: func(s string) error {
					if s == "" {
						return fmt.Errorf("please enter your bio")
					}
					return nil
				},
			})

			if err != nil {
				ctx.Error(err.Error())
				return
			}

			ctx.Info(fmt.Sprintf("You said: %s", answer))
	*/
}

func TestCreateProgressBar(_ *testing.T) {
	/*
		ctx := &CliContext{}
		bar := ctx.CreateProgressBar(100)
		err := bar.Start()
		if err != nil {
			ctx.Error(err.Error())
			return
		}

		for i := 1; i < 100; i++ {
			// performTask()
			if i%2 == 0 {
				bar.Advance(2)
			} else {
				bar.Advance()
			}
			time.Sleep(time.Millisecond * 50)
		}

		err = bar.Finish()
		if err != nil {
			ctx.Error(err.Error())
			return
		}
	*/
}

func TestChoice(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "What is your favorite programming language?"
		options := []console.Choice{
			{Key: "go", Value: "Go"},
			{Key: "php", Value: "PHP"},
			{Key: "python", Value: "Python"},
			{Key: "cpp", Value: "C++", Selected: true},
		}

		answer, err := ctx.Choice(question, options, console.ChoiceOption{
			Default:     "cpp",
			Description: "Please select your favorite programming language",
			Validate: func(s string) error {
				if s == "Python" {
					return fmt.Errorf("you can't have Python as your favorite programming language")
				}

				return nil
			},
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info(fmt.Sprintf("You selected: %s", answer))
	*/
}

func TestConfirm(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "Are you sure you want to continue?"
		confirmation := false
		confirmed, err := ctx.Confirm(question, console.ConfirmOption{
			Affirmative: "Hell Yeah",
			Default:     &confirmation,
			Description: "Please confirm to proceed",
			Negative:    "Nah",
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		if confirmed {
			ctx.Info("You confirmed to proceed.")
		} else {
			ctx.Info("You declined to proceed.")
		}
	*/
}

func TestMultiSelect(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "What are your favorite colors?"
		options := []console.Choice{
			{Key: "red", Value: "Red"},
			{Key: "blue", Value: "Blue"},
			{Key: "green", Value: "Green"},
			{Key: "yellow", Value: "Yellow", Selected: true},
			{Key: "purple", Value: "Purple"},
		}
		filterable := true
		answers, err := ctx.MultiSelect(question, options, console.MultiSelectOption{
			Default:     []string{"yellow"},
			Description: "Please select your favorite colors",
			Filterable:  &filterable,
			Limit:       3,
			Validate: func(s []string) error {
				if len(s) == 0 {
					return fmt.Errorf("please select at least one color")
				}
				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info(fmt.Sprintf("You selected: %v", answers))
	*/
}

func TestNewLine(_ *testing.T) {
	/*
		ctx := &CliContext{}
		ctx.NewLine()
		ctx.NewLine(3)
	*/
}

func TestQuestion(_ *testing.T) {
	/*
		ctx := &CliContext{}
		ctx.Question("What is your name?")
	*/
}

func TestSecret(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "What is your password?"
		password, err := ctx.Secret(question, console.SecretOption{
			Default:     "password",
			Description: "Please enter your password",
			Limit:       15,
			Placeholder: "password",
			Validate: func(s string) error {
				if len(s) < 8 {
					return fmt.Errorf("password must be at least 8 characters")
				}

				return nil
			},
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info(fmt.Sprintf("You entered: %s", password))
	*/
}

func TestSpinner(_ *testing.T) {
	/*
		ctx := &CliContext{}
		err := ctx.Spinner("Loading...", console.SpinnerOption{
			Action: func() error {
				// when to stop the spinner
				time.Sleep(2 * time.Second)
				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info("Task completed successfully.")
	*/
}

func TestWarn(_ *testing.T) {
	/*
		ctx := &CliContext{}
		ctx.Warn("This is a warning message.")
	*/
}

func TestWithProgressBar(_ *testing.T) {
	/*
		ctx := &CliContext{}
		items := []any{"item1", "item2", "item3"}
		_, err := ctx.WithProgressBar(items, func(item any) error {
			// performTask(item)
			return nil
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info("Task completed successfully.")
	*/
}
