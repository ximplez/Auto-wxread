# Auto-wxread
[![BuildWxread](https://github.com/ximplez/Auto-wxread/actions/workflows/build.yml/badge.svg)](https://github.com/ximplez/Auto-wxread/actions/workflows/build.yml)

自动化的微信阅读打卡工具
## 项目介绍 📚
本项目主要用于自动化的微信阅读打卡，通过模拟浏览器操作，实现自动登录、自动阅读等功能。
相比于其他的微信阅读打卡工具，采用浏览器模拟的方式能够更加准确地模拟用户的行为，避免因网络不稳定等因素导致的打卡失败等问题。
## 功能特性 🚀
- 支持自动登录微信阅读。
- 支持自动阅读指定的书籍。
- 支持自定义阅读时间，支持指定阅读时间和阅读页数。
- 支持自定义通知，包括飞书等。
---
## 使用方法 📝
**GitHub Action部署运行（GitHub运行）**
1. fork 本项目。
2. **设置Repository secrets**：仓库 Settings -> 左侧列表中的 Secrets and variables -> Actions，然后在右侧的 Repository secrets 中添加如下值
   - `NOTIFICATION_CONFIG_JSON`：飞书卡片通知配置 JSON。为空时关闭通知；程序会调用 feishu_bot_gateway 的 `/send_card` 接口
   - `PAT`：GitHub PAT Token（用于上传cookies到GitHub），需要具备repo secret写权限
3. **设置Repository variables**：仓库 Settings -> 左侧列表中的 Secrets and variables -> Actions，然后在左侧的 Repository Variables 中添加如下值
   - `TARGET_TIME`：目标阅读时间(分钟)
   - `BOOK_NAME`：目标书名（可以为空，默认选择最近读的书）

飞书通知已切换为模板卡片通知。任务开始时会发送一张卡片，登录二维码、选书、阅读进度、异常、失败和完成总结都会通过 feishu_bot_gateway 的卡片更新能力更新到同一条消息中，减少刷屏并保留完整执行状态。

`NOTIFICATION_CONFIG_JSON` 示例：

```json
{
  "enabled": true,
  "gatewayBaseUrl": "https://your-worker.example.workers.dev",
  "gatewayAuthToken": "your-gateway-token",
  "appId": "cli_xxx",
  "templateId": "ctp_xxx",
  "templateVersionName": "1.0.0",
  "receiveIdType": "email",
  "receiveId": "name@example.com",
  "appName": "微信读书",
  "openId": "",
  "defaultUrl": "https://weread.qq.com/",
  "progressNotifySeconds": 60
}
```

## 后续计划 📆
- [x] 支持登录后自动保存cookies
- [x] 支持自定义设备模拟
- [ ] 支持自定义浏览器代理
- [ ] 支持更多的通知方式

## 欢迎贡献 🤝
