import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export function ModelsPage() {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('userCenter.models.title') || '可用模型'}</CardTitle>
        <CardDescription>{t('userCenter.models.description') || '您当前可使用的 AI 模型列表'}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {['GPT-4', 'GPT-3.5', 'Claude-3', 'Gemini Pro', 'Deepseek', 'Qwen'].map((model) => (
            <div key={model} className="p-4 bg-gray-50 rounded-lg">
              <p className="font-medium">{model}</p>
              <p className="text-sm text-gray-500">{t('userCenter.models.available') || '可用'}</p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
