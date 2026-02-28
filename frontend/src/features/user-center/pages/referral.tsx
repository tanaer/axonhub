import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';
import { Copy, Gift } from 'lucide-react';

interface ReferralInfo {
  code: string;
  link: string;
  total_referrals: number;
  total_rewards: number;
}

interface ReferralRecord {
  id: string;
  referred_email: string;
  reward_amount: number;
  status: string;
  created_at: string;
}

export function ReferralPage() {
  const { t } = useTranslation();
  const { auth } = useAuthStore();
  const [info, setInfo] = useState<ReferralInfo | null>(null);
  const [records, setRecords] = useState<ReferralRecord[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [infoRes, recordsRes] = await Promise.all([
        fetch('/admin/referral/info', {
          headers: { 'Authorization': `Bearer ${auth.accessToken}` },
        }),
        fetch('/admin/referral/records', {
          headers: { 'Authorization': `Bearer ${auth.accessToken}` },
        }),
      ]);

      if (infoRes.ok) setInfo(await infoRes.json());
      if (recordsRes.ok) {
        const data = await recordsRes.json();
        setRecords(data.records || []);
      }
    } catch (error) {
      console.error('Failed to fetch referral data:', error);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success(t('common.copied') || '已复制');
  };

  const formatAmount = (amount: number) => `¥${(amount / 100).toFixed(2)}`;
  const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('zh-CN');

  return (
    <div className="space-y-6">
      {/* 推荐码 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Gift className="w-5 h-5" />
            {t('userCenter.referral.title') || '邀请好友'}
          </CardTitle>
          <CardDescription>{t('userCenter.referral.description') || '邀请好友注册获得奖励'}</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-20 w-full" />
          ) : (
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium">{t('userCenter.referral.code') || '推荐码'}</label>
                <div className="flex gap-2 mt-1">
                  <Input value={info?.code || '-'} readOnly className="bg-gray-50" />
                  <Button variant="outline" onClick={() => copyToClipboard(info?.code || '')}>
                    <Copy className="w-4 h-4" />
                  </Button>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium">{t('userCenter.referral.link') || '推荐链接'}</label>
                <div className="flex gap-2 mt-1">
                  <Input value={info?.link || '-'} readOnly className="bg-gray-50" />
                  <Button variant="outline" onClick={() => copyToClipboard(info?.link || '')}>
                    <Copy className="w-4 h-4" />
                  </Button>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4 pt-4">
                <div className="p-4 bg-indigo-50 rounded-lg">
                  <p className="text-2xl font-bold text-indigo-600">{info?.total_referrals || 0}</p>
                  <p className="text-sm text-gray-600">{t('userCenter.referral.totalReferrals') || '已邀请人数'}</p>
                </div>
                <div className="p-4 bg-green-50 rounded-lg">
                  <p className="text-2xl font-bold text-green-600">{formatAmount(info?.total_rewards || 0)}</p>
                  <p className="text-sm text-gray-600">{t('userCenter.referral.totalRewards') || '累计奖励'}</p>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 邀请记录 */}
      <Card>
        <CardHeader>
          <CardTitle>{t('userCenter.referral.records') || '邀请记录'}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-40 w-full" />
          ) : records.length === 0 ? (
            <p className="text-center text-gray-500 py-8">{t('userCenter.referral.noRecords') || '暂无邀请记录'}</p>
          ) : (
            <div className="space-y-2">
              {records.map((record) => (
                <div key={record.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div>
                    <p className="font-medium">{record.referred_email}</p>
                    <p className="text-sm text-gray-500">{formatDate(record.created_at)}</p>
                  </div>
                  <p className="font-semibold text-green-600">+{formatAmount(record.reward_amount)}</p>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
