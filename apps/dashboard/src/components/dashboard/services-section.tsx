'use client'

import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/components/ui/card'

interface ServicesSectionProps {
  data: any
}

export function ServicesSection({ data }: ServicesSectionProps) {
  if (!data) return null

  const { services } = data

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
      case 'stopped':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
      case 'error':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'running':
        return '実行中'
      case 'stopped':
        return '停止'
      case 'error':
        return 'エラー'
      default:
        return '不明'
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          サービス
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          マイクロサービスの状況とパフォーマンス
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {services.map((service: any) => (
          <Card key={service.name}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">{service.name}</CardTitle>
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(
                    service.status
                  )}`}
                >
                  {getStatusText(service.status)}
                </span>
              </div>
              <CardDescription>
                稼働時間: {service.uptime}%
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      リクエスト数
                    </span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {service.metrics.requests.toLocaleString()}
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full"
                      style={{
                        width: `${Math.min((service.metrics.requests / 1000) * 100, 100)}%`,
                      }}
                    />
                  </div>
                </div>

                <div>
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      エラー率
                    </span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {((service.metrics.errors / service.metrics.requests) * 100).toFixed(2)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-red-600 h-2 rounded-full"
                      style={{
                        width: `${Math.min((service.metrics.errors / service.metrics.requests) * 100, 100)}%`,
                      }}
                    />
                  </div>
                </div>

                <div>
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      平均レイテンシー
                    </span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {service.metrics.latency}ms
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-green-600 h-2 rounded-full"
                      style={{
                        width: `${Math.min((service.metrics.latency / 100) * 100, 100)}%`,
                      }}
                    />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* サービス概要 */}
      <Card>
        <CardHeader>
          <CardTitle>サービス概要</CardTitle>
          <CardDescription>
            全サービスの統合メトリクス
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                {services.length}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">
                総サービス数
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                {services.filter((s: any) => s.status === 'running').length}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">
                実行中
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                {services.filter((s: any) => s.status === 'stopped').length}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">
                停止中
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
                {services.filter((s: any) => s.status === 'error').length}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">
                エラー
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
} 