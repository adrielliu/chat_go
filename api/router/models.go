package router

type FormCheckSessionId struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}
