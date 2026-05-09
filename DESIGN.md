### Mashiro monitor

mashiro 是一款轻量级的自托管服务器监控工具，旨在提供简单、高效的服务器性能监控解决方案。它支持通过 Web 界面查看服务器状态，并通过轻量级 Agent 收集数据。

## 技术架构

Web：React

后端：Go

探针实现：Go

示例项目：https://github.com/komari-monitor/komari

## 特性

- **轻量高效**：低资源占用，适合各种规模的服务器。
- **自托管**：完全掌控数据隐私，部署简单。
- **Web 界面**：直观的监控仪表盘，易于使用。

## 前台主要功能

1、汇总在线情况，流量详情，当前速率
2、单个服务器支持查看cpu，内存，磁盘，总流量使用情况
3、单个服务器支持查看下载和上传速率以及流量详情
4、单个服务器支持查看延迟，丢包，使用不同颜色色块显示
5、单个服务器支持查看剩余时间以及运行时间
6、前台支持表格和网格显示切换

## 后台主要功能

1、支持添加服务器，设置月度总流量，流量刷新日，到期时间
2、添加一键部署命令

3、通知功能，支持对接tgbot
4、延迟检测功能，支持TCP，IMCP，HTTP
5、账户管理，修改账户密码

## 图例

![image-20260508145321665](C:\Users\79254\Desktop\Mashiro\DESIGN.assets\image-20260508145321665.png)![image](https://cdn3.ldstatic.com/optimized/4X/a/b/f/abf66b1acb718c9543c27ceae1a99962d566f090_2_690x363.jpeg)

![image](https://cdn3.ldstatic.com/optimized/4X/5/d/b/5db5b29d9c76fe8f6725ea3a8f8d2f97bd80be56_2_690x377.png)

![image-20260508145348487](C:\Users\79254\Desktop\Mashiro\DESIGN.assets\image-20260508145348487.png)
