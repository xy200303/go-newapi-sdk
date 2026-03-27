# xy200303/go-newapi-sdk

`xy200303/go-newapi-sdk` 是一个面向 [New API 官方文档](https://docs.newapi.pro/zh/docs/api) 的 Go SDK。

项目目标是把文档中的 AI 模型接口和管理接口整理成标准、可维护、可扩展的 Go SDK，同时保留一部分更适合业务代码直接调用的手写辅助方法。

当前仓库基于 `2026-03-27` 的文档快照整理接口清单，并生成对应的调用树。

## 特性

- 标准 Go 模块结构，模块名为 `xy200303/go-newapi-sdk`
- 对外统一入口包为 `newapi/`
- 同时提供“文档镜像接口树”和“高频业务辅助方法”两种使用方式
- 已按职责拆分为 `newapi/core`、`newapi/aimodel`、`newapi/management`
- 支持默认认证、管理员认证、调用级认证覆盖
- 支持 `Operation.Do` 和 `Operation.DoRaw`
- 测试已迁移到顶层 `test/` 目录，便于从 SDK 外部验证公开 API
- README 使用中文，方便直接对照 New API 文档阅读

## 安装

```bash
go get xy200303/go-newapi-sdk
```

导入方式：

```go
import newapi "xy200303/go-newapi-sdk/newapi"
```

## 适用场景

这个 SDK 适合以下两类场景：

- 你希望完整覆盖官方文档中的接口，并通过路径树直观调用
- 你只关心常见业务动作，例如用户登录、创建用户、创建令牌、查询日志等

如果你更喜欢 `NewApiClient(...)` 这种命名风格，可以直接这样创建客户端：

```go
client, err := newapi.NewApiClient("https://api.example.com")
```

项目中也保留了 `New(...)` 和 `NewClient(...)`，用于兼容其他调用习惯。

## 快速开始

下面是一个最小化示例，使用管理员认证调用系统状态接口：

```go
package main

import (
	"context"
	"fmt"
	"log"

	newapi "xy200303/go-newapi-sdk/newapi"
)

func main() {
	client, err := newapi.NewApiClient(
		"https://api.example.com",
		newapi.WithAdminAuth("root-token", 1),
	)
	if err != nil {
		log.Fatal(err)
	}

	var status map[string]any
	err = client.Management.System.StatusGet.Do(context.Background(), nil, &status)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("system status: %+v\n", status)
}
```

## 客户端创建方式

### 1. 推荐方式

```go
client, err := newapi.NewApiClient(
	"https://api.example.com",
	newapi.WithAdminAuth("root-token", 1),
)
```

### 2. 兼容方式

```go
client, err := newapi.New(
	"https://api.example.com",
	newapi.WithBearerToken("token"),
)
```

### 3. 旧风格快捷构造

```go
client := newapi.NewClient(
	"https://api.example.com",
	"root-token",
	1,
	15,
)
```

## 认证方式

SDK 支持以下认证输入方式。

### 默认认证

适合普通用户令牌或会话场景。

- `WithDefaultAuth(auth)`
- `WithBearerToken(token)`
- `WithSessionCookie(cookie)`
- `WithDefaultUserID(userID)`

示例：

```go
client, err := newapi.NewApiClient(
	"https://api.example.com",
	newapi.WithBearerToken("user-token"),
	newapi.WithDefaultUserID(1001),
)
```

### 管理员认证

适合管理接口调用。

- `WithAdminAuth(rootToken, rootUserID)`

示例：

```go
client, err := newapi.NewApiClient(
	"https://api.example.com",
	newapi.WithAdminAuth("root-token", 1),
)
```

### 调用级覆盖认证

对于自动生成的接口，还可以在 `CallConfig.Auth` 中为单次调用单独覆盖认证：

```go
var result map[string]any

err := client.Management.System.StatusGet.Do(
	context.Background(),
	&newapi.CallConfig{
		Auth: &newapi.Auth{
			BearerToken: "override-token",
			UserID:      1001,
		},
	},
	&result,
)
```

## 两种调用方式

### 1. 使用自动生成的文档接口树

这是这个 SDK 的核心能力，适合覆盖全部文档接口。

调用路径会尽量贴近文档目录，例如：

```go
client.Management.System.StatusGet
client.Management.UserManagement.UserSelfGet
client.AIModel.Chat.OpenAI.CreateChatCompletionPost
client.AIModel.Audio.OpenAI.CreateSpeechPost
```

最简单的调用方式：

```go
var result map[string]any

err := client.Management.UserManagement.UserSelfGet.Do(
	context.Background(),
	nil,
	&result,
)
```

带 JSON Body：

```go
var result map[string]any

err := client.AIModel.Chat.OpenAI.CreateChatCompletionPost.Do(
	context.Background(),
	&newapi.CallConfig{
		JSONBody: map[string]any{
			"model": "gpt-4o-mini",
			"messages": []map[string]any{
				{"role": "user", "content": "hello"},
			},
		},
	},
	&result,
)
```

带路径参数和查询参数：

```go
var result map[string]any

err := client.AIModel.Chat.Gemini.GeminiRelayV1BetaPost.Do(
	context.Background(),
	&newapi.CallConfig{
		PathParams: map[string]any{
			"model": "gemini-2.0-flash",
		},
		Query: map[string]any{
			"stream": true,
		},
		JSONBody: map[string]any{
			"contents": []map[string]any{
				{
					"parts": []map[string]any{
						{"text": "hello"},
					},
				},
			},
		},
	},
	&result,
)
```

### 2. 使用手写辅助方法

这部分 API 更偏向业务操作，适合高频、常见、固定结构的调用。

目前主要集中在：

- 用户登录
- 用户信息获取
- 访问令牌生成
- 用户令牌管理
- 管理员用户管理
- 日志查询
- 兑换码创建

示例：

```go
user, err := client.AdminCreateUserAndGetContext(
	context.Background(),
	"demo-user",
	"strong-password",
	"Demo User",
)
```

```go
sessionCookie, err := client.UserLoginContext(
	context.Background(),
	"demo-user",
	"password",
)
```

```go
token, err := client.UserCreateTokenContext(
	context.Background(),
	"user-access-token",
	1001,
	newapi.CreateTokenRequest{
		Name:           "demo-token",
		UnlimitedQuota: true,
	},
)
```

## `CallConfig` 说明

`Operation.Do(...)` 和 `Operation.DoRaw(...)` 使用 `CallConfig` 控制请求参数。

常用字段：

- `PathParams`：路径参数，例如 `/api/user/{id}` 中的 `id`
- `Query`：查询参数，支持 map 或结构体
- `JSONBody`：自动序列化为 JSON 请求体
- `Body`：原始请求体
- `ContentType`：请求体类型
- `Headers`：额外请求头
- `Auth`：覆盖默认认证

注意：

- `JSONBody` 和 `Body` 不能同时设置
- 如果路由包含路径参数但未传入，会返回错误

## 原始响应处理

如果接口返回流、文件、音频，或者你希望自行处理响应体，可以使用 `DoRaw()`：

```go
resp, err := client.AIModel.Audio.OpenAI.CreateSpeechPost.DoRaw(
	context.Background(),
	&newapi.CallConfig{
		JSONBody: map[string]any{
			"model": "tts-1",
			"input": "hello",
			"voice": "alloy",
		},
	},
)
if err != nil {
	log.Fatal(err)
}
defer resp.Body.Close()
```

如果只是想保留状态码、Header 和 Body，也可以让 `Do(...)` 输出到 `newapi.RawResponse`。

## 错误处理

SDK 使用 `APIError` 表示接口错误：

```go
var result map[string]any
err := client.Management.System.StatusGet.Do(context.Background(), nil, &result)
if err != nil {
	if apiErr, ok := err.(*newapi.APIError); ok {
		fmt.Println(apiErr.StatusCode)
		fmt.Println(apiErr.Message)
		fmt.Println(apiErr.Body)
	}
}
```

某些辅助方法会返回 `newapi.ErrNotFound`，例如管理员搜索用户但没有找到匹配项时。

## 示例目录

`examples/basic/` 目前包含以下示例：

- `examples/basic/main.go`：最小化系统状态查询
- `examples/basic/system-status/main.go`：管理员调用系统状态接口
- `examples/basic/chat-completion/main.go`：OpenAI 兼容聊天补全
- `examples/basic/user-token/main.go`：用户登录、生成访问令牌、创建令牌
- `examples/basic/raw-response/main.go`：直接获取原始 `http.Response`

这些示例中使用到的环境变量包括：

- `NEWAPI_BASE_URL`
- `NEWAPI_TOKEN`
- `NEWAPI_ROOT_TOKEN`
- `NEWAPI_USERNAME`
- `NEWAPI_PASSWORD`

## 目录结构

```text
.
├── endpoint-manifest.json        # 从文档整理出的接口清单
├── examples/basic/               # 示例代码
├── newapi/                       # 对外 SDK 入口
│   ├── aimodel/                  # AI 模型接口树
│   ├── core/                     # 请求、认证、Operation 运行时
│   └── management/               # 管理接口树
├── test/                         # 外部测试
├── tools/generate-operations/    # 接口树生成工具
├── go.mod
└── README.md
```

## 接口生成机制

自动生成接口树的流程如下：

1. 根据 New API 文档整理出接口清单
2. 将接口信息保存到 `endpoint-manifest.json`
3. 运行 `tools/generate-operations/main.go`
4. 生成：
   - `newapi/aimodel/operations_gen.go`
   - `newapi/management/operations_gen.go`

重新生成命令：

```bash
go run ./tools/generate-operations
```

`tools/generate-operations` 本身不负责抓取线上文档，它负责读取本地 `endpoint-manifest.json` 并生成 Go 接口树代码。

## 测试

运行全部测试：

```bash
go test ./...
```

当前测试位于顶层 `test/` 目录，使用外部包方式验证 SDK 的公开 API。

## 注意事项

- 本项目优先保证接口覆盖和 SDK 结构清晰，不保证所有接口都带有强类型请求/响应结构体
- 自动生成接口默认以 `map[string]any` 或自定义结构体承接响应
- 如果 New API 文档发生变化，需要先更新 `endpoint-manifest.json`，再重新生成接口树
- 某些接口命名来自文档路径自动转换，字段名会尽量可读，但仍以文档路径为主要映射依据

## License

本项目使用 [MIT License](./LICENSE)。
