# 账户与成员管理 API

## 登录

```bash
isrvd_post "/account/login" '{"username":"<USER>","password":"<PASS>"}'
```

返回：`{"token": "eyJ...", "username": "<USER>"}`

已启用 TOTP 二次验证的账号首次密码校验通过时返回：

```json
{"username":"<USER>","twoFactorRequired":true}
```

随后携带认证器验证码再次登录：

```bash
isrvd_post "/account/login" '{"username":"<USER>","password":"<PASS>","totpCode":"123456"}'
```

| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |
| password | string | 密码 |
| totpCode | string | TOTP 二次验证码；仅账号启用 TOTP 且密码登录时需要 |

> 通常使用 `isrvd_login` 命令而非直接调用此接口。启用 TOTP 后可使用 `isrvd_login <base_url> <username> <password> <totpCode>`，或使用已创建的 API Token。

## OIDC 登录

```bash
# 浏览器跳转，完成 OIDC Authorization Code Flow
GET /api/account/oidc/login
```

回调路径：`GET /api/account/oidc/callback`。登录成功后重定向到 `/?oidc_code=<CODE>`；失败重定向到 `/?oidc_error=<ERROR>`。

用一次性登录码换取系统 JWT：

```bash
isrvd_post "/account/oidc/exchange" '{"code":"<OIDC_CODE>"}'
```

返回：`{"token": "eyJ...", "username": "<USER>"}`

> OIDC 提取的用户名由 `oidc.usernameClaim` 指定，默认 `sub`，必须存在于 `members.username`；不存在时与代理 Header 登录一致，登录失败且不会自动创建成员。一次性 `oidc_code` 是短期凭证，勿复制或写入外部日志；代理 Header 登录模式下不显示 OIDC 登录入口。

## 列出路由权限

```bash
isrvd_get "/account/routes"
```

返回按 `module`、`key` 排序的路由列表。

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | 路由权限键，格式为 `METHOD /api/path`，历史兼容标识 |
| permission | string | 稳定权限 ID，如 `docker.container.list`，独立于 API 路径 |
| module | string | 模块名 |
| label | string | 路由显示名 |
| access | number | 访问级别：`0`=需要具体权限，`1`=登录即可访问，`-1`=匿名访问 |

> 路由未显式配置 `access` 时默认为 `0`，即需要具体权限。稳定权限 ID (`permission`) 不受 API 路径变更影响，推荐使用稳定权限 ID 配置成员权限。

## 列出角色定义

```bash
isrvd_get "/account/roles"
```

返回所有已定义的角色列表。

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 角色名称（唯一标识） |
| description | string | 角色描述 |
| permissions | string[] | 角色包含的稳定权限 ID 列表，如 `["docker.container.list", "docker.image.list"]` |

## 创建 API Token

```bash
isrvd_post "/account/token" '{"name":"<TOKEN_NAME>","expiresIn":2592000}'
```

返回：`{"token": "长效token..."}`

## 修改密码

```bash
isrvd_put "/account/password" '{"oldPassword":"<OLD>","newPassword":"<NEW>"}'
```

## TOTP 二次验证

TOTP 仅作用于账号密码登录；Passkey、OIDC 登录和已签发的 API Token 不触发二次验证。

```bash
# 查询状态
isrvd_get "/account/2fa/status"
# 返回：{"enabled":true}

# 开始绑定，返回 secret 和 otpauth URI
isrvd_post "/account/2fa/totp/begin" '{}'
# 返回：{"secret":"BASE32...","uri":"otpauth://totp/..."}

# 在认证器 App 中添加后，提交验证码完成启用
isrvd_post "/account/2fa/totp/enable" '{"secret":"BASE32...","code":"123456"}'

# 禁用前需输入当前验证码
isrvd_post "/account/2fa/totp/disable" '{"code":"123456"}'
```

| 接口 | 请求字段 | 响应字段 | 说明 |
|------|----------|----------|------|
| `/account/2fa/status` | - | `enabled` boolean | 当前用户是否启用 TOTP |
| `/account/2fa/totp/begin` | - | `secret` string, `uri` string | 生成绑定密钥；`secret` 仅绑定流程返回 |
| `/account/2fa/totp/enable` | `secret` string, `code` string | - | 验证通过后保存密钥并启用 |
| `/account/2fa/totp/disable` | `code` string | - | 验证当前验证码后禁用 |

## 列出成员

```bash
isrvd_get "/account/members"
```

| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |
| homeDirectory | string | 主目录 |
| founder | boolean | 是否为创建者 |
| description | string | 描述 |
| roles | string[] | 成员角色列表（角色权限与 permissions 合并计算） |
| permissions | string[] | 直接权限列表（支持稳定权限 ID 如 `docker.container.list` 和兼容旧格式 `GET /api/docker/containers`） |
| twoFactor | object | 二次验证配置；`totp.secret` 不返回 |

## 创建成员

```bash
isrvd_post "/account/member" '{"username":"<USER>","password":"<PASS>","homeDirectory":"<HOME_DIR>","description":"<DESC>","roles":["viewer"],"permissions":["docker.container.list","docker.image.list"]}'
```

## 更新成员

```bash
isrvd_put "/account/member/<USER>" '{"description":"<DESC>","roles":["viewer"],"permissions":["docker.container.list","docker.image.list","swarm.service.list"]}'
```

> password 为空则不修改。`roles` 引用配置文件中定义的 `roles` 角色名，角色权限与 `permissions` 直接权限合并计算。`permissions` 支持稳定权限 ID（如 `docker.container.list`）和旧格式（如 `GET /api/docker/containers`），推荐使用稳定权限 ID，不受 API 路径变更影响。

## 删除成员

```bash
isrvd_delete "/account/member/<USER>"
```

## Passkey 登录（无需认证）

```bash
# 开始登录（username 为空则使用可发现凭证）
isrvd_post "/account/passkey/login/begin" '{"username":"<USER>"}'
# 返回：{"sessionId":"...","options":{...}}

# 完成登录（凭证数据由浏览器 WebAuthn API 生成，直接发送 JSON）
isrvd_post "/account/passkey/login/finish?sessionId=<SESSION_ID>" '<CREDENTIAL_JSON>'
# 返回：{"token":"eyJ...","username":"<USER>"}
```

## Passkey 管理（需认证）

```bash
# 开始注册
isrvd_post "/account/passkey/register/begin" '{"displayName":"<NAME>"}'
# 返回：{"sessionId":"...","options":{...}}

# 完成注册（凭证数据由浏览器 WebAuthn API 生成，直接发送 JSON）
isrvd_post "/account/passkey/register/finish?sessionId=<SESSION_ID>" '<CREDENTIAL_JSON>'

# 列出当前用户的凭证
isrvd_get "/account/passkey/credentials"
# 返回：[{"idBase64":"...","aaguidBase64":"...","signCount":0,"displayName":"...","addedAt":"..."}]

# 重命名凭证
isrvd_put "/account/passkey/credential/<CREDENTIAL_ID>" '{"displayName":"<NEW_NAME>"}'

# 删除凭证
isrvd_delete "/account/passkey/credential/<CREDENTIAL_ID>"
```
