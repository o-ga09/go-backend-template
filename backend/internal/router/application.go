package router

// SetupApplicationRoute はアプリケーション固有のAPIルーティングを登録する。
func (r *route) SetupApplicationRoute() {
	users := r.rooAPI.Group("/users")
	users.POST("", r.user.Create)
	users.GET("/:id", r.user.GetByID)

	auth := r.rooAPI.Group("/auth")
	auth.GET("/user", r.auth.CurrentUser)
	auth.POST("/logout", r.auth.Logout)
}
