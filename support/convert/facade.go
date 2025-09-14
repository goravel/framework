package convert

import "github.com/goravel/framework/support/str"

func BindingToFacade(binding string) string {
	return str.Of(binding).After("goravel.").Studly().WhenIs("Db", func(s *str.String) *str.String {
		return s.Upper()
	}).String()
}

func FacadeToBinding(facade string) string {
	return "goravel." + str.Of(facade).Snake().String()
}
