# MuskAPI v2.0 开发日志

## 版本信息
- **开发分支**: feature/muskapi-v2
- **Git Commit**: d66a38b2
- **开发环境**: https://dev.claudeai.best
- **生产日期**: 2026-02-24

## 新增功能

### 1. Landing Page (首页)
**文件**: `frontend/src/features/landing/index.tsx`

完整的营销页面，包含：
- Hero 区域：主标题、描述、CTA 按钮
- 功能展示：6 个核心功能卡片
- 定价计划：基础版/专业版/企业版
- CTA 区域：引导用户注册
- Footer：版权和链接

### 2. 社交登录
**文件**: `frontend/src/features/auth/components/social-login-buttons.tsx`

支持：
- Google OAuth
- GitHub OAuth

已集成到：
- 登录表单 (`sign-in/components/user-auth-form.tsx`)
- 注册表单 (`sign-up/components/sign-up-form.tsx`)

### 3. 用户中心
**文件**: `frontend/src/features/user-center/index.tsx`

包含 6 个子功能：
- 账户设置：显示用户信息
- 模型中心：查看可用 AI 模型
- API Keys：管理 API 密钥
- 套餐管理：选择和购买套餐
- 充值记录：查看充值历史
- 推荐返利：推荐码管理

### 4. OAuth 回调
**文件**: `frontend/src/routes/(auth)/auth/callback.tsx`

处理社交登录后的回调流程：
- 接收 OAuth code
- 交换 token（需配置后端）
- 重定向到首页

### 5. 首页路由
**文件**: `frontend/src/routes/index.tsx`

智能路由逻辑：
- 未登录 → 显示 Landing Page
- 已登录 → 显示 Dashboard

### 6. 语言包
**文件**: 
- `frontend/src/locales/en/base.json`
- `frontend/src/locales/zh-CN/base.json`

新增 26 条翻译：
- `auth.socialLogin.*` - 社交登录相关
- `userCenter.*` - 用户中心相关（23 条）

## 技术栈
- React 19
- TanStack Router
- Tailwind CSS
- react-i18next

## 待完成
1. OAuth 后端 API 配置（需要 Google/GitHub OAuth 凭证）
2. 生产环境部署和测试

## 访问地址
| 环境 | 域名 | 端口 |
|------|------|------|
| 生产 | https://hub.claudeai.best | 5173 |
| 开发 | https://dev.claudeai.best | 5174 |
