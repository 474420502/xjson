# 贡献指南

感谢您对 XJSON 项目的关注！我们欢迎各种形式的贡献，包括代码、文档、问题报告和功能建议。

## 目录
- [快速开始](#快速开始)
- [开发环境设置](#开发环境设置)
- [代码规范](#代码规范)
- [提交代码](#提交代码)
- [问题报告](#问题报告)
- [功能请求](#功能请求)

## 快速开始

### 1. Fork 项目

点击 GitHub 上的 "Fork" 按钮来创建项目的副本。

### 2. 克隆到本地

```bash
git clone https://github.com/your-username/xjson.git
cd xjson
```

### 3. 创建分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

## 开发环境设置

### 要求

- Go 1.20 或更高版本
- Git

### 安装依赖

```bash
# 克隆项目后
go mod tidy
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行具体包的测试
go test ./internal/scanner/
go test ./internal/parser/

# 运行基准测试
go test -bench=. ./...

# 运行测试覆盖率
go test -cover ./...
```

### 代码检查

我们使用以下工具确保代码质量：

```bash
# 格式化代码
go fmt ./...

# 静态分析
go vet ./...

# 如果安装了 golangci-lint
golangci-lint run
```

## 代码规范

### Go 代码风格

我们遵循标准的 Go 代码风格：

1. **命名规范**:
   - 包名使用小写，简短且有意义
   - 函数和变量使用 camelCase
   - 常量使用 CamelCase 或 UPPER_CASE
   - 私有成员以小写字母开头，公有成员以大写字母开头

2. **注释规范**:
   ```go
   // Package xjson provides high-performance JSON operations.
   package xjson
   
   // Document represents a JSON document with lazy parsing capabilities.
   type Document struct {
       // ...
   }
   
   // Parse creates a new Document from JSON bytes.
   // It returns an error if the JSON is invalid.
   func Parse(data []byte) (*Document, error) {
       // ...
   }
   ```

3. **错误处理**:
   ```go
   // ✅ 好的错误处理
   result, err := someOperation()
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }
   
   // ❌ 避免忽略错误
   result, _ := someOperation()
   ```

4. **接口设计**:
   - 接口应该简小且专注
   - 优先返回接口而不是具体类型
   - 使用组合而不是继承

### 测试规范

1. **测试文件命名**: `*_test.go`

2. **测试函数命名**: `TestFunctionName` 或 `Test_function_name`

3. **基准测试命名**: `BenchmarkFunctionName`

4. **测试结构**:
   ```go
   func TestDocumentParse(t *testing.T) {
       tests := []struct {
           name    string
           input   string
           want    bool
           wantErr bool
       }{
           {
               name:    "valid JSON",
               input:   `{"name": "test"}`,
               want:    true,
               wantErr: false,
           },
           {
               name:    "invalid JSON",
               input:   `{"name": }`,
               want:    false,
               wantErr: true,
           },
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               doc, err := ParseString(tt.input)
               if (err != nil) != tt.wantErr {
                   t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
                   return
               }
               if got := doc.IsValid(); got != tt.want {
                   t.Errorf("IsValid() = %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

### 文档规范

1. **README 更新**: 如果添加新功能，请更新 README.md 中的示例

2. **API 文档**: 所有公有函数和类型都需要有详细的注释

3. **示例代码**: 复杂功能需要提供使用示例

## 提交代码

### 1. 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <description>

<body>

<footer>
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式化，不影响代码逻辑
- `refactor`: 重构，既不修复 bug 也不添加功能
- `perf`: 性能优化
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动

**示例**:
```
feat(parser): add support for array slice syntax

Add support for Python-style array slicing in XPath queries.
This allows queries like "items[1:3]" to select array elements
from index 1 to 3 (exclusive).

Closes #123
```

### 2. Pull Request 流程

1. **确保测试通过**:
   ```bash
   go test ./...
   go test -race ./...
   ```

2. **运行基准测试**:
   ```bash
   go test -bench=. ./...
   ```

3. **提交前检查**:
   ```bash
   go fmt ./...
   go vet ./...
   ```

4. **创建 Pull Request**:
   - 使用清晰的标题描述更改
   - 在描述中解释更改的原因和方法
   - 引用相关的 issue

5. **Code Review**:
   - 响应评审意见
   - 根据反馈修改代码
   - 确保 CI 通过

## 问题报告

### Bug 报告

提交 bug 报告时，请包含以下信息：

1. **Bug 描述**: 清晰地描述问题

2. **复现步骤**: 详细的复现步骤
   ```
   1. 创建 Document: `doc, _ := xjson.ParseString('...')`
   2. 执行查询: `result := doc.Query('...')`
   3. 观察到的错误行为
   ```

3. **期望行为**: 描述您期望的正确行为

4. **环境信息**:
   - Go 版本
   - 操作系统
   - XJSON 版本

5. **最小复现代码**:
   ```go
   package main
   
   import "github.com/474420502/xjson"
   
   func main() {
       doc, _ := xjson.ParseString(`{"test": "value"}`)
       result := doc.Query("test")
       // 这里出现问题...
   }
   ```

### 性能问题

报告性能问题时，请提供：

1. **性能基准**: 使用 `go test -bench=.` 的输出
2. **对比数据**: 与其他库或期望性能的对比
3. **测试数据**: 使用的 JSON 数据样本
4. **硬件信息**: CPU、内存等规格

## 功能请求

### 新功能建议

1. **用例描述**: 详细描述使用场景

2. **API 设计建议**:
   ```go
   // 建议的 API 用法
   result := doc.Query("//items[?(@.price.between(10, 20))]")
   ```

3. **替代方案**: 当前的解决方法和局限性

4. **实现考虑**: 如果有实现想法，请分享

### XPath 语法扩展

如果建议扩展查询语法：

1. **语法规范**: 提供 BNF 或 EBNF 语法定义
2. **示例查询**: 多个使用示例
3. **与现有语法的兼容性**: 确保不破坏现有功能
4. **性能影响**: 考虑对现有性能的影响

## 开发路线图

目前的开发优先级：

### Phase 1: 核心功能 (v0.1.0)
- [x] 基础架构设计
- [ ] JSON 扫描器实现
- [ ] XPath 解析器实现
- [ ] 查询执行引擎
- [ ] 基础测试套件

### Phase 2: 性能优化 (v0.2.0)
- [ ] 懒解析优化
- [ ] 零分配查询路径
- [ ] 写时物化机制
- [ ] 性能基准测试

### Phase 3: 功能完善 (v0.3.0)
- [ ] 完整的 XPath 子集支持
- [ ] 错误处理优化
- [ ] 文档完善
- [ ] 示例应用

### Phase 4: 高级特性 (v1.0.0)
- [ ] 线程安全版本
- [ ] 流式处理支持
- [ ] 插件机制
- [ ] 跨语言绑定

## 社区

- **Discussion**: 使用 GitHub Discussions 进行技术讨论
- **Issues**: 报告 bug 或提出功能请求
- **Pull Requests**: 提交代码贡献

## 许可证

通过贡献代码，您同意您的贡献将在 MIT 许可证下发布。

感谢您的贡献！🚀
