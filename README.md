# 本地文件搜索引擎

这是一个可以快速检索本地文件名称和**内容**的工具。同类工具有：`AnyTXT Searcher`

## 用法

1. 下载最新发布包（目前只编译了Windows包，其他操作系统暂时只能自己源码编译）：[https://github.com/MACDfree/asearch/releases](https://github.com/MACDfree/asearch/releases)；
2. 配置`config.json`，主要配置项为需要需要检索的文件路径`matches`；
3. 运行`asearch.exe`，首次运行会根据配置创建索引，索引创建完成后会自动打开浏览器；
4. 在搜索框中输入想要搜索的内容，点击回车进行搜索。（高级搜索语法见[https://blevesearch.com/docs/Query-String-Query/](https://blevesearch.com/docs/Query-String-Query/)）

![首页](/images/首页.png)

![搜索结果](/images/搜索结果.png)

## 技术栈

golang

bleve，搜索引擎，github.com/blevesearch/bleve/v2

sego，中文分词，github.com/huichen/sego

gin，web框架，github.com/gin-gonic/gin

gooxml，msoffice解析，baliance.com/gooxml

bolt，kv存储，go.etcd.io/bbolt

## 局限

1. 无法解析doc和xls文件，只能解析docx和xlsx
2. 还没有实现增量的索引重建功能
