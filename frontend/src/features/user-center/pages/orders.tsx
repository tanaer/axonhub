import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuthStore } from '@/stores/authStore';

interface Order {
  id: string;
  order_no: string;
  type: string;
  amount: number;
  status: string;
  payment_method?: string;
  created_at: string;
  paid_at?: string;
}

export function OrdersPage() {
  const { t } = useTranslation();
  const { auth } = useAuthStore();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const response = await fetch('/admin/orders?limit=20', {
        headers: { 'Authorization': `Bearer ${auth.accessToken}` },
      });
      if (response.ok) {
        const data = await response.json();
        setOrders(data.orders || []);
      }
    } catch (error) {
      console.error('Failed to fetch orders:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatAmount = (amount: number) => `¥${(amount / 100).toFixed(2)}`;
  const formatDate = (dateStr: string) => new Date(dateStr).toLocaleString('zh-CN');

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      pending: 'bg-yellow-100 text-yellow-800',
      paid: 'bg-green-100 text-green-800',
      failed: 'bg-red-100 text-red-800',
      cancelled: 'bg-gray-100 text-gray-800',
    };
    const labels: Record<string, string> = {
      pending: t('userCenter.orders.pending') || '待支付',
      paid: t('userCenter.orders.paid') || '已支付',
      failed: t('userCenter.orders.failed') || '支付失败',
      cancelled: t('userCenter.orders.cancelled') || '已取消',
    };
    return (
      <span className={`px-2 py-1 rounded text-xs ${styles[status] || styles.pending}`}>
        {labels[status] || status}
      </span>
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.orders.title') || '订单管理'}</CardTitle>
        <CardDescription>{t('userCenter.orders.description') || '查看您的订单历史'}</CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="space-y-2">
            {[...Array(5)].map((_, i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))}
          </div>
        ) : orders.length === 0 ? (
          <p className="text-center text-gray-500 py-8">{t('userCenter.orders.noOrders') || '暂无订单'}</p>
        ) : (
          <div className="space-y-2">
            {orders.map((order) => (
              <div key={order.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                <div>
                  <p className="font-medium">{order.order_no}</p>
                  <p className="text-sm text-gray-500">{formatDate(order.created_at)}</p>
                </div>
                <div className="text-right">
                  <p className="font-semibold">{formatAmount(order.amount)}</p>
                  {getStatusBadge(order.status)}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
