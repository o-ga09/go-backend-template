package router

// SetupApplicationRoute はアプリケーション固有のAPIルーティングを登録する。
func (r *route) SetupApplicationRoute() {
	users := r.rooAPI.Group("/users")
	users.POST("", r.user.Create)
	users.GET("/:id", r.user.GetByID)

	auth := r.rooAPI.Group("/auth")
	auth.GET("/user", r.auth.CurrentUser)
	auth.POST("/logout", r.auth.Logout)

	products := r.rooAPI.Group("/products")
	products.GET("", r.product.List)
	products.GET("/:id", r.product.GetByID)

	cart := r.rooAPI.Group("/cart")
	cart.GET("", r.cart.Get)
	cart.POST("", r.cart.AddItem)
	cart.PUT("", r.cart.UpdateItem)
	cart.DELETE("", r.cart.RemoveItem)

	orders := r.rooAPI.Group("/orders")
	orders.POST("", r.order.Create)
	orders.GET("", r.order.List)
	orders.GET("/:id", r.order.GetByID)
}
