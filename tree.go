/**
路由信息存储结构
*/
package goil

import (
	"fmt"
	"strconv"
)

//package init method
func init() {
	//开启严格模式
	setStrictConflictChecked(true)
}

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
	/**
	严格冲突检查：如果静态路由和参数路由(:或者*)冲突，则报错，默认关闭
	非严格冲突检查：优先级静态路由>:>*
	*/
	_STRICT_CONFILCT_MASK
	_DEBUG
)

var (
	flag uint8 = 0
)

func setStrictConflictChecked(f bool) {
	if f {
		flag = flag | _STRICT_CONFILCT_MASK
	} else {
		flag = flag & (^_STRICT_CONFILCT_MASK)
	}
}

func isStrictConflictChecked() bool {
	return flag&_STRICT_CONFILCT_MASK > 0
}

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
		typ          uint8       //节点类别
		priority     uint8       //节点优先级
		maxParams    uint8       //最大参数个数
		pattern      string      //节点对应的模式
		children     *node       //子节点
		next         *node       //友邻兄弟节点
		handlerChain *Middleware //作用于该节点的中间件
	}
)

/**
插入路由节点
参数：
root：路由树根节点
url： 目标url
handler：对应的处理函数
chain：注册节点时传入的中间件
*/
func (root *node) addNode(url string, chain *Middleware) *node {
	if url == "" || url[0] != '/' {
		panic("url must start with '/'")
	}
	if chain == nil {
		panic("fatal error:handler can't be nil")
	}
	//tree root can't be nil
	if root == nil || root.pattern != "/" {
		panic("fatal error:invalid tree node")
	}

	paramNum := getParamNum(url)

	parent := root
	pPattern, cPattern := parent.pattern, url
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
					typ: static,

					pattern:      pPattern[preIdx:],
					children:     parent.children,
					handlerChain: parent.handlerChain,
				}
				for ch := child.children; ch != nil; ch = ch.next {
					if child.maxParams < ch.maxParams {
						child.maxParams = ch.maxParams
					}
				}

				parent.handlerChain = nil
				parent.pattern = pPattern[:preIdx]
				parent.children = child
				//parent.priority++
			}
			//刚好是前缀
			if preIdx == len(cPattern) {
				if parent.handlerChain != nil {

					panic("a handle is already registered for path '" + url + "'")
				}

				parent.handlerChain = chain
				return parent
			}

			cPattern = cPattern[preIdx:]
			idx += preIdx
			cc := cPattern[0]
			//从子节点中查找与cPattern具有相同前缀的节点，进一步提取前缀
			for child := parent.children; child != nil; child = child.next {
				if child.pattern == "" {
					//println something to report a nil node
					continue
				}
				cp := child.pattern[0]
				//严格冲突检查模式
				if cc != '/' {
					if isStrictConflictChecked() {
						//there has been wild node
						if cp == ':' || cp == '*' {
							if cp == cc { //判断是否相同wild节点，如果是,continue
								i := len(child.pattern)
								//长度相同，则需要pattern一致，handler为nil
								if len(cPattern) == i && cPattern == child.pattern {
									if child.handlerChain == nil {
										child.handlerChain = chain
										child.priority++
										return child
									}
								} else if len(cPattern) > i && cPattern[:i] == child.pattern && cPattern[i] == '/' {
									parent = child
									pPattern = parent.pattern
									parent.priority++
									if parent.maxParams < paramNum {
										parent.maxParams = paramNum
									}
									paramNum--
									continue loop
								}
							}
							panic("new path '" + url + "' conflicts with existing wildcard '" + child.pattern + "' in existing prefix '" + url[:idx] + "'")
						}
						//will to add a wild node,and has some others node
						if cc == ':' || cc == '*' {
							panic("new path '" + url + "' conflicts with existing wildcard '" + child.pattern + "' in existing prefix '" + url[:idx] + "'")
						}

					} else {
						//判断是否是相同的 wild node
						if (cp == ':' || cp == '*') && cp == cc {
							i := len(child.pattern)
							if len(cPattern) == i && cPattern == child.pattern {
								if child.handlerChain == nil {
									child.handlerChain = chain
									child.priority++
									return child
								}
							} else if len(cPattern) > i && cPattern[:i] == child.pattern && cPattern[i] == '/' {
								parent = child
								pPattern = parent.pattern
								parent.priority++
								if parent.maxParams < paramNum {
									parent.maxParams = paramNum
								}
								paramNum--
								continue loop
							}
						}
					}
				}

				if cc == cp {
					parent = child
					pPattern = parent.pattern
					parent.priority++
					if parent.maxParams < paramNum {
						parent.maxParams = paramNum
					}
					continue loop
				}
			}
			//已经没有公共前缀了，添加新的子节点
			return parent.appendChild(paramNum, cPattern, url, chain)

		}
	} else { //insert root "/"
		if root.handlerChain == nil {
			root.handlerChain = chain
			root.typ = static
			return root
		} else {
			panic("a handle is already registered for path '" + url + "'")
		}
	}
	panic("fatal error when add route node '" + url + "'")
}

