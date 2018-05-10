/**
路由信息存储结构
*/
package goil

import (
	"fmt"
	"strconv"
)

const (
	_ = iota
	//静态路由
	static
	//参数路由(:)
	param
	//通配路由(*)
	catchAll
)

const (
	_ uint8 = 1 << iota

	_DEBUG
)

var (
	flag uint8 = 0
)

func setDebug(f bool) {
	if f {
		flag = flag | _DEBUG
	} else {
		flag = flag & (^_DEBUG)
	}
}

func isDebug() bool {
	return flag&_DEBUG > 0
}

type (
	node struct {
		typ          uint8        //节点类别
		priority     uint8        //节点优先级
		maxParams    uint8        //最大参数个数
		pattern      string       //节点对应的模式
		head         *node        //子节点
		tail         *node        //子节点
		next         *node        //右邻兄弟节点
		pre          *node        //左邻兄弟节点
		handlerChain HandlerChain //作用于该节点的中间件
	}
)

func (p *node) adjustPriority(c *node) *node {
	if c == nil {
		return nil
	}
	cur, pre := c, c.pre
	for pre != nil {
		if cur.priority > pre.priority {
			*cur, *pre = *pre, *cur
			cur.next, pre.next = pre.next, cur.next
			cur.pre, pre.pre = pre.pre, cur.pre
			cur, pre = pre, pre.pre
			continue
		}
		break
	}
	return cur
}

/**
插入路由节点
参数:
root: 路由树根节点
path: 目标path
handler: 对应的处理函数
chain: 注册节点时传入的中间件
*/
func (root *node) addNode(path string, chain HandlerChain) *node {
	if path == "" || path[0] != '/' {
		panic("url must start with '/'")
	}
	if chain == nil {
		panic("fatal error:handler can't be nil")
	}
	//tree root can't be nil
	if root == nil || root.pattern != "/" {
		panic("fatal error:invalid tree node")
	}

	paramNum := getParamNum(path)

	parent := root
	pPattern, cPattern := parent.pattern, path
	idx := 0
	preIdx := 0
	//url must start with /
	if cPattern != "/" {
	loop:
		for {

			if parent.maxParams < paramNum {
				parent.maxParams = paramNum
			}
			//存找最大公共前缀
			preIdx = getPrefix(pPattern, cPattern)

			//提取公共前缀
			if preIdx > 0 && preIdx < len(pPattern) {

				child := &node{
					typ:          static,
					pattern:      pPattern[preIdx:],
					head:         parent.head,
					tail:         parent.tail,
					handlerChain: parent.handlerChain,
				}
				for ch := child.head; ch != nil; ch = ch.next {
					if child.maxParams < ch.maxParams {
						child.maxParams = ch.maxParams
					}
				}

				parent.handlerChain = nil
				parent.pattern = pPattern[:preIdx]
				parent.head = child
				parent.tail = child
			}
			//刚好是前缀
			if preIdx == len(cPattern) {
				if parent.handlerChain != nil {

					panic("a handle is already registered for path '" + path + "'")
				}

				parent.handlerChain = chain
				return parent
			}

			cPattern = cPattern[preIdx:]
			idx += preIdx
			cc := cPattern[0]
			//从子节点中查找与cPattern具有相同前缀的节点，进一步提取前缀
			for child := parent.head; child != nil; child = child.next {
				if child.pattern == "" {
					//println something to report a nil node
					continue
				}
				cp := child.pattern[0]

				if cp == ':' || cp == '*' {
					//there has been wild node
					if cp == cc { //判断是否相同wild节点，如果是,continue
						i := len(child.pattern)
						//长度相同，则需要pattern一致，handler为nil
						if len(cPattern) == i && cPattern == child.pattern {
							if child.handlerChain == nil {
								child.handlerChain = chain
								child.priority++
								child = parent.adjustPriority(child)
								return child
							}
						} else if len(cPattern) > i && cPattern[:i] == child.pattern && cPattern[i] == '/' {
							child.priority++
							child = parent.adjustPriority(child)
							parent = child
							pPattern = parent.pattern
							if parent.maxParams < paramNum {
								parent.maxParams = paramNum
							}
							paramNum--
							continue loop
						}
					}

					panic("new path '" + path + "' conflicts with existing wildcard '" + child.pattern + "' in existing prefix '" + path[:idx] + "'")

				}

				if cc == cp {
					child.priority++
					child = parent.adjustPriority(child)
					parent = child
					pPattern = parent.pattern

					if parent.maxParams < paramNum {
						parent.maxParams = paramNum
					}
					continue loop
				}
			}
			//已经没有公共前缀了，添加新的子节点
			return parent.appendChild(paramNum, cPattern, path, chain)

		}
	} else { //insert root "/"
		if root.handlerChain == nil {
			root.handlerChain = chain
			root.typ = static
			return root
		} else {
			panic("a handle is already registered for path '" + path + "'")
		}
	}
	panic("fatal error when add route node '" + path + "'")
}

