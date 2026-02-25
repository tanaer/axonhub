import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { useAuthStore } from '@/stores/authStore';

const presetAmounts = [
  { value: 1000, label: '¥10' },
  { value: 5000, label: '¥50' },
  { value: 10000, label: '¥100' },
  { value: 50000, label: '¥500' },
  { value: 100000, label: '¥1000' },
];

interface RechargeDialogProps {
  onSuccess?: () => void;
}

export function RechargeDialog({ onSuccess }: RechargeDialogProps) {
  const { t } = useTranslation();
  const { auth } = useAuthStore();
  const [open, setOpen] = useState(false);
  const [amount, setAmount] = useState(1000);
  const [paymentMethod, setPaymentMethod] = useState<'stripe' | 'epusdt'>('epusdt');
  const [loading, setLoading] = useState(false);

  const handleRecharge = async () => {
    setLoading(true);
    try {
      const response = await fetch('/admin/quota/recharge', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.accessToken}`,
        },
        body: JSON.stringify({
          amount,
          payment_method: paymentMethod,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        if (paymentMethod === 'epusdt' && data.payment_url) {
          // 跳转到 EPUSDT 支付页面
          window.open(data.payment_url, '_blank');
        } else {
          // Stripe 支付
          alert('充值请求已提交，请完成支付');
        }
        setOpen(false);
        onSuccess?.();
      } else {
        alert('充值失败，请重试');
      }
    } catch (error) {
      console.error('Recharge error:', error);
      alert('充值失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="bg-indigo-600 hover:bg-indigo-700">立即充值</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>账户充值</DialogTitle>
          <DialogDescription>
            选择充值金额，支持多种支付方式
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label>选择金额</Label>
            <div className="grid grid-cols-3 gap-2">
              {presetAmounts.map((preset) => (
                <Button
                  key={preset.value}
                  variant={amount === preset.value ? 'default' : 'outline'}
                  className={cn(
                    'h-12 text-lg',
                    amount === preset.value && 'bg-indigo-600 hover:bg-indigo-700'
                  )}
                  onClick={() => setAmount(preset.value)}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="custom-amount">自定义金额 (元)</Label>
            <Input
              id="custom-amount"
              type="number"
              placeholder="输入充值金额"
              value={amount / 100}
              onChange={(e) => setAmount(Math.round(parseFloat(e.target.value) * 100))}
              min={1}
            />
          </div>
          <div className="grid gap-2">
            <Label>支付方式</Label>
            <div className="grid grid-cols-2 gap-2">
              <Button
                variant={paymentMethod === 'epusdt' ? 'default' : 'outline'}
                className={cn(
                  'h-12',
                  paymentMethod === 'epusdt' && 'bg-indigo-600 hover:bg-indigo-700'
                )}
                onClick={() => setPaymentMethod('epusdt')}
              >
                💰 USDT
              </Button>
              <Button
                variant={paymentMethod === 'stripe' ? 'default' : 'outline'}
                className={cn(
                  'h-12',
                  paymentMethod === 'stripe' && 'bg-indigo-600 hover:bg-indigo-700'
                )}
                onClick={() => setPaymentMethod('stripe')}
              >
                💳 信用卡
              </Button>
            </div>
          </div>
          <div className="rounded-lg bg-gray-50 p-3 text-sm text-gray-600">
            <p>充值金额: <span className="font-bold text-indigo-600">¥{(amount / 100).toFixed(2)}</span></p>
            <p className="text-xs mt-1">充值后余额将立即到账</p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
          <Button 
            onClick={handleRecharge} 
            disabled={loading || amount < 100}
            className="bg-indigo-600 hover:bg-indigo-700"
          >
            {loading ? '处理中...' : `充值 ¥${(amount / 100).toFixed(2)}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
