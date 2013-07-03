package dbgutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
)

type breaker struct {
}

func (this breaker) Break(condition bool) {
	if condition {
		fmt.Fprintf(os.Stderr, "\n[Stack]\n%s", Stack(2))
		fmt.Fprint(os.Stderr, "\npress ENTER to continue")
		fmt.Scanln()
	}
}

func Display(data ...interface{}) breaker {
	var _, file, line, ok = runtime.Caller(1)

	if !ok {
		return breaker{}
	}

	var buf = new(bytes.Buffer)

	fmt.Fprintf(buf, "[Debug] %s:%d \n", strings.SplitN(file, "src/", 2)[1], line)

	fmt.Fprintf(buf, "\n[Variables]")

	for i := 0; i < len(data); i += 2 {
		fmt.Fprintf(buf, "\n%s = %s", data[i], Print(len(data[i].(string))+3, true, data[i+1]))

		if i < len(data)-2 {
			fmt.Fprint(buf, ", ")
		}
	}

	log.Print(buf)

	return breaker{}
}

func Break() {
	var _, file, line, ok = runtime.Caller(1)

	if !ok {
		return
	}

	var buf = new(bytes.Buffer)

	fmt.Fprintf(buf, "[Debug] %s:%d \n", strings.SplitN(file, "src/", 2)[1], line)

	fmt.Fprintf(buf, "\n[Stack]\n%s", Stack(2))

	fmt.Fprintf(buf, "\npress ENTER to continue")

	log.Print(buf)

	fmt.Scanln()
}

func doBreak(skip int) {
}

func Stack(skip int) []byte {
	var buf = new(bytes.Buffer)

	var lines [][]byte
	var lastFile string

	for i := skip; ; i++ {
		var _, file, line, ok = runtime.Caller(i)

		if !ok {
			break
		}

		if !strings.Contains(file, "src/pkg/runtime/") {
			fmt.Fprintf(buf, "%s:%d\n", strings.SplitN(file, "src/", 2)[1], line)

			if file != lastFile {
				data, err := ioutil.ReadFile(file)

				if err != nil {
					continue
				}

				lines = bytes.Split(data, []byte{'\n'})
				lastFile = file
			}

			line-- // in stack trace, lines are 1-indexed but our array is 0-indexed

			fmt.Fprintf(buf, "    %s\n", source(lines, line))
		}
	}

	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.Trim(lines[n], " \t")
}

type pointerInfo struct {
	prev *pointerInfo
	n    int
	addr uintptr
	pos  int
	used []int
}

func Print(headlen int, printPointers bool, data ...interface{}) []byte {
	var buf = new(bytes.Buffer)

	if len(data) > 1 {
		fmt.Fprint(buf, "[")
	}

	for k, v := range data {
		var buf2 = new(bytes.Buffer)
		var pointers *pointerInfo

		printKeyValue(buf2, reflect.ValueOf(v), &pointers)

		if k < len(data)-1 {
			fmt.Fprint(buf, ", ")
		}

		fmt.Fprintln(buf2)

		if printPointers && pointers != nil {
			var anyused = false
			var pointerNum = 0

			for p := pointers; p != nil; p = p.prev {
				if len(p.used) > 0 {
					anyused = true
				}
				pointerNum += 1
				p.n = pointerNum
			}

			if anyused {
				var pointerBufs = make([][]rune, pointerNum+1)

				for i := 0; i < len(pointerBufs); i++ {
					var pointerBuf = make([]rune, buf2.Len()+headlen)

					for j := 0; j < len(pointerBuf); j++ {
						pointerBuf[j] = ' '
					}

					pointerBufs[i] = pointerBuf
				}

				for pn := 0; pn <= pointerNum; pn++ {
					for p := pointers; p != nil; p = p.prev {
						if len(p.used) > 0 && p.n >= pn {
							if pn == p.n {
								pointerBufs[pn][p.pos+headlen] = '└'

								var maxpos = 0

								for i, pos := range p.used {
									if i < len(p.used)-1 {
										pointerBufs[pn][pos+headlen] = '┴'
									} else {
										pointerBufs[pn][pos+headlen] = '┘'
									}

									maxpos = pos
								}

								for i := 0; i < maxpos-p.pos-1; i++ {
									if pointerBufs[pn][i+p.pos+headlen+1] == ' ' {
										pointerBufs[pn][i+p.pos+headlen+1] = '─'
									}
								}
							} else {
								pointerBufs[pn][p.pos+headlen] = '│'

								for _, pos := range p.used {
									if pointerBufs[pn][pos+headlen] == ' ' {
										pointerBufs[pn][pos+headlen] = '│'
									} else {
										pointerBufs[pn][pos+headlen] = '┼'
									}
								}
							}
						}
					}

					buf2.WriteString(string(pointerBufs[pn]) + "\n")
				}

				fmt.Fprintln(buf2)
			}
		}

		buf.Write(buf2.Bytes())
	}

	if len(data) > 1 {
		fmt.Fprint(buf, "]")
	}

	return buf.Bytes()
}

