import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuthStore } from '@/stores/authStore';
import { RechargeDialog } from '../components/recharge-dialog';

interface QuotaInfo {
  BalanceYuan: number;
  Quota: number;
  UsedQuota: number;
}

interface Transaction {
  id: string;
  type: string;
  amount: number;
  balance_after: number;
  description: string;
  created_at: string;
}

export function BillingPage() {
  const { t } = useTranslation();
  const { auth } = useAuthStore();
  const [quotaInfo, setQuotaInfo] = useState<QuotaInfo | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [rechargeOpen, setRechargeOpen] = useState(false);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [quotaRes, txRes] = await Promise.all([
        fetch('/admin/quota/me', {
          headers: { 'Authorization': `Bearer ${auth.accessToken}` },
        }),
        fetch('/admin/quota/transactions?limit=10', {
          headers: { 'Authorization': `Bearer ${auth.accessToken}` },
        }),
      ]);

      if (quotaRes.ok) {
        setQuotaInfo(await quotaRes.json());
      }
      if (txRes.ok) {
        const data = await txRes.json();
        setTransactions(data.transactions || []);
      }
    } catch (error) {
      console.error('Failed to fetch billing data:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatAmount = (amount: number) => {
    return `¥${(amount / 100).toFixed(2)}`;
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  return (
    <div className="space-y-6">
      {/* 余额卡片 */}
      <Card>
        <CardHeader>
          <CardTitle>{t('userCenter.balance.title') || '账户余额'}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-20 w-full" />
          ) : (
            <div className="flex items-center justify-between">
              <div>
                <p className="text-4xl font-bold text-indigo-600">
                  ¥{quotaInfo?.BalanceYuan?.toFixed(2) ?? '0.00'}
                </p>
                <p className="text-sm text-gray-500 mt-1">
                  {t('userCenter.balance.totalSpent') || '累计消费'}: ¥{quotaInfo?.UsedQuota ? (quotaInfo.UsedQuota / 500000).toFixed(2) : '0.00'}
                </p>
              </div>
              <Button onClick={() => setRechargeOpen(true)}>
                {t('userCenter.balance.recharge') || '立即充值'}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 充值记录 */}
      <Card>
        <CardHeader>
          <CardTitle>{t('userCenter.billing.history') || '交易记录'}</CardTitle>
          <CardDescription>{t('userCenter.billing.historyDesc') || '最近10条交易记录'}</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="space-y-2">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : transactions.length === 0 ? (
            <p className="text-center text-gray-500 py-8">{t('userCenter.billing.noTransactions') || '暂无交易记录'}</p>
          ) : (
            <div className="space-y-2">
              {transactions.map((tx) => (
                <div key={tx.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div>
                    <p className="font-medium">{tx.description || tx.type}</p>
                    <p className="text-sm text-gray-500">{formatDate(tx.created_at)}</p>
                  </div>
                  <div className="text-right">
                    <p className={`font-semibold ${tx.amount > 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {tx.amount > 0 ? '+' : ''}{formatAmount(tx.amount)}
                    </p>
                    <p className="text-sm text-gray-500">
                      {t('userCenter.billing.balance') || '余额'}: {formatAmount(tx.balance_after)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 充值弹窗 */}
      <RechargeDialog open={rechargeOpen} onOpenChange={setRechargeOpen} onSuccess={fetchData} />
    </div>
  );
}
