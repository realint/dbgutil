package dbgutil

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"reflect"
	"runtime"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	lbr       = []byte("{")
	lbrn      = []byte("{\n")
	com       = []byte(",")
	comn      = []byte(",\n")
	rbr       = []byte("}")
	comnrbr   = []byte(",\n}")
)

//
// 变量打印时不格式化代码，（支持指针关系打印）
//
func Display(data ...interface{}) breaker {
	return display(false, data...)
}

//
// 变量打印时格式化代码（不支持指针关系打印）
//
func FormatDisplay(data ...interface{}) breaker {
	return display(true, data...)
}

func display(formated bool, data ...interface{}) breaker {
	var pc, file, line, ok = runtime.Caller(2)

	if !ok {
		return breaker{}
	}

	var buf = new(bytes.Buffer)

	fmt.Fprintf(buf, "[Debug] at %s() [%s:%d]\n", function(pc), file, line)

	fmt.Fprintf(buf, "\n[Variables]\n")

	for i := 0; i < len(data); i += 2 {
		var output []byte

		if formated {
			output = FormatPrint(len(data[i].(string))+3, true, data[i+1])
		} else {
			output = Print(len(data[i].(string))+3, true, data[i+1])
		}

		fmt.Fprintf(buf, "%s = %s", data[i], output)
	}

	log.Print(buf)

	return breaker{}
}

type breaker struct {
}

//
// 根据条件暂停程序并打印堆栈信息，回车后继续运行
//
func (this breaker) Break(condition bool) {
	if condition {
		fmt.Fprintf(os.Stderr, "\n[Stack]\n%s", Stack(2))
		fmt.Fprint(os.Stderr, "\npress ENTER to continue")
		fmt.Scanln()
	}
}

//
// 暂停程序并打印堆栈信息，回车后继续运行
//
func Break() {
	var pc, file, line, ok = runtime.Caller(1)

	if !ok {
		return
	}

	var buf = new(bytes.Buffer)

	fmt.Fprintf(buf, "[Debug] at %s() [%s:%d]\n", function(pc), file, line)

	fmt.Fprintf(buf, "\n[Stack]\n%s", Stack(2))

	fmt.Fprintf(buf, "\npress ENTER to continue")

	log.Print(buf)

	fmt.Scanln()
}

type pointerInfo struct {
	prev *pointerInfo
	n    int
	addr uintptr
	pos  int
	used []int
}

//
// 格式化输出变量值
//
func FormatPrint(headlen int, printPointers bool, data ...interface{}) []byte {
	var code1 = Print(headlen, false, data...)
	var code2 = bytes.Replace(bytes.Replace(bytes.Replace(code1, lbr, lbrn, -1), com, comn, -1), rbr, comnrbr, -1)
	var code3, err = format.Source(code2)

	if err == nil {
		return code3
	}

	return code2
}

//
// 输出变量值
//
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
			fmt.Fprint(buf2, ", ")
		}

		fmt.Fprintln(buf2)

		if printPointers && pointers != nil {
			printPointerInfo(buf2, headlen, pointers)
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
	case reflect.UnsafePointer:
		fmt.Fprintf(buf, "unsafe.Pointer(0x%x)", val.Pointer())
	case reflect.Ptr:
		if val.IsNil() {
			fmt.Fprint(buf, "nil")
			return
		}

		var addr = val.Elem().UnsafeAddr()

		for p := *pointers; p != nil; p = p.prev {
			if addr == p.addr {
				p.used = append(p.used, buf.Len())
				fmt.Fprintf(buf, "0x%x", addr)
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
			fmt.Fprint(buf, "nil")
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
		fmt.Fprint(buf, "[]")
		fmt.Fprint(buf, val.Type())
		fmt.Fprint(buf, "{")

		for i := 0; i < val.Len(); i++ {
			printKeyValue(buf, val.Index(i), pointers)

			if i < val.Len()-1 {
				fmt.Fprint(buf, ", ")
			}
		}
		fmt.Fprint(buf, "}")
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
		fmt.Fprint(buf, "0 /* Invalid Type */")
	default:
		fmt.Fprint(buf, "0 /* Could't Print */")
	}
}

func printPointerInfo(buf *bytes.Buffer, headlen int, pointers *pointerInfo) {
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
			var pointerBuf = make([]rune, buf.Len()+headlen)

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

			buf.WriteString(string(pointerBufs[pn]) + "\n")
		}
	}
}

//
// 获取堆栈信息（从系统自带的debug模块中提取代码改造的）
//
func Stack(skip int) []byte {
	var buf = new(bytes.Buffer)

	for i := skip; ; i++ {
		var pc, file, line, ok = runtime.Caller(i)

		if !ok {
			break
		}

		fmt.Fprintf(buf, "at %s() [%s:%d]\n", function(pc), file, line)
	}

	return buf.Bytes()
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