func printKeyValue(buf *bytes.Buffer, val reflect.Value, pointers **pointerInfo) {
	var t = val.Kind()

	switch t {
	case reflect.Bool:
		fmt.Fprint(buf, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(buf, val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
		fmt.Fprint(buf, val.Uint())
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(buf, val.Float())
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(buf, val.Complex())
	case reflect.Ptr, reflect.UnsafePointer:
		if val.IsNil() {
			fmt.Fprint(buf, "nil")
			return
		}

		var addr = val.Elem().UnsafeAddr()

		for p := *pointers; p != nil; p = p.prev {
			if addr == p.addr {
				p.used = append(p.used, buf.Len())
				fmt.Fprint(buf, "&")
				return
			}
		}

		*pointers = &pointerInfo{
			prev: *pointers,
			addr: addr,
			pos:  buf.Len(),
			used: make([]int, 0),
		}

		fmt.Fprint(buf, "&")

		printKeyValue(buf, val.Elem(), pointers)
	case reflect.String:
		fmt.Fprint(buf, "\"", val.String(), "\"")
	case reflect.Interface:
		var value = val.Elem()

		if !value.IsValid() {
			fmt.Fprint(buf, "`Invalid Interface`")
		} else {
			printKeyValue(buf, value, pointers)
		}
	case reflect.Struct:
		var t = val.Type()

		fmt.Fprint(buf, t)
		fmt.Fprint(buf, "{ ")

		for i := 0; i < val.NumField(); i++ {
			fmt.Fprint(buf, t.Field(i).Name)
			fmt.Fprint(buf, ": ")

			printKeyValue(buf, val.Field(i), pointers)

			if i < val.NumField()-1 {
				fmt.Fprint(buf, ", ")
			}
		}
		fmt.Fprint(buf, " }")
	case reflect.Array, reflect.Slice:
		fmt.Fprint(buf, val.Type())
		fmt.Fprint(buf, "[")

		for i := 0; i < val.Len(); i++ {
			printKeyValue(buf, val.Index(i), pointers)

			if i < val.Len()-1 {
				fmt.Fprint(buf, ", ")
			}
		}
		fmt.Fprint(buf, "]")
	case reflect.Map:
		var t = val.Type()
		var keys = val.MapKeys()

		fmt.Fprint(buf, t)
		fmt.Fprint(buf, "{ ")

		for i := 0; i < len(keys); i++ {
			printKeyValue(buf, keys[i], pointers)
			fmt.Fprint(buf, ": ")
			printKeyValue(buf, val.MapIndex(keys[i]), pointers)

			if i < len(keys)-1 {
				fmt.Fprint(buf, ", ")
			}
		}
		fmt.Fprint(buf, " }")
	case reflect.Chan:
		fmt.Fprint(buf, val.Type())
	case reflect.Invalid:
		fmt.Fprint(buf, "`Invalid Type`")
	default:
		fmt.Fprint(buf, "`Could't Print`")
	}
}
