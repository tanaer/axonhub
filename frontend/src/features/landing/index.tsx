import { Button } from '@/components/ui/button';
import { Link } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { useSignOut } from '@/features/auth/data/auth';

export default function LandingPage() {
  const auth = useAuthStore((state) => state.auth);
  const isLoggedIn = !!auth.accessToken;
  const signOut = useSignOut();

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 to-slate-800">
      {/* Header */}
      <header className="container mx-auto px-4 py-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-3xl">🚀</span>
          <span className="text-2xl font-bold text-white">MuskAPI</span>
        </div>
        <nav className="flex items-center gap-6">
          <a href="#features" className="text-slate-300 hover:text-white transition-colors">功能</a>
          <a href="#pricing" className="text-slate-300 hover:text-white transition-colors">定价</a>
          {isLoggedIn ? (
            <>
              <span className="text-slate-300">
                欢迎，{auth.user?.firstName || auth.user?.email || '用户'}
              </span>
              <Link to="/dashboard">
                <Button className="bg-indigo-600 hover:bg-indigo-700">进入控制台</Button>
              </Link>
              <Button
                variant="ghost"
                className="text-slate-300 hover:text-white"
                onClick={signOut}
              >
                退出
              </Button>
            </>
          ) : (
            <>
              <Link to="/sign-in">
                <Button className="bg-white/10 text-white border border-white/30 hover:bg-white/20">登录</Button>
              </Link>
              <Link to="/sign-up">
                <Button className="bg-indigo-600 hover:bg-indigo-700">免费开始</Button>
              </Link>
            </>
          )}
        </nav>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 py-20 text-center">
        <h1 className="text-5xl md:text-6xl font-bold text-white mb-6">
          统一的 AI 模型接入平台
        </h1>
        <p className="text-xl text-slate-400 mb-8 max-w-2xl mx-auto">
          一站式接入 20+ AI 服务商，100+ 大语言模型。简化开发，降低成本，提升效率。
        </p>
        <div className="flex gap-4 justify-center">
          <Link to="/sign-up">
            <Button size="lg" className="bg-indigo-600 hover:bg-indigo-700 text-lg px-8">
              免费试用
            </Button>
          </Link>
          <a href="https://docs.muskapi.com" target="_blank" rel="noopener">
            <Button size="lg" className="bg-white/10 text-white border border-white/30 hover:bg-white/20 text-lg px-8">
              查看文档
            </Button>
          </a>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="container mx-auto px-4 py-20">
        <h2 className="text-3xl font-bold text-white text-center mb-12">核心功能</h2>
        <div className="grid md:grid-cols-3 gap-8">
          {[
            { icon: '🔌', title: '统一接口', desc: '兼容 OpenAI/Anthropic API 格式，无缝切换模型' },
            { icon: '💰', title: '成本优化', desc: '智能负载均衡，自动选择最优渠道' },
            { icon: '📊', title: '用量追踪', desc: '实时监控 API 调用，精确计费统计' },
            { icon: '🔒', title: '安全可靠', desc: '企业级安全，支持私有化部署' },
            { icon: '⚡', title: '高性能', desc: '全球 CDN 加速，低延迟响应' },
            { icon: '🛠️', title: '易集成', desc: '5 分钟快速接入，完善的 SDK 支持' },
          ].map((f) => (
            <div key={f.title} className="bg-slate-800/50 rounded-xl p-6 hover:bg-slate-800 transition">
              <div className="text-4xl mb-4">{f.icon}</div>
              <h3 className="text-xl font-semibold text-white mb-2">{f.title}</h3>
              <p className="text-slate-400">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Pricing Section */}
      <section id="pricing" className="container mx-auto px-4 py-20">
        <h2 className="text-3xl font-bold text-white text-center mb-12">选择适合您的套餐</h2>
        <div className="grid md:grid-cols-3 gap-8 max-w-4xl mx-auto">
          {[
            { name: '基础版', price: '¥99', period: '/月', tokens: '100K Tokens', features: ['5 个 API Key', '基础模型接入', '邮件支持'] },
            { name: '专业版', price: '¥299', period: '/月', tokens: '500K Tokens', features: ['无限 API Key', '全模型接入', '优先技术支持', '用量分析报告'], popular: true },
            { name: '企业版', price: '联系我们', period: '', tokens: '无限制', features: ['专属部署', 'SLA 保障', '7×24 技术支持', '定制化开发'] },
          ].map((p) => (
            <div key={p.name} className={`rounded-xl p-6 ${p.popular ? 'bg-indigo-600 ring-2 ring-indigo-400' : 'bg-slate-800/50'}`}>
              {p.popular && <div className="text-indigo-200 text-sm font-medium mb-2">最受欢迎</div>}
              <h3 className="text-xl font-semibold text-white mb-2">{p.name}</h3>
              <div className="mb-4">
                <span className="text-3xl font-bold text-white">{p.price}</span>
                <span className="text-slate-300">{p.period}</span>
              </div>
              <p className="text-slate-300 mb-4">{p.tokens}</p>
              <ul className="space-y-2 mb-6">
                {p.features.map((f) => (
                  <li key={f} className="flex items-center gap-2 text-slate-300">
                    <span>✓</span> {f}
                  </li>
                ))}
              </ul>
              <Button className={`w-full ${p.popular ? 'bg-white text-indigo-600 hover:bg-slate-100' : 'bg-indigo-600 hover:bg-indigo-700'}`}>
                {p.name === '企业版' ? '联系销售' : '立即订阅'}
              </Button>
            </div>
          ))}
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-20 text-center">
        <h2 className="text-3xl font-bold text-white mb-4">准备好开始了吗？</h2>
        <p className="text-slate-400 mb-8">立即注册，获得免费试用额度</p>
        <Link to="/sign-up">
          <Button size="lg" className="bg-indigo-600 hover:bg-indigo-700 text-lg px-8">
            免费开始使用
          </Button>
        </Link>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-700 py-8">
        <div className="container mx-auto px-4 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="text-slate-400">© 2026 MuskAPI. All rights reserved.</div>
          <div className="flex gap-6 text-slate-400">
            <a href="#" className="hover:text-white">隐私政策</a>
            <a href="#" className="hover:text-white">服务条款</a>
            <a href="#" className="hover:text-white">联系我们</a>
          </div>
        </div>
      </footer>
    </div>
  );
}
