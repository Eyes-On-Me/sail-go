## Sail Go - Go语言通用基础库
### 基本路径说明
* com Component 组件
* mod Module 模块
### 已有功能
1. 编码转换 com/data/convert
2. 加密解密 com/data/crypt
3. 时间处理 com/data/datetime
4. 数学计算 com/data/number
5. 随机数 com/data/random
6. 切片数据处理 com/data/slice
7. 字符串数据处理 com/data/strings
8. ID生成和处理 com/data/uuid
9. 文件系统 com/sys/fs
10. 日期处理 mod/data/calendar
11. Json处理 mod/data/json
12. 日志Log mod/data/log
13. 网络数据抓取 mod/net/fetcher
14. 网络服务框架 mod/net/service (无反射, 性能好, 功能全, 模块化)
15. Web服务端 mod/net/web (开箱即用 100行代码即可包含日志记录, 国际化, 安全加密, 模板功能, 静态文件绑定与路径绑定, 开发模式与发布模式)

### Web示例

```
package main

import (
	"os"

	"github.com/sail-services/sail-go/mod/data/log"
	"github.com/sail-services/sail-go/mod/net/service"
	sWeb "github.com/sail-services/sail-go/mod/net/web"
)

var (
	web = &sWeb.Web{
		Project: &sWeb.Project{
			Name:    "HelloWorld",
			Version: "0.1",
			Url:     "helloworld.com",
		},
		Base: &sWeb.Base{
			Port:            8888,                               // 端口
			SecretKey:       `C\WVM(A&yX/dScm503YdD5.,\>(~c>X?`, // AES统一安全码 (Cookie, CSRF)
			DefaultLang:     "en-US",                            // 默认语言
			I18nLangs:       []string{"en-US", "zh-CN"},         // 语言ID
			I18nNames:       []string{"Engligh", "中文"},          // 语言列表
			PathRootDev:     "../",                              // 开发模式的相对路径
			PathRootRelease: "./",                               // 发布模式的相对路径
			PathLangs:       "langs",                            // 语言文件路径
			PathTemplate:    "html",                             // 模版文件路径
		},
		Pro: &sWeb.Pro{
			PathStatics: [][]string{ // URL绑定静态目录
				{"s", "public"},               // 静态文件目录 (JS, CSS...)
				{"hongbao", "public/hongbao"}, // 推广活动页面 (用HTML写的推广主题)
			},
			StaticFiles: [][]string{ // URL绑定静态文件
				{"/favicon.ico", "public/img/logo_icon.ico"}, // favicon
			},
		},
	}
)

func getIndex(con *service.Context) {
	web.Tpl("index", 100, false, con) // 1.模版文件 2.模版标记(用于主题分类, 例100~199是前台, 200~299是后台, 300~399是专题页面...) 3.是否为PJAX(让网站体验更好) 4.上下文变量
}

func get404(con *service.Context) {
	web.Tpl("404", 200, false, con)
}

func main() {
	web.Log = log.New(os.Stdout, log.LEVEL_DATA, log.DATA_BASIC) // 1.Log输出流(打印到文件还是终端) 2.Log数据显示等级 3.数据详细等级(日期时间代码文件名和行号等)
	web.Init()
	web.Ser.Rou.Get("/", getIndex)
	web.Ser.Rou.NotFound(get404)
	web.Run()
}
```
