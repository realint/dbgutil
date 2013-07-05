项目介绍
========

这是一个模仿hopwatch做的Go语言命令行程序调试工具，原理很简单，就是打印变量和按条件暂停程序。

最早是真有趣团队中的同事做了一个简单的变量打印函数，可以支持嵌套结构的打印。最近看到hopwatch项目后，想到何不把暂停和条件暂停也集成进去呢？于是便有了现在这个项目。

为什么不用hopwatch呢？因为hopwatch要通过页面显示和操作，略繁琐了，Go是针对服务端应用设计的编程语言，所做的大多数都是命令行下运行的程序，只要能通过命令行操作和输出就够用了。

这个调试工具的特色有:

1. 打印变量的时候支持结构体的嵌套
2. 支持指针引用关系的打印
3. 内置了一个堆栈打印函数，内部做了忽略Go语言运行时库堆栈跟踪，可以让程序错误日志内容更清晰

这些代码原理虽然简单，自己写起来还是得花不少力气调试的，现在现成的工具就在眼前，赶紧下载使用吧！

安装方式
========

首选方式:

	go get github.com/realint/dbgutil

也可以手工从github下载最新版的代码放到自己项目中。

在需要调试的代码中使用以下方式引用调试模块:

	import "github.com/realint/dbgutil"

示例 - 输出变量
=======

程序调试时，最常用的方式是打印关键的几个变量，下面演示如何用dbgutil打印变量。

测试代码[test1.go](http://dl.dropboxusercontent.com/s/jj5mtnbwcqv51r6/test1.go)：

![示例代码1](http://dl.dropboxusercontent.com/s/zedzgp79zsg08rt/code1.png)

运行测试：

	go run test1.go

输出结果：

![示例输出1](http://dl.dropboxusercontent.com/s/p1ip1rlbsa4qbux/output1.png)

示例 - 暂停程序
========

有时候调试程序时会需要暂停程序，等待调整测试环境或者查看各种参数后再继续执行，下面演示如何做到这一点。

测试代码[test2.go](http://dl.dropboxusercontent.com/s/2pj04sqqxiiisxr/test2.go)：

![示例代码2](http://dl.dropboxusercontent.com/s/dnem9ewlsxijjiv/code2.png)

运行测试：

	go run test2.go

输出结果：

![示例输出2](http://dl.dropboxusercontent.com/s/cjawd5k9z4fnbly/output2.png)

程序暂停时会打印堆栈跟踪信息，回车后程序将继续运行。

示例 - 输出变量并按条件暂停程序
========

有时候调试程序时会需要在特定条件满足的时才需要候暂停程序，下面演示如何做到这一点。

测试代码[test3.go](http://dl.dropboxusercontent.com/s/o3x8faa7hrwqspv/test3.go)：

![示例代码3](http://dl.dropboxusercontent.com/s/k1gmalmuy5skhsz/code3.png)

运行测试：

	go run test3.go

输出结果：

![示例输出3](http://dl.dropboxusercontent.com/s/cfi5mgn2fd1oxjc/output3.png)

程序暂停时会打印堆栈跟踪信息，回车后程序将继续运行。

示例 - 指针关系打印
========

dbgutil对指针关系打印做了特别处理，一方面可以避免递归引用导致打印陷入死循环，另一方面可以让指针引用关系可视化显示。

示例代码[test4.go](http://dl.dropboxusercontent.com/s/2tfyxqiijdwu6ee/test4.go)：

![示例代码4](http://dl.dropboxusercontent.com/s/rv0jdeccsbx4mcp/code4.png)

运行测试：

	go run test4.go

输出结果:

![示例输出4](http://dl.dropboxusercontent.com/s/6dcvhntt4t4q2xq/output4.png)

示例 - 格式化打印
========

默认的Display行为是把复杂的结构体都用一行文本输出，这是为了便于可视化指针关系，但是有时候复杂结构体格式化输出会更方便查看，下面演示如何格式化输出。

示例代码[test5.go](http://dl.dropboxusercontent.com/s/vrm0k72tuee3siz/test5.go)：

![示例代码5](http://dl.dropboxusercontent.com/s/11eg8tcvc6hcg4n/code5.png)

跟上一个示例的区别是调用了FormatDisplay。

运行测试：

	go run test5.go

输出结果：

![示例输出5](http://dl.dropboxusercontent.com/s/pqdxxx43rmtlzql/output5.png?i=1)
