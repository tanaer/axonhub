import { useState, useEffect } from 'react';
import { useSearch } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';
import { CheckCircle2, Loader2 } from 'lucide-react';
import { Link } from '@tanstack/react-router';

export function PaymentPage() {
  const { t } = useTranslation();
  const search = useSearch({ from: '/_authenticated/user-center/payment' });
  const { auth } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const orderId = (search as any).order_id || '';
  const amount = parseFloat((search as any).amount || '0');
  const method = (search as any).method || 'epusdt';

  const handleConfirmPayment = async () => {
    if (!orderId || amount <= 0) {
      toast.error(t('userCenter.payment.invalidOrder') || '无效的订单');
      return;
    }

    setLoading(true);
    try {
      const amountFen = Math.round(amount * 7 * 100); // USDT 转 CNY 再转分
      const response = await fetch('/api/payment/mock/confirm', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.accessToken}`,
        },
        body: JSON.stringify({
          order_id: orderId,
          amount: amountFen,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        setSuccess(true);
        toast.success(t('userCenter.payment.success') || '支付成功！余额已到账');
      } else {
        const error = await response.json();
        toast.error(error.error?.message || t('userCenter.payment.failed') || '支付失败');
      }
    } catch (error) {
      toast.error(t('userCenter.payment.failed') || '支付失败');
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <CheckCircle2 className="w-16 h-16 text-green-500 mx-auto mb-4" />
            <h2 className="text-2xl font-bold mb-2">{t('userCenter.payment.successTitle') || '支付成功'}</h2>
            <p className="text-gray-500 mb-4">
              {t('userCenter.payment.successDesc') || '您的余额已到账，可以开始使用了！'}
            </p>
            <p className="text-sm text-gray-400 mb-6">
              {t('userCenter.payment.orderId') || '订单号'}: {orderId}
            </p>
            <Link to="/user-center/billing">
              <Button>{t('userCenter.payment.backToBilling') || '返回账单'}</Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>{t('userCenter.payment.title') || '确认支付'}</CardTitle>
          <CardDescription>
            {t('userCenter.payment.desc') || '请确认您的订单信息'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="flex justify-between py-2 border-b">
              <span className="text-gray-500">{t('userCenter.payment.orderId') || '订单号'}</span>
              <span className="font-mono text-sm">{orderId}</span>
            </div>
            <div className="flex justify-between py-2 border-b">
              <span className="text-gray-500">{t('userCenter.payment.amount') || '金额'}</span>
              <span className="font-bold text-lg">${amount.toFixed(2)} USDT</span>
            </div>
            <div className="flex justify-between py-2 border-b">
              <span className="text-gray-500">{t('userCenter.payment.cnyAmount') || '约合人民币'}</span>
              <span>¥{(amount * 7).toFixed(2)}</span>
            </div>
            <div className="flex justify-between py-2 border-b">
              <span className="text-gray-500">{t('userCenter.payment.method') || '支付方式'}</span>
              <span className="uppercase">{method}</span>
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 text-sm text-yellow-800">
            {t('userCenter.payment.testMode') || '⚠️ 测试模式：点击确认后将直接到账'}
          </div>

          <div className="flex gap-3">
            <Link to="/user-center/billing" className="flex-1">
              <Button variant="outline" className="w-full">
                {t('common.cancel') || '取消'}
              </Button>
            </Link>
            <Button 
              className="flex-1" 
              onClick={handleConfirmPayment}
              disabled={loading}
            >
              {loading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
              {t('userCenter.payment.confirm') || '确认支付'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