func (n *node) appendChild(numParams uint8, pattern, url string, chain *Middleware) (child *node) {
	pl := len(pattern)
	if pl == 0 {
		return nil
	}
	if pattern[0] == '*' && n.pattern[len(n.pattern)-1] != '/' {
		panic("no '/' before '*' in path '" + url + "'")
	}
	parent := n
	buf := make([]byte, 0, pl)
	idx := 0
	for idx < pl {
		//append wild node
		if pattern[idx] == ':' || pattern[idx] == '*' {
			i := idx

			if pattern[idx] == '*' && idx > 0 && pattern[idx-1] != '/' {
				panic("no '/' before '*' in path '" + url + "'")
			}

			if len(buf) > 0 {
				child = &node{
					pattern:   string(buf),
					maxParams: numParams,
					next:      parent.children,
					typ:       static,
				}
				parent.children = child
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
					panic("only one wildcard per path segment is allowed, has:'" + pattern[i:] + "' in path '" + url + "'")
				}
				//* can't end with /
				if pattern[i] == '*' && pattern[idx] == '/' {
					panic("wildcard '*' are only allowed at the end of the path in path '" + url + "'")
				}
				buf = append(buf, pattern[idx])
				idx++
			}
			//* must in the end of url
			if pattern[i] == '*' && idx < pl {
				panic("wildcard '*' are only allowed at the end of the path in path '" + url + "'")
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
				next:      parent.children,
				typ:       typ,
			}
			parent.children = child
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
			next:         parent.children,
			typ:          static,
		}
		parent.children = child
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

func (root *node) routerMapping(path string) (chain *Middleware, params Params, tsr bool) {
	idx := 0
	pl := len(path)

	curNode := root
	//保存当前节点的父节点
	var preNode *node = nil
	//保存可用的参数节点
	var paramNode *node = nil
	//保存可用的通配符节点
	var catchAllNode *node = nil
	//使用局部函数处理返回值
	//ret := func() {
	//	chain = curNode.getHandlerChain()
	//	if chain == nil {
	//		for ch := curNode.children; ch != nil; ch = ch.next {
	//			if ch.pattern == "/" {
	//				tsr = true
	//				break
	//			}
	//		}
	//	}
	//}

loop:
	for idx < pl {
		switch curNode.typ {
		case param:
			if params == nil {
				params = make(Params, 0, curNode.maxParams)
			}
			param := Param{
				Key: curNode.pattern[1:],
			}
			//匿名参数，添加默认参数名
			if param.Key == "" {
				param.Key = "param" + strconv.Itoa(len(params))
			}
			i := idx
			for idx < pl && path[idx] != '/' {
				idx++
			}
			param.Value = path[i:idx]
			params = append(params, param)

			if idx >= pl {
				chain = curNode.getHandlerChain()
				if chain == nil {
					for ch := curNode.children; ch != nil; ch = ch.next {
						if ch.pattern == "/" {
							tsr = true
						} else if ch.typ == catchAll {
							tsr = false
							catchAllNode = ch
							goto final
						}
					}
				}
				//ret()
				return
			}
			goto next

		case catchAll:
			catchAllNode = curNode
			goto final
		default:
			lg := len(curNode.pattern)
			lv := len(path) - idx
			//匹配当前节点
			if lv >= lg && path[idx:idx+lg] == curNode.pattern {
				idx += lg
				if lv == lg {
					chain = curNode.getHandlerChain()
					if chain == nil {
						for ch := curNode.children; ch != nil; ch = ch.next {
							if ch.pattern == "/" {
								tsr = true
							} else if ch.typ == catchAll {
								tsr = false
								catchAllNode = ch
								goto final
							}
						}
					}
					//ret()
					return
				} else {
					goto next
				}
			}

			if !isStrictConflictChecked() && paramNode != nil {
				curNode = paramNode
				paramNode = nil
				continue loop
			}
			goto final

		}
	next:
		preNode = curNode
		paramNode = nil
		for ch := curNode.children; ch != nil; ch = ch.next {
			switch ch.typ {
			case param:
				//strict conflict check mode
				if isStrictConflictChecked() {
					curNode = ch
					continue loop
				} else {
					paramNode = ch
					continue
				}
			case catchAll:
				if isStrictConflictChecked() {
					curNode = ch
					//continue loop
					catchAllNode = ch
					goto final
				} else {
					catchAllNode = ch
					continue
				}
			default:
				if ch.pattern[0] == path[idx] {
					curNode = ch
					continue loop
				}
			}
		}
		break
	}

final:
	//catchAllNode is not nil,use it
	if catchAllNode != nil {
		chain = catchAllNode.getHandlerChain()
		if params == nil {
			params = make(Params, 0, catchAllNode.maxParams)
		}
		param := Param{
			Key: catchAllNode.pattern[1:],
		}
		param.Value = path[idx:]
		if param.Key == "" {
			param.Key = "param" + strconv.Itoa(len(params))
		}
		params = append(params, param)
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

func (root *node) adjustPriority() {

}

/**
打印Tree，深度优先遍历
*/
func printTree(n *node, level int, pre string) {
	i := level
	prefix := ""
	for i > 0 {
		prefix += pre
		i--
	}
	fmt.Printf("%spattern:%s  maxParamNum:%d  type:%d  priority:%d hasHandler:%v\n", prefix, n.pattern, n.maxParams, n.typ, n.priority, n.handlerChain != nil)
	for child := n.children; child != nil; child = child.next {
		printTree(child, level+1, pre)
	}
}

func (n *node) getHandlerChain() *Middleware {
	return n.handlerChain
}
