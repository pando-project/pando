package http

func (a *API) registerSwagger() {
	a.router.StaticFile("/swagger/specs", "./docs/swagger.yml")
}
