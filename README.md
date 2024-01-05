<div align="center">

# go-kyutorin

_✨ 基于 [Satori](https://satori.js.org/zh-CN/) 协议的 QQ 官方机器人 API GoLang 实现 ✨_

</div>

## 引用

本项目引用了这些项目，并进行了一些改动

- [`tencent-connect/botgo`](https://github.com/tencent-connect/botgo)
- [`dezhishen/satori-model-go`](https://github.com/dezhishen/satori-model-go)

## 说明

本项目是一个基于 [Satori](https://satori.js.org/zh-CN/) 协议的 QQ 官方机器人 API GoLang 实现，目的是可以使用户能够快速地建立起一个 `Satori` 服务端，并能够通过 `Satori` 协议规定的规范化 API 接口实现 QQ 官方机器人。

### 接口

- [x] HTTP API
- [x] WebSocket
- [ ] WebHook

### 实现

<details>
<summary>已实现消息元素</summary>

#### 符合 Satori 协议标准的消息元素

| 元素标签   | 功能      | QQ 频道 | QQ 单聊/群聊 |
|-----------|-----------|---------|-------------|
| -         | [纯文本]   | [x]    | [x]          |
| `<at>`    | [提及用户] | [x]    | [ ]          |
| `<sharp>` | [提及频道] | [x]    | [ ]          |
| `<img>`   | [图片]     | [x]    | [x]          |
| `<audio>` | [语音]     | [ ]    | [x]          |
| `<video>` | [视频]     | [ ]    | [x]          |
| `<quote>` | [引用]     | [x]    | [ ]          |

[纯文本]: https://satori.js.org/zh-CN/protocol/elements.html#%E7%BA%AF%E6%96%87%E6%9C%AC
[提及用户]: https://satori.js.org/zh-CN/protocol/elements.html#%E6%8F%90%E5%8F%8A%E7%94%A8%E6%88%B7
[提及频道]: https://satori.js.org/zh-CN/protocol/elements.html#%E6%8F%90%E5%8F%8A%E9%A2%91%E9%81%93
[图片]: https://satori.js.org/zh-CN/protocol/elements.html#%E5%9B%BE%E7%89%87
[语音]: https://satori.js.org/zh-CN/protocol/elements.html#%E8%AF%AD%E9%9F%B3
[视频]: https://satori.js.org/zh-CN/protocol/elements.html#%E8%A7%86%E9%A2%91
[引用]: https://satori.js.org/zh-CN/protocol/elements.html#%E5%BC%95%E7%94%A8

#### 拓展消息元素

| 拓展元素标签 | 功能       | QQ 频道 | QQ 单聊/群聊 |
|-------------|-----------|---------|-------------|
| `<passive>` | [被动消息] | [x]     | [x]         |

</details>

<details>
<summary>已实现 API</summary>

#### 符合 Satori 协议标准的 API

| API                  | 功能              | QQ 频道 | QQ 单聊/群聊 |
|----------------------|-------------------|--------|--------------|
| /channel.get         | [获取群组频道]     | [x]    | [ ]          |
| /channel.list        | [获取群组频道列表] | [x]    | [ ]          |
| /channle.create      | [创建群组频道]     | [x]    | [ ]          |
| /channel.update      | [修改群组频道]     | [x]    | [ ]          |
| /channel.delete      | [删除群组频道]     | [x]    | [ ]          |
| /user.channel.create | [创建私聊频道]     | [x]    | [x]          |
| /guild.get           | [获取群组]         | [x]    | [ ]          |
| /guild.list          | [获取群组列表]     | [x]    | [ ]          |
| /guild.member.get    | [获取群组成员]     | [x]    | [ ]          |
| /guild.member.list   | [获取群组成员列表] | [x]    | [ ]          |
| /guild.member.kick   | [踢出群组成员]     | [x]    | [ ]          |
| /guild.role.list     | [获取群组角色列表] | [x]    | [ ]          |
| /guild.role.create   | [创建群组角色]     | [x]    | [ ]          |
| /guild.role.update   | [修改群组角色]     | [x]    | [ ]          |
| /guild.role.delete   | [删除群组角色]     | [x]    | [ ]          |
| /login.get           | [获取登录信息]     | [x]    | [x]          |
| /message.create      | [发送消息]         | [x]    | [x]          |
| /message.get         | [获取消息]         | [x]    | [ ]          |
| /message.delete      | [撤回消息]         | [x]    | [ ]          |
| /message.update      | [编辑消息]         | [x]    | [ ]          |
| /message.list        | [获取消息列表]     | [x]    | [ ]          |
| /reaction.create     | [添加表态]         | [x]    | [ ]          |
| /reaction.delete     | [删除表态]         | [x]    | [ ]          |
| /reaction.list       | [获取表态列表]     | [x]    | [ ]          |

[获取群组频道]: https://satori.js.org/zh-CN/resources/channel.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84%E9%A2%91%E9%81%93
[获取群组频道列表]: https://satori.js.org/zh-CN/resources/channel.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84%E9%A2%91%E9%81%93%E5%88%97%E8%A1%A8
[创建群组频道]: https://satori.js.org/zh-CN/resources/channel.html#%E5%88%9B%E5%BB%BA%E7%BE%A4%E7%BB%84%E9%A2%91%E9%81%93
[修改群组频道]: https://satori.js.org/zh-CN/resources/channel.html#%E4%BF%AE%E6%94%B9%E7%BE%A4%E7%BB%84%E9%A2%91%E9%81%93
[删除群组频道]: https://satori.js.org/zh-CN/resources/channel.html#%E5%88%A0%E9%99%A4%E7%BE%A4%E7%BB%84%E9%A2%91%E9%81%93
[创建私聊频道]: https://satori.js.org/zh-CN/resources/channel.html#%E5%88%9B%E5%BB%BA%E7%A7%81%E8%81%8A%E9%A2%91%E9%81%93
[获取群组]: https://satori.js.org/zh-CN/resources/guild.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84
[获取群组列表]: https://satori.js.org/zh-CN/resources/guild.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84%E5%88%97%E8%A1%A8
[获取群组成员]: https://satori.js.org/zh-CN/resources/member.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84%E6%88%90%E5%91%98
[获取群组成员列表]: https://satori.js.org/zh-CN/resources/member.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84%E6%88%90%E5%91%98%E5%88%97%E8%A1%A8
[踢出群组成员]: https://satori.js.org/zh-CN/resources/member.html#%E8%B8%A2%E5%87%BA%E7%BE%A4%E7%BB%84%E6%88%90%E5%91%98
[获取群组角色列表]: https://satori.js.org/zh-CN/resources/role.html#%E8%8E%B7%E5%8F%96%E7%BE%A4%E7%BB%84%E8%A7%92%E8%89%B2%E5%88%97%E8%A1%A8
[创建群组角色]: https://satori.js.org/zh-CN/resources/role.html#%E5%88%9B%E5%BB%BA%E7%BE%A4%E7%BB%84%E8%A7%92%E8%89%B2
[修改群组角色]: https://satori.js.org/zh-CN/resources/role.html#%E4%BF%AE%E6%94%B9%E7%BE%A4%E7%BB%84%E8%A7%92%E8%89%B2
[删除群组角色]: https://satori.js.org/zh-CN/resources/role.html#%E5%88%A0%E9%99%A4%E7%BE%A4%E7%BB%84%E8%A7%92%E8%89%B2
[获取登录信息]: https://satori.js.org/zh-CN/resources/login.html#%E8%8E%B7%E5%8F%96%E7%99%BB%E5%BD%95%E4%BF%A1%E6%81%AF
[发送消息]: https://satori.js.org/zh-CN/resources/message.html#%E5%8F%91%E9%80%81%E6%B6%88%E6%81%AF
[获取消息]: https://satori.js.org/zh-CN/resources/message.html#%E8%8E%B7%E5%8F%96%E6%B6%88%E6%81%AF
[撤回消息]: https://satori.js.org/zh-CN/resources/message.html#%E6%92%A4%E5%9B%9E%E6%B6%88%E6%81%AF
[编辑消息]: https://satori.js.org/zh-CN/resources/message.html#%E7%BC%96%E8%BE%91%E6%B6%88%E6%81%AF
[获取消息列表]: https://satori.js.org/zh-CN/resources/message.html#%E8%8E%B7%E5%8F%96%E6%B6%88%E6%81%AF%E5%88%97%E8%A1%A8
[添加表态]: https://satori.js.org/zh-CN/resources/reaction.html#%E6%B7%BB%E5%8A%A0%E8%A1%A8%E6%80%81
[删除表态]: https://satori.js.org/zh-CN/resources/reaction.html#%E5%88%A0%E9%99%A4%E8%A1%A8%E6%80%81
[获取表态列表]: https://satori.js.org/zh-CN/resources/reaction.html#%E8%8E%B7%E5%8F%96%E8%A1%A8%E6%80%81%E5%88%97%E8%A1%A8

</details>

<details>
<summary>已实现的事件</summary>

#### 符合 Satori 协议标准的事件

| 事件类型              | 事件                    | QQ 频道 | QQ 单聊/群聊 |
|----------------------|-------------------------|---------|-------------|
| guild-added          | [加入群组时触发]         | [x]     | [ ]         |
| guild-updated        | [群组被修改时触发]       | [x]     | [ ]         |
| guild-removed        | [退出群组时触发]         | [x]     | [ ]         |
| guild-member-added   | [群组成员增加时触发]     | [x]     | [ ]         |
| guild-member-updated | [群组成员信息更新时触发] | [x]     | [ ]         |
| guild-member-removed | [群组成员移除时触发]     | [x]     | [ ]         |
| message-created      | [当消息被创建时触发]     | [x]     | [x]         |
| message-deleted      | [当消息被删除时触发]     | [x]     | [ ]         |
| reaction-added       | [当表态被添加时触发]     | [x]     | [ ]         |
| reaction-removed     | [当表态被移除时触发]     | [x]     | [ ]         |

[加入群组时触发]: https://satori.js.org/zh-CN/resources/guild.html#guild-added
[群组被修改时触发]: https://satori.js.org/zh-CN/resources/guild.html#guild-updated
[退出群组时触发]: https://satori.js.org/zh-CN/resources/guild.html#guild-removed
[群组成员增加时触发]: https://satori.js.org/zh-CN/resources/member.html#guild-member-added
[群组成员信息更新时触发]: https://satori.js.org/zh-CN/resources/member.html#guild-member-updated
[群组成员移除时触发]: https://satori.js.org/zh-CN/resources/member.html#guild-member-removed
[当消息被创建时触发]: https://satori.js.org/zh-CN/resources/message.html#message-created
[当消息被删除时触发]: https://satori.js.org/zh-CN/resources/message.html#message-deleted
[当表态被添加时触发]: https://satori.js.org/zh-CN/resources/reaction.html#reaction-added
[当表态被移除时触发]: https://satori.js.org/zh-CN/resources/reaction.html#reaction-removed

#### 不符合 Satori 协议标准的事件

Satori 协议为无法直接通过 Satori 服务端获取的事件提供了 `internal` 事件，这意味着当用户收到 `internal` 事件后，可以直接通过事件结构的 `_type` 字段获取原生事件类型，并通过 `_data` 字段获取原生事件数据。

| 事件类型  | 事件          | QQ 频道 | QQ 单聊/群聊 |
|----------|---------------|---------|-------------|
| internal | [平台原生事件] | [x]     | [ ]         |

[平台原生事件]: https://satori.js.org/zh-CN/advanced/internal.html#%E5%B9%B3%E5%8F%B0%E5%8E%9F%E7%94%9F%E4%BA%8B%E4%BB%B6

与此同时，部分 Satori 协议标准事件也会存在 `_type` 字段和 `_data` 字段，用户可以通过该字段直接访问 QQ 原生事件数据。
