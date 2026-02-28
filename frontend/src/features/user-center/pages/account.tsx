import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';
import { Eye, EyeOff, KeyRound } from 'lucide-react';

export function AccountPage() {
  const { t } = useTranslation();
  const { auth } = useAuthStore();
  const user = auth.user;

  const [firstName, setFirstName] = useState(user?.firstName || '');
  const [lastName, setLastName] = useState(user?.lastName || '');
  const [email] = useState(user?.email || '');
  const [loading, setLoading] = useState(false);

  // Password change state
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);

  const handleSave = async () => {
    setLoading(true);
    try {
      const response = await fetch('/admin/user/profile', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.accessToken}`,
        },
        body: JSON.stringify({ firstName, lastName }),
      });

      if (response.ok) {
        toast.success(t('common.saved') || '已保存');
      } else {
        toast.error(t('common.saveFailed') || '保存失败');
      }
    } catch (error) {
      toast.error(t('common.saveFailed') || '保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async () => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      toast.error(t('userCenter.account.fillAllFields') || '请填写所有字段');
      return;
    }

    if (newPassword !== confirmPassword) {
      toast.error(t('userCenter.account.passwordMismatch') || '两次输入的密码不一致');
      return;
    }

    if (newPassword.length < 7) {
      toast.error(t('userCenter.account.passwordTooShort') || '密码长度至少7位');
      return;
    }

    setPasswordLoading(true);
    try {
      const response = await fetch('/admin/user/password', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.accessToken}`,
        },
        body: JSON.stringify({
          currentPassword,
          newPassword,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        toast.success(t('userCenter.account.passwordChanged') || '密码修改成功');
        setCurrentPassword('');
        setNewPassword('');
        setConfirmPassword('');
      } else {
        toast.error(data.error?.message || t('userCenter.account.passwordChangeFailed') || '密码修改失败');
      }
    } catch (error) {
      toast.error(t('userCenter.account.passwordChangeFailed') || '密码修改失败');
    } finally {
      setPasswordLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t('userCenter.account.title') || '账户信息'}</CardTitle>
          <CardDescription>{t('userCenter.account.description') || '管理您的个人资料'}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="firstName">{t('userCenter.account.firstName') || '名'}</Label>
              <Input
                id="firstName"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                placeholder={t('userCenter.account.firstNamePlaceholder') || '请输入名'}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="lastName">{t('userCenter.account.lastName') || '姓'}</Label>
              <Input
                id="lastName"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                placeholder={t('userCenter.account.lastNamePlaceholder') || '请输入姓'}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="email">{t('userCenter.account.email') || '邮箱'}</Label>
            <Input id="email" value={email} disabled className="bg-gray-50" />
            <p className="text-sm text-gray-500">{t('userCenter.account.emailDisabled') || '邮箱不可修改'}</p>
          </div>
          <Button onClick={handleSave} disabled={loading}>
            {loading ? t('common.saving') || '保存中...' : t('common.save') || '保存'}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <KeyRound className="w-5 h-5" />
            <CardTitle>{t('userCenter.account.changePassword') || '修改密码'}</CardTitle>
          </div>
          <CardDescription>{t('userCenter.account.changePasswordDesc') || '输入当前密码和新密码来修改您的密码'}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="currentPassword">{t('userCenter.account.currentPassword') || '当前密码'}</Label>
            <div className="relative">
              <Input
                id="currentPassword"
                type={showCurrentPassword ? 'text' : 'password'}
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder={t('userCenter.account.currentPasswordPlaceholder') || '请输入当前密码'}
              />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                onClick={() => setShowCurrentPassword(!showCurrentPassword)}
              >
                {showCurrentPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="newPassword">{t('userCenter.account.newPassword') || '新密码'}</Label>
            <div className="relative">
              <Input
                id="newPassword"
                type={showNewPassword ? 'text' : 'password'}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder={t('userCenter.account.newPasswordPlaceholder') || '请输入新密码（至少7位）'}
              />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                onClick={() => setShowNewPassword(!showNewPassword)}
              >
                {showNewPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmPassword">{t('userCenter.account.confirmPassword') || '确认新密码'}</Label>
            <Input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder={t('userCenter.account.confirmPasswordPlaceholder') || '请再次输入新密码'}
            />
          </div>
          <Button onClick={handleChangePassword} disabled={passwordLoading}>
            {passwordLoading ? t('common.saving') || '保存中...' : t('userCenter.account.updatePassword') || '更新密码'}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
