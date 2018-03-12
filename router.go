/**
 *路由操作接口
 */
package goil

//HTTP METHOD TYPE
const (
	//the DEFAULT method when not other method handler define
	DEFAULT = iota
	GET
	POST
	PUT
	DELETE
	HEAD
	OPTIONS
	ALL
)

type (
	Param struct {
		Key   string
		Value string
	}

	Params []Param
)

type Router struct {
}
