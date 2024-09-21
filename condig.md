# Go OpenMCIM HAZARD
为各平台的 Mod 的缓存加速，基于 [Go OpenBmclAPI](https://github.com/LiterMC/go-openbmclapi) 修改的节点客户端
## HAZARD 分支强制跳过了本地存储文件校验，用于测试等不违反社区规范的个人测试用途，如必须使用此分支，请在运营方的许可下执行！！！！

### HAZARD 分支强制跳过了本地存储文件校验，用于测试等不违反社区规范的个人测试用途，如必须使用此分支，请在运营方的许可下执行！！！！

#### HAZARD 分支强制跳过了本地存储文件校验，用于测试等不违反社区规范的个人测试用途，如必须使用此分支，请在运营方的许可下执行！！！！

如需更多关于 MCIM 的信息或有接入等需求，请到[此处](https://github.com/mcmod-info-mirror/mcim)了解

> [!TIP]
> 非官方, 不保证时效性
*如果本项目有用, 就给个 star ⭐️ 吧 :)*

## 部署

以下是基于本项目现有 OpenMCIM 节点端发布的办法
> [!IMPORTANT]
> 如果跑不起来请自行解决或开 issue 进行询问，不要当作 bug 发到人家仓库去

```yaml
log-slots: 7
no-access-log: false
access-log-slots: 16
byoc: true #OpenMCIM 暂时无法下发证书，默认跳过颁发证书
use-cert: true #OpenMCIM 暂时无法下发证书，默认使用你自己的证书
trusted-x-forwarded-for: true
public-host: example.wetem.cn #连接到你使用的域名或IP地址
public-port: 443 #连接到你使用的端口
port: 4000 #内网端口，一般情况不用改
cluster-id: 123456789012345678901234 #24位的节点ID，请自己想办法获取
cluster-secret: 12345678901234567890123456789012 #32位的节点密钥，一样，你都用 HAZARD 分支了你不可能不会获取
sync-interval: 100
only-gc-when-start: false
download-max-conn: 1024
max-reconnect-count: 100
certificates:
  - cert: cert.pem #你自己的证书文件位置，这是pem证书
    key: pem.pem #你自己的证书密钥文件位置，pem证书的key密钥
tunneler: #想必你也不需要打洞
  enable: false
  tunnel-program: ./path/to/tunnel/program
  output-regex: \bNATedAddr\s+(?<host>[0-9.]+|\[[0-9a-f:]+\]):(?<port>\d+)$
  tunnel-timeout: 0
cache:
  type: inmem
serve-limit:
  enable: false
  max-conn: -1
  upload-rate: 0
api-rate-limit:
  anonymous:
    per-minute: 0
    per-hour: 0
  logged:
    per-minute: 0
    per-hour: 0
notification:
  enable-email: false
  email-smtp: smtp.example.com:25
  email-smtp-encryption: tls
  email-sender: noreply@example.com
  email-sender-password: example-password
  enable-webhook: true
dashboard:
  enable: true
  username: akiame
  password: akiame
  pwa-name: GoOpenMCIM Dashboard
  pwa-short_name: GOMCIM Dash
  pwa-description: Go-OpenMCIM Internal Dashboard
  notification-subject: mailto:user@example.com
github-api:
  update-check-interval: 1h0m0s
  authorization: Bearer ghp_xxxx
database:
  driver: sqlite
  data-source-name: files.db
hijack:
  enable: false
  enable-local-cache: false
  local-cache-path: hijack_cache
  require-auth: true
  auth-users:
    - username: example-username
      password: example-password
      
#01 webdav+alist 存储示例
#路径对照示例：
# Alist地址：http://127.0.0.1:5244/ 打开后会显示你挂载的文件夹，例如 mcim
#存储 MCIM 文件的文件夹名称：mcim
storages:
  - type: mount
    id: alist-mount
    weight: 100
    data:
      path: /tmp
      #根据 Alist 文件下载重定向规则，格式为 Alist地址/d/文件夹名称
      #以开头所述的地址举例：http://127.0.0.1:5244/d/mcim
      redirect-base: http://127.0.0.1:5244/d/mcim
      pre-gen-measures: false
  - type: webdav
    id: alist-webdav
    weight: 0
    data:
      max-conn: 32
      max-upload-rate: 0
      max-download-rate: 0
      pre-gen-measures: false
      follow-redirect: false
      redirect-link-cache: 2h0m0s
      alias: alist-user
      #文件夹名称
      #以开头所述的地址举例：mcim
      endpoint: mcim
webdav-users:
  alist-user:
    #Alist 的 WebDav 地址，格式为 Alist地址/dav/
    #以开头所述的地址举例：http://127.0.0.1:5244/dav/
    endpoint: http://127.0.0.1:5244/dav/
    #Alist 后台的账号密码，不用我多说了吧
    #别忘了开启游客访问权限，否则 OpenMCIM 用户民不聊生嗷
    #要是忘了你就给我飞起来
    username: admin
    password: admin
advanced:
  debug-log: true
  socket-io-log: true
  no-heavy-check: true
  no-gc: false
  heavy-check-interval: 120
  keepalive-timeout: 10
  skip-first-sync: true
  skip-signature-check: true
  no-fast-enable: true
  wait-before-enable: 0
  do-NOT-redirect-https-to-SECURE-hostname: true
  do-not-open-faq-on-windows: true

```

