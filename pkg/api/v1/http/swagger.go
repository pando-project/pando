package http

func (a *API) registerSwagger() {
	a.router.StaticFile("/swagger/specs", "/opt/swagger.yml")
}
