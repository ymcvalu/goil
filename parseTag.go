package goil

import (
	"bytes"
	"fmt"
	"strings"
)

func parseTag(tag string) ([]string, [][]string, error) {
	var (
		fns      = make([]string, 0, 1)
		argses   = make([][]string, 0, 1)
		cur      = 0
		pre      = 0
		at       = 0
		bound    = len(tag)
		needArgs bool
	)

	for cur < bound {
		needArgs = false
		if tag[cur] == ' ' {
			cur++
			continue
		}
		pre = cur
	parseFn:
		for cur < bound {
			switch tag[cur] {
			case ' ':
				at = cur
				cur++
				for cur < bound {
					if tag[cur] != ' ' {
						break parseFn
					}
					cur++
				}
				if cur < bound && tag[cur] == '(' {
					needArgs = true
				}
				break parseFn
			case '(':
				at = cur
				needArgs = true
				break parseFn

				// case '!', '@', '#', '$', '%', '^', '&', '*', '"', '\'', '~', '`', '?', '/', '<', '>', '+', '-', '.', ',':
				// 	return nil, nil, fmt.Errorf("unsupport char '%c' for tag '%s'", tag[cur], tag)
			default:
				cur++
				at = cur
			}

		}
		if at > pre {
			fns = append(fns, tag[pre:at])
		}
		if needArgs {
			args, rb, err := parseArgs(tag, cur)
			if err != nil {
				return nil, nil, err
			}
			argses = append(argses, args)
			cur = rb + 1

		} else {
			argses = append(argses, nil)
		}

	}
	return fns, argses, nil
}

func parseArgs(tag string, lbrace int) ([]string, int, error) {
	rbrace, err := hasArgs(tag, lbrace)
	if err != nil {
		return nil, 0, err
	}
	maxParams := strings.Count(tag, ",")
	args := make([]string, 0, maxParams)
	cur := lbrace + 1
	bound := rbrace
	buf := bytes.Buffer{}
parse:
	for cur < bound {
		if tag[cur] == ' ' {
			cur++
			continue
		}
		if tag[cur] == ',' {
			cur++
			continue
		}
		if tag[cur] == '/' {
			if !isBachslash(tag, cur) {
				arg, rb, err := parseBachslash(tag, cur, bound)
				if err != nil {
					return nil, -1, err
				}
				args = append(args, arg)
				cur = rb
				continue parse
			} else {
				buf.WriteByte('/')
				cur += 2
			}
		}
		for cur < bound {
			switch tag[cur] {
			case '/':
				if !isBachslash(tag, cur) {
					return nil, 0, fmt.Errorf("syntax error: mismatch '/' in expression '%s' at %d", tag, cur)
				}
				buf.WriteByte('/')
				cur += 2
			case ' ':
				cur++
				for cur < bound {
					if tag[cur] == ',' {
						break
					}
					if tag[cur] == ' ' {
						cur++
						continue
					}
					return nil, 0, fmt.Errorf("4:syntax error: mismatch '/' in expression '%s' at %d", tag, cur)
				}
				fallthrough
			case ',':
				arg := string(buf.Bytes())
				args = append(args, arg)
				buf.Reset()
				continue parse
			default:
				buf.WriteByte(tag[cur])
				cur++
			}
		}
		if buf.Len() > 0 {
			arg := string(buf.Bytes())
			args = append(args, arg)
		}

	}

	return args, rbrace, nil
}

func hasArgs(tag string, at int) (int, error) {
	if tag[at] != '(' {
		return -1, nil
	}
	needArgs := false
	pos := at + 1
	bound := len(tag)
match:
	for pos < bound {
		switch tag[pos] {
		case ')':
			needArgs = true
			break match
		case '/':
			if isBachslash(tag, pos) {
				pos += 2
				continue match
			}
			mark := pos
			pos++
			for pos < bound {
				if tag[pos] == '/' {
					if !isBachslash(tag, pos) {
						pos++
						continue match
					} else {
						pos += 2
						continue
					}

				}
				pos++
			}
			return -1, fmt.Errorf("syntax error: mismatch '/' in expression '%s' at %d", tag, mark)
		default:
			pos++
		}
	}
	if needArgs {
		return pos, nil
	}
	return -1, fmt.Errorf("syntax error: mismatch '(' in expression '%s' at %d", tag, at)

}

func parseBachslash(tag string, lb int, bound int) (string, int, error) {
	buf := bytes.Buffer{}
	cur := lb + 1
	mark := cur
	for cur < bound {
		if tag[cur] == '/' {
			if isBachslash(tag, cur) {
				buf.WriteByte('/')
				cur += 2
				continue
			} else {
				cur++
				break
			}
		}
		buf.WriteByte(tag[cur])
		cur++
	}
space:
	for cur < bound {
		switch tag[cur] {
		case ' ':
			cur++
		case ',':
			cur++
			break space
		default:
			return "", 0, fmt.Errorf("1:syntax error:don't expecting char '%c' in '%s' at %d", tag[cur], tag, cur)
		}
	}
	if buf.Len() > 0 {
		arg := string(buf.Bytes())
		return arg, cur, nil
	}
	return "", -1, fmt.Errorf("2:syntax error: mismatch '/' in expression '%s' at %d", tag, mark)
}

func isBachslash(tag string, idx int) bool {
	if tag[idx] == '/' && idx+1 < len(tag) && tag[idx+1] == '/' {
		return true
	}
	return false
}
