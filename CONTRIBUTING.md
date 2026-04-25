# 贡献指南

感谢你对 Nanobot-Go 项目的关注！我们欢迎所有形式的贡献。

## 🎯 贡献方式

### 1. 报告问题
- 使用 GitHub Issues 提交 Bug 报告或功能请求
- 提供详细的复现步骤和环境信息
- 附上相关日志和错误信息

### 2. 修复 Bug
- 在 Issues 中认领 Bug 修复任务
- 遵循代码风格和测试要求
- 确保修复不会破坏现有功能

### 3. 添加新功能
- 先在 Issues 中讨论新功能的想法
- 实现功能后提交 Pull Request
- 包含必要的文档和测试

### 4. 改进文档
- 修复文档中的错误或不清楚的地方
- 添加使用示例和配置说明
- 更新 API 文档和开发指南

## 🚀 开发流程

### 1. Fork 项目
```bash
git clone https://github.com/your-username/nanobot-go.git
cd nanobot-go
```

### 2. 创建分支
```bash
git checkout -b feature/AmazingFeature
# 或者
git checkout -b fix/bug-fix
```

### 3. 本地开发
```bash
# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 运行程序
go run cmd/nanobot/main.go -mode cli
```

### 4. 提交更改
```bash
# 添加修改的文件
git add .

# 提交更改
git commit -m "描述你的更改"

# 推送到你的仓库
git push origin feature/AmazingFeature
```

### 5. 创建 Pull Request
- 在 GitHub 上创建 Pull Request
- 详细描述你的更改
- 关联相关的 Issue

## 📝 代码规范

### Go 语言规范
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代码
- 编写清晰的注释和文档

### 命名规范
- **包名**: 小写，简洁
- **函数名**: 驼峰式，动词开头
- **变量名**: 驼峰式，有意义的名称
- **常量名**: 大写，下划线分隔

### 代码结构
```
internal/
├── core/        # 核心逻辑
│   └── agent.go # Agent 主类
├── llm/         # LLM 适配器
├── tools/       # 工具实现
└── hooks/       # 事件钩子
```

## 🧪 测试要求

### 单元测试
- 为关键功能编写单元测试
- 使用 `go test` 运行测试
- 保持测试覆盖率 > 80%

### 集成测试
- 测试 LLM 集成
- 验证配置文件读取
- 测试工具执行

### 测试命令
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/llm/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📚 文档要求

### 代码注释
- 公共函数必须有文档注释
- 复杂逻辑需要解释
- 示例代码要完整可运行

### 文档更新
- 修改 README.md 如果 API 有变化
- 更新 DEPLOYMENT.md 如果有部署变更
- 添加新功能的使用示例

## 🔍 代码审查

### PR 要求
- 通过所有 CI/CD 检查
- 代码风格符合规范
- 包含必要的测试
- 文档完整

### 审查标准
- 代码质量高，可读性强
- 功能正确，边界情况处理完善
- 性能优化合理
- 安全考虑充分

## 🎨 设计原则

### 可扩展性
- 使用接口设计，便于扩展
- 插件化架构支持新 LLM 提供商
- Hook 系统支持自定义行为

### 安全性
- 输入验证和过滤
- 权限控制机制
- 敏感操作确认

### 可靠性
- 错误处理和恢复
- 日志记录完整
- 数据持久化可靠

## 🤝 社区准则

### 行为准则
- 尊重他人，保持友好
- 建设性反馈，避免攻击性言论
- 包容不同背景和经验的开发者

### 沟通方式
- 使用清晰、礼貌的语言
- 提供具体的反馈和建议
- 乐于帮助新手开发者

## 📄 许可证

通过贡献代码，你同意你的贡献将在 MIT 许可证下发布。

## 🙏 致谢

感谢所有为 Nanobot-Go 做出贡献的开发者！

---

**注意**: 本指南可能会随着项目发展而更新，请定期查看最新版本。