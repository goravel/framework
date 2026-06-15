package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageMakeCommandStubsFacadesUsesFullMainImportPath(t *testing.T) {
	stubs := PackageMakeCommandStubs{
		main: "gitlab.com/ijobs.uz/medclinic.uz/api",
		pkg:  "sms",
		root: "packages/sms",
		name: "sms",
	}

	content := stubs.Facades()

	assert.Contains(t, content, `"gitlab.com/ijobs.uz/medclinic.uz/api/packages/sms"`)
	assert.Contains(t, content, `"gitlab.com/ijobs.uz/medclinic.uz/api/packages/sms/contracts"`)
}
