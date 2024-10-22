### 首先，我们要知道GFW是如何封锁我们的流量的

1. IP黑洞：目前无解，但仅对部分服务黑洞，如谷歌系（谷歌、推特、YouTube等）

2. DNS污染：为域名返回一个假的IP。使用hosts文件强制指定域名对应ip或者使用加密的DNS（DoH、DNS 签名等）

3. HTTP劫持：因为流量不是加密的，GFW作为天然的中间人可以直接进行篡改（如：重定向到404页面、劫持到反诈页面等）。可以使用HTTPS连接规避，但你可能会遇到SNI阻断

4. SNI阻断：在客户端与服务器建立加密连接前，客户端会发送 `Client Hello` 报文，而这个报文是明文，并且一般都会携带 `server_name` ，GFW可以知道你要访问哪个网站，对不在白名单（如：discord.com）的域名进行阻断。因为 `server_name` 实际上是一个扩展，并不强制，你可以不发送它来规避SNI阻断

### 那么，让我们分析一下GFW对于不同网站的封锁情况

百度（[www.baidu.com](https://www.baidu.com)）

1. 我们先ping一下得到ip： `2408:873d:22:18ac:0:ff:b021:1393`

2. 通过Hosts强制绑定通过WireShark进行抓包，可以看到，客户端发送的 `Client Hello` 可以清晰地看到 `Server Name` 字段，并且也能正常收到 `Server Hello` 然后双方便开始通信查看浏览器，网站正常访问让我们试试访问 `discord.com`

DIscord（[discord.com](https://discord.com/)）

1. 我们先ping一下，域名和解析到的IP均不通

2. 此时我们尝试使用 `itdog.cn` 进行v4 ping，并且依次对解析出的域名进行ping

3. 找到通的IP强制绑定Hosts，尝试抓包

4. 在客户端发送 `Client Hello` 后被GFW检测到 `Server Name` 字段，然后GFW向客户端发送一个 `RST` 报文，即要求重置客户端连接。在客户端侧，则会收到 `ERR_CONNECTION_RESET` 即：连接已重置。用户无法访问网页。

### 继续，尝试发送空 `Server Name` 报文

成功访问。在WireShark中并未发现 `Server Name` 字段



---



## 一些解决方案



### tcpioneer

它通过魔改TCP数据包使得GFW无法检测，并且WireShark也无法抓取到 `Client Hello` 报文，但是仍然能建立连接，即服务端发送 `Server Hello` 。使用tcpioneer可以访问大多非IP黑洞的网站

### accesser

它通过创建一个本地代理，将主机的所有流量都路由到accesser中，将TCP握手包中的 `server_name` 删除后再发送给服务端以绕过GFW的封禁