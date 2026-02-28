import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';
import { Copy, Trash2, Plus, Eye, EyeOff } from 'lucide-react';

interface APIKey {
  id: string;
  name: string;
  key: string;
  prefix: string;
  created_at: string;
  last_used_at?: string;
  is_active: boolean;
}

export function APIKeysPage() {
  const { t } = useTranslation();
  const { auth } = useAuthStore();
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set());

  useEffect(() => {
    fetchAPIKeys();
  }, []);

  const fetchAPIKeys = async () => {
    setLoading(true);
    try {
      const response = await fetch('/admin/api-keys', {
        headers: { 'Authorization': `Bearer ${auth.accessToken}` },
      });
      if (response.ok) {
        setApiKeys(await response.json());
      }
    } catch (error) {
      console.error('Failed to fetch API keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!newKeyName.trim()) {
      toast.error(t('userCenter.apiKeys.nameRequired') || '请输入名称');
      return;
    }

    try {
      const response = await fetch('/admin/api-keys', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.accessToken}`,
        },
        body: JSON.stringify({ name: newKeyName }),
      });

      if (response.ok) {
        const newKey = await response.json();
        setApiKeys([...apiKeys, newKey]);
        setNewKeyName('');
        setShowCreateModal(false);
        toast.success(t('userCenter.apiKeys.created') || 'API Key 创建成功');
        // 显示新创建的 key
        setVisibleKeys(new Set([...visibleKeys, newKey.id]));
      } else {
        toast.error(t('userCenter.apiKeys.createFailed') || '创建失败');
      }
    } catch (error) {
      toast.error(t('userCenter.apiKeys.createFailed') || '创建失败');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t('userCenter.apiKeys.confirmDelete') || '确定要删除此 API Key 吗？')) return;

    try {
      const response = await fetch(`/admin/api-keys/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${auth.accessToken}` },
      });

      if (response.ok) {
        setApiKeys(apiKeys.filter(k => k.id !== id));
        toast.success(t('userCenter.apiKeys.deleted') || '已删除');
      }
    } catch (error) {
      toast.error(t('userCenter.apiKeys.deleteFailed') || '删除失败');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success(t('userCenter.apiKeys.copied') || '已复制到剪贴板');
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t('userCenter.apiKeys.title') || 'API Keys'}</CardTitle>
              <CardDescription>{t('userCenter.apiKeys.description') || '管理您的 API 密钥'}</CardDescription>
            </div>
            <Button onClick={() => setShowCreateModal(true)}>
              <Plus className="w-4 h-4 mr-2" />
              {t('userCenter.apiKeys.create') || '创建'}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="space-y-2">
              {[...Array(3)].map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : apiKeys.length === 0 ? (
            <p className="text-center text-gray-500 py-8">{t('userCenter.apiKeys.noKeys') || '暂无 API Key'}</p>
          ) : (
            <div className="space-y-2">
              {apiKeys.map((key) => (
                <div key={key.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div className="flex-1">
                    <p className="font-medium">{key.name}</p>
                    <div className="flex items-center gap-2 mt-1">
                      <code className="text-sm bg-gray-200 px-2 py-1 rounded">
                        {visibleKeys.has(key.id) ? key.key : key.prefix + '****************'}
                      </code>
                      <button onClick={() => setVisibleKeys(new Set(visibleKeys.has(key.id) ? [] : [key.id]))}>
                        {visibleKeys.has(key.id) ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                      <button onClick={() => copyToClipboard(key.key)}>
                        <Copy className="w-4 h-4" />
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                      {t('userCenter.apiKeys.createdAt') || '创建于'}: {formatDate(key.created_at)}
                      {key.last_used_at && ` | ${t('userCenter.apiKeys.lastUsed') || '最后使用'}: ${formatDate(key.last_used_at)}`}
                    </p>
                  </div>
                  <Button variant="ghost" size="icon" onClick={() => handleDelete(key.id)}>
                    <Trash2 className="w-4 h-4 text-red-500" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 创建弹窗 */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>{t('userCenter.apiKeys.createTitle') || '创建 API Key'}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">{t('userCenter.apiKeys.keyName') || '名称'}</label>
                <Input
                  value={newKeyName}
                  onChange={(e) => setNewKeyName(e.target.value)}
                  placeholder={t('userCenter.apiKeys.keyNamePlaceholder') || '例如：生产环境'}
                />
              </div>
              <div className="flex gap-2 justify-end">
                <Button variant="outline" onClick={() => setShowCreateModal(false)}>
                  {t('common.cancel') || '取消'}
                </Button>
                <Button onClick={handleCreate}>
                  {t('common.create') || '创建'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