func (n *node) appendChild(numParams uint8, pattern, path string, chain HandlerChain) (child *node) {
	pl := len(pattern)
	if pl == 0 {
		return nil
	}
	if pattern[0] == '*' && n.pattern[len(n.pattern)-1] != '/' {
		panic("no '/' before '*' in path '" + path + "'")
	}
	parent := n
	buf := make([]byte, 0, pl)
	idx := 0
	for idx < pl {
		//append wild node
		if pattern[idx] == ':' || pattern[idx] == '*' {
			i := idx

			if pattern[idx] == '*' && idx > 0 && pattern[idx-1] != '/' {
				panic("no '/' before '*' in path '" + path + "'")
			}

			if len(buf) > 0 {
				child = &node{
					pattern:   string(buf),
					maxParams: numParams,
					typ:       static,
				}
				if parent.tail == nil {
					parent.tail = child
					parent.head = child
				} else {
					parent.tail.next = child
					child.pre = parent.tail
					parent.tail = child
				}
				parent = child
				child = nil
				//clear the buf
				buf = buf[0:0:pl]
			}
			buf = append(buf, pattern[idx])
			idx++
			for idx < pl && pattern[idx] != '/' {
				//repeat wild char,return error
				if pattern[idx] == ':' || pattern[idx] == '*' {
					panic("only one wildcard per path segment is allowed, has:'" + pattern[i:] + "' in path '" + path + "'")
				}
				//* can't end with /
				if pattern[i] == '*' && pattern[idx] == '/' {
					panic("wildcard '*' are only allowed at the end of the path in path '" + path + "'")
				}
				buf = append(buf, pattern[idx])
				idx++
			}
			//* must in the end of url
			if pattern[i] == '*' && idx < pl {
				panic("wildcard '*' are only allowed at the end of the path in path '" + path + "'")
			}
			typ := uint8(0)
			if pattern[i] == ':' {
				typ = param
			} else {
				typ = catchAll
			}
			child = &node{
				pattern:   string(buf),
				maxParams: numParams,
				typ:       typ,
			}
			if parent.tail == nil {
				parent.tail = child
				parent.head = child
			} else {
				parent.tail.next = child
				child.pre = parent.tail
				parent.tail = child
			}
			parent = child
			numParams--
			if idx < pl {
				child = nil
				buf = buf[0:0:pl]
			} else {
				child.handlerChain = chain
				return
			}

		} else {
			buf = append(buf, pattern[idx])
			idx++
		}
	}

	if len(buf) > 0 {
		child = &node{
			pattern:      string(buf),
			maxParams:    numParams,
			handlerChain: chain,

			typ: static,
		}
		if parent.tail == nil {
			parent.tail = child
			parent.head = child
		} else {
			parent.tail.next = child
			child.pre = parent.tail
			parent.tail = child
		}
	}

	return
}

func getPrefix(p string, c string) int {
	pl := len(p)
	cl := len(c)
	if pl == 0 || cl == 0 {
		return 0
	}
	minl := pl
	if minl > cl {
		minl = cl
	}
	for i := 0; i < minl; i++ {
		if p[i] != c[i] {
			return i
		}
		//if p[i] == ':' || c[i] == ':' {
		//	return i
		//}
		//if p[i] == '*' || c[i] == '*' {
		//	return i
		//}
	}
	return minl
}

func getParamNum(url string) uint8 {
	paramNum := 0
	for i, l := 0, len(url); i < l; i++ {
		if url[i] == ':' || url[i] == '*' {
			paramNum++
			i++
			for i < len(url) && url[i] != '/' {
				i++
			}
		}
	}
	if paramNum > 255 {
		paramNum = 255
	}
	return uint8(paramNum)
}

