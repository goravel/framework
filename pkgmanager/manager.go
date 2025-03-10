package pkgmanager

import (
	"github.com/goravel/framework/contracts/pkgmanager"
	"github.com/goravel/framework/support/color"
)

type Manager struct {
	ContinueOnError bool
	Module          string
	OnInstall       []pkgmanager.FileModifier
	OnUninstall     []pkgmanager.FileModifier
}

func (m *Manager) Install(dir string) error {
	for i := range m.OnInstall {
		if err := m.OnInstall[i].Apply(dir); err != nil {
			if m.ContinueOnError {
				color.Warningln(err)
				continue
			}

			return err
		}
	}

	return nil
}

func (m *Manager) Uninstall(dir string) error {
	for i := range m.OnUninstall {
		if err := m.OnUninstall[i].Apply(dir); err != nil {
			if m.ContinueOnError {
				color.Warningln(err)
				continue
			}

			return err
		}
	}

	return nil
}
