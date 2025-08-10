package template

import (
	"fmt"
	"sync"

	"github.com/goravel/framework/contracts/config"
	contractsmail "github.com/goravel/framework/contracts/mail"
)

var engines sync.Map

// Get retrieves a cached mail template engine, creating it if it doesn't exist.
// This function is safe for concurrent use.
func Get(config config.Config) (contractsmail.Template, error) {
	driver := config.GetString("mail.template.driver", "default")

	if cached, ok := engines.Load(driver); ok {
		return cached.(contractsmail.Template), nil
	}

	engine, err := createEngine(config, driver)
	if err != nil {
		return nil, err
	}

	// Atomically load an existing engine or store the new one.
	// This prevents a race condition where two goroutines might create the same engine.
	// The one that gets to LoadOrStore first wins, and the other's created engine is discarded.
	actual, _ := engines.LoadOrStore(driver, engine)

	return actual.(contractsmail.Template), nil
}

func createEngine(config config.Config, driver string) (contractsmail.Template, error) {
	switch driver {
	case "default":
		viewsPath := config.GetString("mail.template.views_path", "resources/views/mail")
		return NewHtml(viewsPath), nil
	default:
		key := fmt.Sprintf("mail.template.drivers.%s.engine", driver)
		engineConfig := config.Get(key)

		switch v := engineConfig.(type) {
		case contractsmail.Template:
			return v, nil
		case func() (contractsmail.Template, error):
			engine, err := v()
			if err != nil {
				return nil, fmt.Errorf("factory for template engine '%s' failed: %w", driver, err)
			}
			return engine, nil
		default:
			return nil, fmt.Errorf("unsupported or misconfigured template engine: %s", driver)
		}
	}
}
