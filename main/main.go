package main

import (
	"goil"
	"goil/session"
	"strconv"
)

type UserInfo struct {
	UserID   int    `path:"user_id"`
	Username string `validator:"reg(^[a-zA-Z]*$)"`
	Sex      string `validator:"reg(^[M|F]$)"`
	Age      int
	Location string
}

func main() {
	app := goil.Default()
	xrouter := app.XRouter()
	xrouter.POST("/user/:user_id", func(c *goil.Context, userInfo *UserInfo) {
		sess := session.SessionRead(c)
		sess.Set(userInfo.UserID, userInfo)
		c.Text("succ")
	})
	xrouter.GET("/user/:user_id", func(c *goil.Context) {
		sess := session.SessionRead(c)
		userID, _ := strconv.Atoi(c.Param("user_id"))
		info := sess.Get(userID)
		c.JSON(info)
	})
	app.Run(":8081")
}
