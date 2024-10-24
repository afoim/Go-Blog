!url: proxy-github
!create: 2024-10-24

### WARNING

1. 请禁用 `/` `/login` `/signup`，否则极易被Cloudflare识别为钓鱼欺诈网站，严重情况下你的域名会被[Hold](https://help.aliyun.com/zh/dws/support/how-to-unlock-a-domain-name-that-is-in-the-serverhold-or-clienthold-state)

2. 请勿在公众网站、论坛大肆宣传，可能会被一些专门投诉钓鱼\欺诈\病毒网站的组织人工审查你的网站，然后给Cloudflare发函，导致第一条情况的发生

### 基本实现思路

1. 先通过一个Workers反代 `github.com` ，并且重写所有HTTP目标为 `github.com` 的请求到你的域名

2. 由于Github有多个外部资源域名，如 `raw.githubusercontent.com` ，需要挨个使用Workers代理并重写
