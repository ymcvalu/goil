/**
定义中间件结构
MiddlewareFunc：
	中间件处理函数，应该和路由处理函数具有相同类型签名：func(Context)
中间件分为：
	全局中间件：挂载在	'/' 路由节点下
	组中间件： 挂载在指定的路由节点下
	具体的路由节点中间件：挂载在具体的处理节点下
中间件优先级：全局>组>具体节点，先注册优先级高
root.next=group
group.next=node
node.next=Middleware{
		handler,	//将路由处理函数包装成middleware链到 tail of middleware chain
		nil
	}
 */
package goil

type (
	Middleware struct {
		handler HandlerFunc
		next    *Middleware
	}
)

//wrap handler to middleware
func wrapHandler(handler HandlerFunc) *Middleware {
	return &Middleware{
		handler: handler,
		next:    nil,
	}
}

//new a middleware
func NewMiddleware(handler HandlerFunc, next *Middleware) *Middleware {
	return &Middleware{handler, next}
}

//append some middleware to chain
func appendChain(head *Middleware, middles ...*Middleware) *Middleware {
	n := len(middles)
	//not middles to append
	if n == 0 {
		return head
	}
	//head is nil,middles is the chain
	if head == nil {
		return middles[0]
	}

	//find the tail of the chain
	tail := head
	for tail.next != nil {
		tail = tail.next
	}

	for _, m := range middles {
		head.next = m
		tail = head.next
	}
	return head
}


func makeChain(middlewares ... *Middleware) (head *Middleware) {
	if len(middlewares) == 0 {
		head = nil
	} else if len(middlewares) == 1 {
		head = middlewares[0]
	} else {
		head = middlewares[0]
		pre := head
		for i := 1; i < len(middlewares); i++ {
			pre.next = middlewares[i]
			pre = pre.next
		}
	}
	return
}
