import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useParams } from '@tanstack/react-router';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuthStore } from '@/stores/authStore';
import { RechargeDialog } from './components/recharge-dialog';

const sidebarItems = [
  { id: 'account', labelKey: 'userCenter.sidebar.account', icon: '👤' },
  { id: 'models', labelKey: 'userCenter.sidebar.models', icon: '🤖' },
  { id: 'api-keys', labelKey: 'userCenter.sidebar.apiKeys', icon: '🔑' },
  { id: 'packages', labelKey: 'userCenter.sidebar.packages', icon: '📦' },
  { id: 'billing', labelKey: 'userCenter.sidebar.billing', icon: '💳' },
  { id: 'orders', labelKey: '订单管理', icon: '📋' },
  { id: 'referral', labelKey: 'userCenter.sidebar.referral', icon: '🎁' },
];

function UserCenterSidebar({ currentSection }: { currentSection: string }) {
  const { t } = useTranslation();

  return (
    <div className="w-64 bg-white border-r border-gray-200 min-h-screen">
      <div className="p-6">
        <h2 className="text-lg font-semibold text-gray-900">{t('userCenter.title') || '用户中心'}</h2>
      </div>
      <nav className="px-4 space-y-1">
        {sidebarItems.map((item) => (
          <Link
            key={item.id}
            to={`/user-center/${item.id}`}
            className={cn(
              'flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition-colors',
              currentSection === item.id
                ? 'bg-indigo-50 text-indigo-700'
                : 'text-gray-700 hover:bg-gray-50'
            )}
          >
            <span className="text-lg">{item.icon}</span>
            <span>{t(item.labelKey) || item.id}</span>
          </Link>
        ))}
      </nav>
    </div>
  );
}

