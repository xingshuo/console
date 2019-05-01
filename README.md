Console
====
    a simple console REPL write by golang
支持功能
----
    1."↑","↓"        查看历史命令
    2."←","→"        控制光标移动
    3."Enter"        完成命令,并纳入历史命令区
    4."Backspace","Del"   删除光标前字符
    5.对指定按键的KeyDown事件设置回调函数(如对Tab键的响应设置自动补全功能的回调)
    6.对"Enter"输入的命令设置解析的回调函数
前置环境
----
    golang
支持平台
----
    linux
安装
----
    go get -v -u -x github.com/xingshuo/console
    git clone https://github.com/xingshuo/console.git
    cd console && go build
运行
----
    ./console  #go build编译出的可执行文件
    CMD>a + 'Tab'   #测试功能5(自动补全)
    CMD>ab + 'Tab'  #测试功能5(自动补全)
    CMD>abs + 'Tab' #测试功能5(自动补全)
    CMD>'Enter'
    CMD>任意字符串 + 'Enter'
    CMD>'↑'         #测试功能1
    CMD>'↓'         #测试功能1
    CMD>quit        #测试功能6

Enjoy it!