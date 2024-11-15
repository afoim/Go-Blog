!url: proxy-github
!create: 2024-11-11

### WARNING

1. 请禁用 `/` `/login` `/signup`，否则极易被Cloudflare识别为钓鱼欺诈网站，严重情况下你的域名会被[Hold](https://help.aliyun.com/zh/dws/support/how-to-unlock-a-domain-name-that-is-in-the-serverhold-or-clienthold-state)

2. 请勿在公众网站、论坛大肆宣传，可能会被一些专门投诉钓鱼\欺诈\病毒网站的组织人工审查你的网站，然后给Cloudflare发函，导致第一条情况的发生

### 基本实现思路

1. 先通过一个Workers反代 `github.com` ，并且重写所有HTTP目标为 `github.com` 的请求到你的域名

2. 由于Github有多个外部资源域名，如 `raw.githubusercontent.com` ，需要挨个使用Workers代理并重写

### 具体步骤
1. 登录Cloudflare，点击Workers，新建一个Worker编辑代码如下：

2. 然后依次绑定域名，这里以`*.acofork.us.kg`为例
```javascript
// 域名映射配置
const domain_mappings = {
  'github.com': 'gh.acofork.us.kg',
  'avatars.githubusercontent.com': 'avatars-githubusercontent-com.acofork.us.kg',
  'github.githubassets.com': 'github-githubassets-com.acofork.us.kg',
  'collector.github.com': 'collector-github-com.acofork.us.kg',
  'api.github.com': 'api-github-com.acofork.us.kg',
  'raw.githubusercontent.com': 'raw-githubusercontent-com.acofork.us.kg',
  'gist.githubusercontent.com': 'gist-githubusercontent-com.acofork.us.kg',
  'github.io': 'github-io.acofork.us.kg',
  'assets-cdn.github.com': 'assets-cdn-github-com.acofork.us.kg',
  'cdn.jsdelivr.net': 'cdn.jsdelivr-net.acofork.us.kg',
  'securitylab.github.com': 'securitylab-github-com.acofork.us.kg',
  'www.githubstatus.com': 'www-githubstatus-com.acofork.us.kg',
  'npmjs.com': 'npmjs-com.acofork.us.kg',
  'git-lfs.github.com': 'git-lfs-github-com.acofork.us.kg',
  'githubusercontent.com': 'githubusercontent-com.acofork.us.kg',
  'github.global.ssl.fastly.net': 'github-global-ssl-fastly-net.acofork.us.kg',
  'api.npms.io': 'api-npms-io.acofork.us.kg',
  'github.community': 'github-community.acofork.us.kg'
};


// 反向映射表，用于快速查找原始域名
const reverse_mappings = Object.fromEntries(
  Object.entries(domain_mappings).map(([key, value]) => [value, key])
);

// 需要重定向的路径
const redirect_paths = ['/', '/login', '/signup'];

addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  const url = new URL(request.url);
  const current_host = url.host;
  
  // 检查特殊路径重定向
  if (redirect_paths.includes(url.pathname)) {
    return Response.redirect('https://www.gov.cn', 302);
  }

  // 强制使用 HTTPS
  if (url.protocol === 'http:') {
    url.protocol = 'https:';
    return Response.redirect(url.href);
  }

  // 查找原始目标域名
  const target_host = reverse_mappings[current_host];
  if (!target_host) {
    return new Response('Domain not configured for proxy', { status: 404 });
  }

  // 构建新的请求URL
  const new_url = new URL(url);
  new_url.host = target_host;
  new_url.protocol = 'https:';

  // 设置新的请求头
  const new_headers = new Headers(request.headers);
  new_headers.set('Host', target_host);
  new_headers.set('Referer', new_url.href);
  
  try {
    // 发起请求
    const response = await fetch(new_url.href, {
      method: request.method,
      headers: new_headers,
      body: request.method !== 'GET' ? request.body : undefined
    });

    // 克隆响应以便处理内容
    const response_clone = response.clone();
    
    // 设置新的响应头
    const new_response_headers = new Headers(response.headers);
    new_response_headers.set('access-control-allow-origin', '*');
    new_response_headers.set('access-control-allow-credentials', 'true');
    new_response_headers.set('cache-control', 'public, max-age=14400');
    new_response_headers.delete('content-security-policy');
    new_response_headers.delete('content-security-policy-report-only');
    new_response_headers.delete('clear-site-data');
    
    // 处理响应内容，替换域名引用
    const modified_body = await modifyResponse(response_clone, current_host);

    return new Response(modified_body, {
      status: response.status,
      headers: new_response_headers
    });
  } catch (err) {
    return new Response(`Proxy Error: ${err.message}`, { status: 502 });
  }
}

async function modifyResponse(response, current_host) {
  // 只处理文本内容
  const content_type = response.headers.get('content-type') || '';
  if (!content_type.includes('text/') && !content_type.includes('application/json') && 
      !content_type.includes('application/javascript') && !content_type.includes('application/xml')) {
    return response.body;
  }

  let text = await response.text();
  
  // 替换所有域名引用
  for (const [original_domain, proxy_domain] of Object.entries(domain_mappings)) {
    const escaped_domain = original_domain.replace(/\./g, '\\.');
    
    // 替换完整URLs
    text = text.replace(
      new RegExp(`https?://${escaped_domain}(?=/|"|'|\\s|$)`, 'g'),
      `https://${proxy_domain}`
    );
    
    // 替换协议相对URLs
    text = text.replace(
      new RegExp(`//${escaped_domain}(?=/|"|'|\\s|$)`, 'g'),
      `//${proxy_domain}`
    );
  }

  // 处理相对路径
  if (current_host === domain_mappings['github.com']) {
    text = text.replace(
      /(?<=["'])\/(?!\/|[a-zA-Z]+:)/g,
      `https://${current_host}/`
    );
  }

  return text;
}
```