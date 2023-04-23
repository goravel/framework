package facades

import (
	"errors"

	"github.com/goravel/framework/contracts/sms"
)

var Sms = &SmsImpl{}

type SmsImpl struct {
	sms.Sms
}

func (receiver *SmsImpl) GetFacadeAccessor() string {
	return "sms"
}

func (receiver *SmsImpl) ResolveFacadeInstance(instance any) error {
	if sms, ok := instance.(sms.Sms); ok {
		receiver.Sms = sms

		return nil
	}

	return errors.New("not implement github.com/goravel/framework/contracts/sms")
}