function AccountSettings() {
  const { t } = useTranslation();
  const auth = useAuthStore((state) => state.auth);
  const [loading, setLoading] = useState(true);
  
  // 获取用户余额信息
  useEffect(() => {
    const fetchQuota = async () => {
      try {
        const response = await fetch('/admin/quota/me', {
          headers: {
            'Authorization': `Bearer ${auth.accessToken}`
          }
        });
        if (response.ok) {
          const data = await response.json();
          setQuotaInfo(data);
        }
      } catch (error) {
        console.error('Failed to fetch quota:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchQuota();
  }, [auth.accessToken]);
  
  const [quotaInfo, setQuotaInfo] = useState<any>(null);
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.account.title') || '账户设置'}</CardTitle>
        <CardDescription>{t('userCenter.account.description') || '管理您的账户信息'}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-sm font-medium text-gray-700">邮箱</label>
            <p className="text-gray-900">{auth.user?.email || '-'}</p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-700">用户名</label>
            <p className="text-gray-900">{auth.user?.firstName || '-'} {auth.user?.lastName || ''}</p>
          </div>
        </div>
        <div className="pt-4 border-t">
          <div className="flex items-center justify-between p-4 bg-indigo-50 rounded-lg">
            <div>
              <p className="text-sm text-gray-600">账户余额</p>
              {loading ? (
                <div className="h-6 w-16 animate-spin rounded-full border-2 border-indigo-300"></div>
              ) : (
                <p className="text-2xl font-bold text-indigo-600">
                  ¥{(quotaInfo?.balance_yuan || 0).toFixed(2)}
                </p>
              )}
            </div>
            <RechargeDialog onSuccess={() => {
              setLoading(true);
              fetch('/admin/quota/me', {
                headers: { 'Authorization': `Bearer ${auth.accessToken}` }
              }).then(r => r.json()).then(setQuotaInfo).finally(() => setLoading(false));
            }} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function ModelsView() {
  const { t } = useTranslation();
  
  const models = [
    { name: 'GPT-4o', provider: 'OpenAI', price: '$0.005/1K tokens', status: 'active' },
    { name: 'Claude 3.5 Sonnet', provider: 'Anthropic', price: '$0.003/1K tokens', status: 'active' },
    { name: 'Gemini 2.0 Flash', provider: 'Google', price: '$0.001/1K tokens', status: 'active' },
    { name: 'GLM-4', provider: '智谱 AI', price: '¥0.01/1K tokens', status: 'active' },
    { name: 'DeepSeek V3', provider: 'DeepSeek', price: '¥0.001/1K tokens', status: 'active' },
    { name: 'Qwen-Max', provider: '阿里云', price: '¥0.02/1K tokens', status: 'active' },
  ];
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.models.title') || '可用模型'}</CardTitle>
        <CardDescription>{t('userCenter.models.description') || '查看所有支持的 AI 模型及价格'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mb-4 flex gap-2">
          <Button variant="outline" size="sm" className="bg-indigo-50">全部</Button>
          <Button variant="outline" size="sm">文本</Button>
          <Button variant="outline" size="sm">图像</Button>
          <Button variant="outline" size="sm">嵌入</Button>
        </div>
        <div className="space-y-3">
          {models.map((model) => (
            <div key={model.name} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition">
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 bg-indigo-100 rounded-full flex items-center justify-center text-indigo-600 font-bold">
                  {model.provider[0]}
                </div>
                <div>
                  <p className="font-medium text-gray-900">{model.name}</p>
                  <p className="text-sm text-gray-500">{model.provider}</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-indigo-600">{model.price}</p>
                <span className="text-xs text-green-600 bg-green-50 px-2 py-0.5 rounded">可用</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function ApiKeysView() {
  const { t } = useTranslation();
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.apiKeys.title') || 'API Keys'}</CardTitle>
        <CardDescription>{t('userCenter.apiKeys.description') || '管理您的 API 密钥'}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex justify-between items-center">
          <p className="text-sm text-gray-500">您可以通过 API Key 访问所有模型</p>
          <Button>{t('userCenter.apiKeys.create') || '创建新密钥'}</Button>
        </div>
        <div className="border rounded-lg p-8 text-center text-gray-500">
          <p className="text-4xl mb-2">🔑</p>
          <p>暂无 API Key</p>
          <p className="text-sm">创建您的第一个 API Key 开始使用</p>
        </div>
      </CardContent>
    </Card>
  );
}

function PackagesView() {
  const { t } = useTranslation();
  
  const packages = [
    { name: '体验版', price: '¥0', tokens: '10K tokens', features: ['免费试用', '基础模型', '社区支持'], current: true },
    { name: '基础版', price: '¥99', period: '/月', tokens: '100K tokens', features: ['5 个 API Key', '全模型接入', '邮件支持'], popular: false },
    { name: '专业版', price: '¥299', period: '/月', tokens: '500K tokens', features: ['无限 API Key', '优先支持', '用量分析'], popular: true },
    { name: '企业版', price: '联系我们', period: '', tokens: '无限制', features: ['专属部署', 'SLA 保障', '7×24 支持'], popular: false },
  ];
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.packages.title') || '套餐管理'}</CardTitle>
        <CardDescription>{t('userCenter.packages.description') || '选择适合您的套餐'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-4 gap-4">
          {packages.map((pkg) => (
            <Card key={pkg.name} className={cn(
              "border-2 cursor-pointer transition-all",
              pkg.popular ? "border-indigo-500 ring-2 ring-indigo-200" : "border-gray-200 hover:border-indigo-300",
              pkg.current && "bg-indigo-50"
            )}>
              <CardHeader className="pb-2">
                {pkg.popular && <span className="text-xs text-indigo-600 font-medium">推荐</span>}
                {pkg.current && <span className="text-xs text-green-600 font-medium">当前套餐</span>}
                <CardTitle className="text-lg">{pkg.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="mb-4">
                  <span className="text-2xl font-bold text-gray-900">{pkg.price}</span>
                  <span className="text-gray-500">{pkg.period}</span>
                </div>
                <p className="text-sm text-indigo-600 mb-4">{pkg.tokens}</p>
                <ul className="space-y-2 mb-4">
                  {pkg.features.map((f) => (
                    <li key={f} className="text-sm text-gray-600 flex items-center gap-2">
                      <span className="text-green-500">✓</span> {f}
                    </li>
                  ))}
                </ul>
                <Button 
                  className="w-full" 
                  variant={pkg.current ? "outline" : "default"}
                  disabled={pkg.current}
                >
                  {pkg.current ? '当前套餐' : '选择'}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function BillingView() {
  const { t } = useTranslation();
  const auth = useAuthStore((state) => state.auth);
  const [transactions, setTransactions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const fetchTransactions = async () => {
      try {
        const response = await fetch('/admin/quota/transactions', {
          headers: { 'Authorization': `Bearer ${auth.accessToken}` }
        });
        if (response.ok) {
          const data = await response.json();
          setTransactions(data.transactions || []);
        }
      } catch (error) {
        console.error('Failed to fetch transactions:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchTransactions();
  }, [auth.accessToken]);
  
  const formatAmount = (amount: number) => {
    const prefix = amount >= 0 ? '+' : '';
    return `${prefix}¥${(amount / 100).toFixed(2)}`;
  };
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.billing.title') || '充值记录'}</CardTitle>
        <CardDescription>{t('userCenter.billing.description') || '查看您的充值和消费记录'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mb-4 flex gap-2">
          <Button variant="outline" size="sm" className="bg-indigo-50">全部</Button>
          <Button variant="outline" size="sm">充值</Button>
          <Button variant="outline" size="sm">消费</Button>
        </div>
        {loading ? (
          <div className="text-center py-8 text-gray-500">加载中...</div>
        ) : transactions.length === 0 ? (
          <div className="text-center py-8 text-gray-500">暂无交易记录</div>
        ) : (
          <div className="border rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">类型</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">金额</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">描述</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">时间</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">状态</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {transactions.map((tx: any) => (
                  <tr key={tx.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm">{tx.type === 'recharge' ? '充值' : tx.type === 'consume' ? '消费' : tx.type}</td>
                    <td className={cn("px-4 py-3 text-sm font-medium", tx.amount >= 0 ? 'text-green-600' : 'text-red-600')}>
                      {formatAmount(tx.amount)}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">{tx.description || '-'}</td>
                    <td className="px-4 py-3 text-sm text-gray-500">{tx.created_at ? new Date(tx.created_at).toLocaleString('zh-CN') : '-'}</td>
                    <td className="px-4 py-3">
                      <span className="text-xs bg-green-50 text-green-600 px-2 py-1 rounded">{tx.status || '成功'}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function OrdersView() {
  const { t } = useTranslation();
  
  const orders = [
    { id: 'RCH20260225123456', type: '充值', amount: '¥100.00', method: 'EPUSDT', status: '已完成', time: '2026-02-25 12:34' },
    { id: 'PKG20260225123457', type: '套餐', amount: '¥99.00', method: 'Stripe', status: '处理中', time: '2026-02-25 12:35' },
  ];
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>订单管理</CardTitle>
        <CardDescription>查看您的所有订单</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mb-4 flex gap-2">
          <Button variant="outline" size="sm" className="bg-indigo-50">全部</Button>
          <Button variant="outline" size="sm">充值</Button>
          <Button variant="outline" size="sm">套餐</Button>
        </div>
        {orders.length === 0 ? (
          <div className="text-center py-8 text-gray-500">暂无订单</div>
        ) : (
          <div className="border rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">订单号</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">类型</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">金额</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">支付方式</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">状态</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">时间</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {orders.map((order) => (
                  <tr key={order.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-mono">{order.id}</td>
                    <td className="px-4 py-3 text-sm">{order.type}</td>
                    <td className="px-4 py-3 text-sm font-medium">{order.amount}</td>
                    <td className="px-4 py-3 text-sm">{order.method}</td>
                    <td className="px-4 py-3">
                      <span className={cn(
                        "text-xs px-2 py-1 rounded",
                        order.status === '已完成' ? "bg-green-50 text-green-600" : "bg-yellow-50 text-yellow-600"
                      )}>{order.status}</span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">{order.time}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function ReferralView() {
  const { t } = useTranslation();
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.referral.title') || '推荐返利'}</CardTitle>
        <CardDescription>{t('userCenter.referral.description') || '邀请好友获得返利'}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-indigo-50 rounded-lg p-4 text-center">
            <p className="text-sm text-gray-600">邀请人数</p>
            <p className="text-2xl font-bold text-indigo-600">0</p>
          </div>
          <div className="bg-green-50 rounded-lg p-4 text-center">
            <p className="text-sm text-gray-600">累计返利</p>
            <p className="text-2xl font-bold text-green-600">¥0.00</p>
          </div>
          <div className="bg-orange-50 rounded-lg p-4 text-center">
            <p className="text-sm text-gray-600">待结算</p>
            <p className="text-2xl font-bold text-orange-600">¥0.00</p>
          </div>
        </div>
        <div className="p-4 bg-gray-50 rounded-lg">
          <p className="text-sm text-gray-700 mb-2">您的推荐码</p>
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-white px-4 py-2 rounded border font-mono text-lg">MUSKAPI-XXXXX</code>
            <Button variant="outline">复制</Button>
          </div>
          <p className="text-xs text-gray-500 mt-2">分享推荐链接，好友注册后您可获得其消费金额的 10% 返利</p>
        </div>
        <Button className="w-full">{t('userCenter.referral.copy') || '复制推荐链接'}</Button>
      </CardContent>
    </Card>
  );
}

export default function UserCenter() {
  const { t } = useTranslation();
  const params = useParams({ strict: false });
  const currentSection = (params as any).section || 'account';

  const renderContent = () => {
    switch (currentSection) {
      case 'models':
        return <ModelsView />;
      case 'api-keys':
        return <ApiKeysView />;
      case 'packages':
        return <PackagesView />;
      case 'billing':
        return <BillingView />;
      case 'orders':
        return <OrdersView />;
      case 'referral':
        return <ReferralView />;
      default:
        return <AccountSettings />;
    }
  };

  return (
    <div className="flex min-h-screen bg-gray-50">
      <UserCenterSidebar currentSection={currentSection} />
      <main className="flex-1 p-8">
        {renderContent()}
      </main>
    </div>
  );
}
