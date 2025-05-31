package config

const ConfigTemplate = `# 配置文件
# 请根据注释进行配置，不要删除任意一项

# 日志等级
# 可选项：
#   - 0：关闭日志
#   - 1：仅输出致命错误日志
#   - 2：输出致命错误日志和错误日志
#   - 3：输出致命错误日志、错误日志和警告日志
#   - 4：输出致命错误日志、错误日志、警告日志和信息日志
#   - 5：输出致命错误日志、错误日志、警告日志、信息日志和调试日志
#   - 6/7：输出所有日志
log_level: %d

account: # QQ 机器人配置
  
  # QQ 机器人配置，需要通过 QQ 机器人管理端/开发/开发设置 获取
  # 所有项皆为必填，且请确保填写正确，否则无法正常启动

  bot_id: %d # 机器人 QQ 号
  app_id: %d # 机器人 ID
  token: "%s" # 机器人令牌
  app_secret: "%s" # 机器人密钥

  # 是否使用沙箱环境
  # 目前沙箱环境与群聊不适配，如果需要使用群聊功能，请关闭沙箱环境
  sandbox: %t
  
  # 配置与 QQ 机器人开放平台的连接
  websocket:
    enable: %t # 是否启用 WebSocket 连接
    shards: %d # 分片数，建议保持默认的 1 ，多了不知道会发生什么

    # 事件订阅，请参考注释进行选择
    # 注释掉不需要的事件类型，将需要的事件类型前的注释删除
    # 对于某些事件需要特定的机器人权限，如果订阅了没有权限的事件，将会导致连接失败
    intents:%s
  
  webhook:
    enable: %t # 是否启用 WebHook 回调
    host: "%s" # WebHook 地址
    port: %d # WebHook 端口
    path: "%s" # WebHook 路径

# 本地文件服务器配置
# 请确保配置正确，否则无法正常启动
# enable 默认设置为 false ，如果需要使用本地文件服务器，请将其设置为 true
file_server:
  enable: %t # 是否使用本地文件服务器
  external_url: "%s" # 本地文件服务器公网地址 {{ .Host }}:{{ .Port }}
  ttl: %d # 文件存储时间，单位秒

# 数据库配置
# 关联到部分单聊/群聊 API 的使用以及程序的空间占用
# 请确保你是否需要使用数据库，若不需要请设置关闭
database:

  # 消息数据库配置
  message_database:

    # 是否启用消息数据库
    # 如果不启用消息数据库，将无法通过消息 ID 获取单聊/群聊消息
    enable: %t
    limit: %d # 消息获取数量限制，决定每次使用 API 可以获取多少消息，设置为 0 则无上限

satori: # Satori 配置
  version: %d # Satori 版本，目前只有 1
  path: "%s" # Satori 部署路径，可以为空，如果不为空需要以 / 开头
  token: "%s" # 鉴权令牌，如果不设置则不会进行鉴权

  # 服务器配置
  server:
    host: "%s" # 服务器监听地址
    port: %d # 服务器端口

  # WebHook 配置
  webhook:
    timeout: %d # WebHook 事件推送超时时间，单位为秒，设置为 0 则时间为无限`
