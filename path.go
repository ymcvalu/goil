/**
 * Inspired by httprouter and reference from it
 */
package goil

/**
预处理url：
	1. 处理后的URL必须以 `/` 开头
	2. 去除多余的 `/`
	3. 去除 `.`
	4. 处理 `..`，回退到上一级
	5. 去除起始的 `/..` 和 `../` 为 `/`
	6. "" 则返回 `/`

case：
	input			output
	""				/
    "./"			/
	a/b				/a/b
	"/."			/
	"../test"		/test
	/../test/.		/test/
	../a/../b/..	/
    "///a///b///"	/a/b/
**/

func CleanPath(p string) string {
	//根据上诉规则，""对应"/"
	if p == "" {
		return "/"
	}

	var n = len(p)
	//下一个要读取的index
	r := 0

	//跳过开头可能重复的 '/'
	for r < n && p[r] == '/' {
		r++
	}

	if r == n {
		return "/"
	}

	var buf []byte

	//如果传入的url没以'/'开头，则需要新建buf保存结果url
	if p[0] != '/' {
		buf = make([]byte, n+1)
		buf[0] = '/'
	}

	//下一个要写入的index，URL固定以'/'开头，因此从 1 开始
	w := 1

	//是否以 '/' 结尾
	trailing := false

	for r < n {
		switch {
		case p[r] == '/':
			//skip '/'
			r++

			//匹配 .
		case p[r] == '.' && r+1 == n:
			trailing = true
			r++
			break

		case p[r] == '.' && p[r+1] == '/':
			r += 2

			//匹配 ..
		case p[r] == '.' && p[r+1] == '.' && ( r+2 == n || p[r+2] == '/'):
			//buf[0] == '/'
			if w > 1 {
				w--
			}
			if buf == nil {
				for w > 1 && p[w] != '/' {
					w--
				}
			} else {
				for w > 1 && buf[w] != '/' {
					w--
				}
			}
			r += 3

		default:
			if w > 1 {
				//使用appendLazy写入
				appendLazy(&buf, p, w, '/')
				w++
			}

			for r < n && p[r] != '/' {
				appendLazy(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}

	if trailing || p[n-1] == '/' {
		if w > 1 {
			appendLazy(&buf, p, w, '/')
			w++
		}
	}

	if buf == nil {
		return p[:w]
	}
	return string(buf[:w])
}

//拷贝字节b到(*buf)[w]
//如果 *buf==nil && p[w] == b ，p[:w+1]等价于buf[:w+1]
func appendLazy(buf *[]byte, p string, w int, b byte) {
	//如果buf未初始化
	if *buf == nil {
		if p[w] == b {	//如果可以使用原path，则无需创建buf
			return
		}
		//初始化buf
		*buf = make([]byte, len(p))
		copy(*buf, p[:w])
	}
	//拷贝字节b到buf
	(*buf)[w] = b
}
