'use client'

import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/components/ui/card'

interface AlertsSectionProps {
  data: any
}

export function AlertsSection({ data }: AlertsSectionProps) {
  if (!data) return null

  const { alerts } = data

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
      case 'high':
        return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
      case 'low':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
    }
  }

  const getSeverityText = (severity: string) => {
    switch (severity) {
      case 'critical':
        return '重大'
      case 'high':
        return '高'
      case 'medium':
        return '中'
      case 'low':
        return '低'
      default:
        return '不明'
    }
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))
    
    if (diffInMinutes < 1) return '今'
    if (diffInMinutes < 60) return `${diffInMinutes}分前`
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}時間前`
    return `${Math.floor(diffInMinutes / 1440)}日前`
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          アラート
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          システムアラートと通知の管理
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* アラート統計 */}
        <Card>
          <CardHeader>
            <CardTitle>アラート統計</CardTitle>
            <CardDescription>
              現在のアラート状況の概要
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">アクティブ</span>
                <span className="text-lg font-semibold text-red-600 dark:text-red-400">
                  {alerts.filter((alert: any) => alert.status === 'active').length}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">解決済み</span>
                <span className="text-lg font-semibold text-green-600 dark:text-green-400">
                  {alerts.filter((alert: any) => alert.status === 'resolved').length}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">総数</span>
                <span className="text-lg font-semibold text-gray-900 dark:text-white">
                  {alerts.length}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* アラート一覧 */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>最近のアラート</CardTitle>
              <CardDescription>
                最新のアラートとその状況
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {alerts.length === 0 ? (
                  <div className="text-center py-8">
                    <p className="text-gray-500 dark:text-gray-400">
                      アクティブなアラートはありません
                    </p>
                  </div>
                ) : (
                  alerts.map((alert: any) => (
                    <div
                      key={alert.id}
                      className="flex items-start space-x-4 p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
                    >
                      <div className="flex-shrink-0">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getSeverityColor(
                            alert.severity
                          )}`}
                        >
                          {getSeverityText(alert.severity)}
                        </span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 dark:text-white">
                          {alert.message}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          {formatTimestamp(alert.timestamp)}
                        </p>
                      </div>
                      <div className="flex-shrink-0">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            alert.status === 'active'
                              ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                              : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                          }`}
                        >
                          {alert.status === 'active' ? 'アクティブ' : '解決済み'}
                        </span>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
} 