func (root *node) routerMapping(path string) (chain HandlerChain, params Params, tsr bool) {
	idx := 0
	pl := len(path)

	curNode := root
	//保存当前节点的父节点
	var preNode *node = nil

	//保存可用的通配符节点
	var catchAllNode *node = nil

	//使用局部函数处理返回值
	//ret := func() {
	//	chain = curNode.getHandlerChain()
	//	if chain == nil {
	//		for ch := curNode.head ; ch != nil; ch = ch.next {
	//			if ch.pattern == "/" {
	//				tsr = true
	//				break
	//			}
	//		}
	//	}
	//}

lookup:
	for idx < pl {
		switch curNode.typ {
		case param:
			if params == nil {
				params = make(Params, 0, curNode.maxParams)
			}

			paramKey := curNode.pattern[1:]

			//匿名参数，添加默认参数名
			if paramKey == "" {
				paramKey = "param" + strconv.Itoa(len(params))
			}
			i := idx
			for idx < pl && path[idx] != '/' {
				idx++
			}
			params.set(paramKey, path[i:idx])
			if idx >= pl {
				chain = curNode.getHandlerChain()
				if chain == nil {
					for ch := curNode.head; ch != nil; ch = ch.next {
						if ch.pattern == "/" {
							tsr = true
						} else if ch.typ == catchAll {
							tsr = false
							catchAllNode = ch
							break lookup
						}
					}
				}
				//ret()
				return
			}
			break

		case catchAll:
			catchAllNode = curNode
			break lookup
		default:
			lg := len(curNode.pattern)
			lv := len(path) - idx
			//匹配当前节点
			if lv >= lg && path[idx:idx+lg] == curNode.pattern {
				idx += lg
				if lv == lg {
					chain = curNode.getHandlerChain()
					if chain == nil {
						for ch := curNode.head; ch != nil; ch = ch.next {
							if ch.pattern == "/" {
								tsr = true
							} else if ch.typ == catchAll {
								tsr = false
								catchAllNode = ch
								break lookup
							}
						}
					}
					//ret()
					return
				} else {
					break
				}
			}

			break lookup
		}

		preNode = curNode
		for ch := curNode.head; ch != nil; ch = ch.next {
			switch ch.typ {
			case param:
				//strict conflict check mode

				curNode = ch
				continue lookup

			case catchAll:

				curNode = ch
				//continue loop
				catchAllNode = ch
				break lookup

			default:
				if ch.pattern[0] == path[idx] {
					curNode = ch
					continue lookup
				}
			}
		}

		break
	}

	//catchAllNode is not nil,use it
	if catchAllNode != nil {
		chain = catchAllNode.getHandlerChain()
		if params == nil {
			params = make(Params, catchAllNode.maxParams)
		}
		paramKey := catchAllNode.pattern[1:]

		if paramKey == "" {
			paramKey = "param" + strconv.Itoa(len(params))
		}

		params.set(paramKey, path[idx:])
		return
	}
	chain = nil
	params = nil
	//one '/' extra,suggest to redirect
	if (idx == pl-1 && path[idx] == '/' && preNode.handlerChain != nil) || (path[pl-1] == '/' && path[idx:pl-1] == curNode.pattern && curNode.handlerChain != nil) {
		tsr = true
	} else if path[idx:]+"/" == curNode.pattern && curNode.handlerChain != nil { //less one '/',suggest to redirect
		tsr = true
	}
	return
}

func (n *node) getHandlerChain() HandlerChain {
	return n.handlerChain
}

/**
打印Tree，深度优先遍历
*/
func WalkTree(n *node, level int, pre string) {
	i := level
	prefix := ""
	for i > 0 {
		prefix += pre
		i--
	}
	fmt.Printf("path:%-30s\tmaxParamsNum:%-5d\ttype:%-5d\thandlerNum:%-5d\tpriority:%-5d\n", prefix+n.pattern, n.maxParams, n.typ, len(n.handlerChain), n.priority)
	for child := n.head; child != nil; child = child.next {
		WalkTree(child, level+1, pre)
	}
}

// func (a *App) WalkTree() {
// 	WalkTree(a.router.trees[GET].node, 0, "-")
// }
