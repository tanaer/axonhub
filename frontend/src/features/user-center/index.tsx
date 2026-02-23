import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from '@tanstack/react-router';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuthStore } from '@/stores/authStore';

const sidebarItems = [
  { id: 'account', labelKey: 'userCenter.sidebar.account', icon: '👤' },
  { id: 'models', labelKey: 'userCenter.sidebar.models', icon: '🤖' },
  { id: 'api-keys', labelKey: 'userCenter.sidebar.apiKeys', icon: '🔑' },
  { id: 'packages', labelKey: 'userCenter.sidebar.packages', icon: '📦' },
  { id: 'billing', labelKey: 'userCenter.sidebar.billing', icon: '💳' },
  { id: 'referral', labelKey: 'userCenter.sidebar.referral', icon: '🎁' },
];

function UserCenterSidebar() {
  const { t } = useTranslation();
  const location = useLocation();
  const currentPath = location.pathname.split('/').pop() || 'account';

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
              currentPath === item.id
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
  const { auth } = useAuthStore();
  
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
        <Button variant="outline">{t('userCenter.account.edit') || '编辑资料'}</Button>
      </CardContent>
    </Card>
  );
}

function ModelsView() {
  const { t } = useTranslation();
  
  const models = [
    { name: 'GPT-4', provider: 'OpenAI', price: '$0.03/1K tokens' },
    { name: 'Claude 3.5 Sonnet', provider: 'Anthropic', price: '$0.003/1K tokens' },
    { name: 'Gemini Pro', provider: 'Google', price: '$0.001/1K tokens' },
    { name: 'GLM-4', provider: '智谱 AI', price: '¥0.01/1K tokens' },
  ];
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.models.title') || '可用模型'}</CardTitle>
        <CardDescription>{t('userCenter.models.description') || '查看所有支持的 AI 模型及价格'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {models.map((model) => (
            <div key={model.name} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div>
                <p className="font-medium text-gray-900">{model.name}</p>
                <p className="text-sm text-gray-500">{model.provider}</p>
              </div>
              <p className="text-sm font-medium text-indigo-600">{model.price}</p>
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
        <Button>{t('userCenter.apiKeys.create') || '创建新密钥'}</Button>
        <div className="text-sm text-gray-500">暂无 API Key</div>
      </CardContent>
    </Card>
  );
}

function PackagesView() {
  const { t } = useTranslation();
  
  const packages = [
    { name: '基础版', price: '¥99/月', tokens: '100K tokens' },
    { name: '专业版', price: '¥299/月', tokens: '500K tokens' },
    { name: '企业版', price: '¥999/月', tokens: '无限制' },
  ];
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.packages.title') || '套餐管理'}</CardTitle>
        <CardDescription>{t('userCenter.packages.description') || '选择适合您的套餐'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-3 gap-4">
          {packages.map((pkg) => (
            <Card key={pkg.name} className="border-2 hover:border-indigo-500 cursor-pointer">
              <CardHeader>
                <CardTitle className="text-lg">{pkg.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-indigo-600">{pkg.price}</p>
                <p className="text-sm text-gray-500 mt-2">{pkg.tokens}</p>
                <Button className="w-full mt-4" variant="outline">{t('userCenter.packages.select') || '选择'}</Button>
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
  
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.billing.title') || '充值记录'}</CardTitle>
        <CardDescription>{t('userCenter.billing.description') || '查看您的充值和消费记录'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500">暂无记录</div>
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
      <CardContent className="space-y-4">
        <div className="p-4 bg-indigo-50 rounded-lg">
          <p className="text-sm text-gray-700">您的推荐码</p>
          <p className="text-lg font-mono font-bold text-indigo-700">MUSKAPI-XXXXX</p>
        </div>
        <Button>{t('userCenter.referral.copy') || '复制推荐链接'}</Button>
      </CardContent>
    </Card>
  );
}

export default function UserCenter() {
  const { t } = useTranslation();
  const location = useLocation();
  const currentPath = location.pathname.split('/').pop() || 'account';

  const renderContent = () => {
    switch (currentPath) {
      case 'models':
        return <ModelsView />;
      case 'api-keys':
        return <ApiKeysView />;
      case 'packages':
        return <PackagesView />;
      case 'billing':
        return <BillingView />;
      case 'referral':
        return <ReferralView />;
      default:
        return <AccountSettings />;
    }
  };

  return (
    <div className="flex min-h-screen bg-gray-50">
      <UserCenterSidebar />
      <main className="flex-1 p-8">
        {renderContent()}
      </main>
    </div>
  );
}
