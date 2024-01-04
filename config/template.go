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
log_level: 4

account: # QQ 机器人配置
  
  # QQ 机器人配置，需要通过 QQ 机器人管理端/开发/开发设置 获取
  # 所有项皆为必填，且请确保填写正确，否则无法正常启动

  bot_id: 123456789 # 机器人 QQ 号
  app_id: 123456789 # 机器人 ID
  token: "" # 机器人令牌
  app_secret: "" # 机器人密钥
    
  # 是否使用沙箱环境
  # 目前沙箱环境与群聊不适配，如果需要使用群聊功能，请关闭沙箱环境
  sandbox: false
  
  # 配置与 QQ 机器人开放平台的连接
  websocket:
    shards: 1 # 分片数，建议保持默认的 1 ，多了不知道会发生什么

    # 事件订阅，请参考注释进行选择
    # 注释掉不需要的事件类型，将需要的事件类型前的注释删除
    # 对于某些事件需要特定的机器人权限，如果订阅了没有权限的事件，将会导致连接失败
    intents:
      - "GUILDS"                       # 频道事件             # 该事件是默认订阅
      - "GUILD_MEMBERS"                # 成员事件             # 该事件是默认订阅的
      #- "GUILD_MESSAGES                # 私域频道消息事件      # 仅 私域 机器人可以设置
      #- "GUILD_MESSAGE_REACTION        # 频道消息表情表态事件
      #- "DIRECT_MESSAGES               # 频道私信事件
      #- "OPEN_FORUMS_EVENT"            # 公域论坛事件          # 此为 公域 事件
      #- "AUDIO_OR_LIVE_CHANNEL_MEMBER" # 音频或直播频道成员事件
      #- "USER_MESSAGES"                # 单聊/群聊消息事件     # 仅拥有单聊/群聊权限的机器人可以设置
      #- "INTERACTION"                  # 互动事件
      #- "MESSAGE_AUDIT"                # 消息审核事件
      #- "FORUMS_EVENT"                 # 私域论坛事件          # 仅 私域 机器人可以设置
      #- "AUDIO_ACTION"                 # 音频机器人事件
      - "PUBLIC_GUILD_MESSAGES"        # 公域频道消息事件       # 该事件是默认订阅

# 本地文件服务器配置
# 请确保配置正确，否则无法正常启动
# use_local_file_server 默认设置为 false ，如果需要使用本地文件服务器，请将其设置为 true
file_server:
  use_local_file_server: false # 是否使用本地文件服务器
  url: "" # 填入公网 IP 或域名
  port: 8080 # 填入端口号

satori: # Satori 配置
  version: 1 # Satori 版本，目前只有 1
  path: "" # Satori 部署路径，可以为空，如果不为空需要以 / 开头
  token: "" # 鉴权令牌，如果不设置则不会进行鉴权

  # 服务器配置
  server:
    host: "127.0.0.1" # 服务器地址
    port: 8080 # 服务器端口`
