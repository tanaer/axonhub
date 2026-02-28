import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export function PackagesPage() {
  const { t } = useTranslation();

  const packages = [
    { name: t('userCenter.packages.basic') || '基础版', price: '¥99', tokens: '100K', features: ['5 API Keys', t('userCenter.packages.basicModels') || '基础模型', t('userCenter.packages.emailSupport') || '邮件支持'] },
    { name: t('userCenter.packages.pro') || '专业版', price: '¥299', tokens: '500K', features: [t('userCenter.packages.unlimitedKeys') || '无限 API Keys', t('userCenter.packages.allModels') || '全模型', t('userCenter.packages.prioritySupport') || '优先支持'] },
    { name: t('userCenter.packages.enterprise') || '企业版', price: t('userCenter.packages.contact') || '联系我们', tokens: t('userCenter.packages.unlimited') || '无限制', features: [t('userCenter.packages.privateDeployment') || '私有部署', 'SLA', '7x24 支持'] },
  ];

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t('userCenter.packages.title') || '套餐管理'}</CardTitle>
          <CardDescription>{t('userCenter.packages.description') || '选择适合您的套餐'}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {packages.map((pkg, idx) => (
              <Card key={idx} className={idx === 1 ? 'border-indigo-500' : ''}>
                <CardHeader>
                  <CardTitle>{pkg.name}</CardTitle>
                  <p className="text-3xl font-bold">{pkg.price}<span className="text-sm text-gray-500">/月</span></p>
                  <p className="text-sm text-gray-500">{pkg.tokens} Tokens</p>
                </CardHeader>
                <CardContent className="space-y-4">
                  <ul className="space-y-2">
                    {pkg.features.map((f, i) => (
                      <li key={i} className="flex items-center gap-2">
                        <span className="text-green-500">✓</span>
                        {f}
                      </li>
                    ))}
                  </ul>
                  <Button className="w-full" variant={idx === 1 ? 'default' : 'outline'}>
                    {t('userCenter.packages.subscribe') || '立即订阅'}
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
