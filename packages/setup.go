package packages

import (
	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support/color"
)

type Setup struct {
	Force       bool
	Module      string
	OnInstall   []packages.FileModifier
	OnUninstall []packages.FileModifier
}

func (r *Setup) Install() error {
	for i := range r.OnInstall {
		if err := r.OnInstall[i].Apply(); err != nil {
			if r.Force {
				color.Warningln(err)
				continue
			}

			return err
		}
	}

	return nil
}

func (r *Setup) Uninstall() error {
	for i := range r.OnUninstall {
		if err := r.OnUninstall[i].Apply(); err != nil {
			if r.Force {
				color.Warningln(err)
				continue
			}

			return err
		}
	}

	return nil
}
