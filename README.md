# proto

#### 介绍
**xx**
web配置->DSL->AppDataUnit
除起始符必须在第一个元素,和校验码必须在最后一个元素,其他元素顺序可以任意

一些思路:
这是[]byte - struct{}
1. 一个结构体A保存 协议组成元素
2. 遍历结构体A,读取完整数据
3. 一个结构体,保存当前 数据单元 解析后的协议元素. 可由结构体A的变量保存
4. 再次遍历结构体A,处理元素功能
5. 处理funcode或data zone时返回一个结构体包含解析数据域后的数据
struce{} - []byte

#### 软件架构
软件架构说明


#### 使用说明

1.  xxxx
2.  xxxx
3.  xxxx

#### 特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  Gitee 官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解 Gitee 上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是 Gitee 最有价值开源项目，是综合评定出的优秀开源项目
5.  Gitee 官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  Gitee 封面人物是一档用来展示 Gitee 会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
