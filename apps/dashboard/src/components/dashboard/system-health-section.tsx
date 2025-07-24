'use client'

import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/components/ui/card'
import { SystemHealthIndicator } from '@/components/ui/system-health-indicator'

interface SystemHealthSectionProps {
  data: any
}

export function SystemHealthSection({ data }: SystemHealthSectionProps) {
  if (!data) return null

  const { overview } = data

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          システム健全性
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          システム全体の健全性とパフォーマンス指標
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* システム健全性概要 */}
        <Card>
          <CardHeader>
            <CardTitle>システム健全性概要</CardTitle>
            <CardDescription>
              現在のシステム状況の詳細
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  全体的な健全性
                </span>
                <SystemHealthIndicator status={overview.systemHealth} />
              </div>
              
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  稼働時間
                </span>
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {overview.uptime}%
                </span>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  総メトリクス数
                </span>
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {overview.totalMetrics.toLocaleString()}
                </span>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  アクティブアラート
                </span>
                <span className="text-sm font-medium text-red-600 dark:text-red-400">
                  {overview.activeAlerts}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 健全性チェックリスト */}
        <Card>
          <CardHeader>
            <CardTitle>健全性チェックリスト</CardTitle>
            <CardDescription>
              システムコンポーネントの健全性確認
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    API Gateway
                  </span>
                </div>
                <span className="text-sm text-green-600 dark:text-green-400">
                  正常
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    Stream Processor
                  </span>
                </div>
                <span className="text-sm text-green-600 dark:text-green-400">
                  正常
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    ML Engine
                  </span>
                </div>
                <span className="text-sm text-yellow-600 dark:text-yellow-400">
                  警告
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    Database
                  </span>
                </div>
                <span className="text-sm text-green-600 dark:text-green-400">
                  正常
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    Cache
                  </span>
                </div>
                <span className="text-sm text-green-600 dark:text-green-400">
                  正常
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    Message Queue
                  </span>
                </div>
                <span className="text-sm text-green-600 dark:text-green-400">
                  正常
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* パフォーマンス指標 */}
        <Card>
          <CardHeader>
            <CardTitle>パフォーマンス指標</CardTitle>
            <CardDescription>
              システムパフォーマンスの主要指標
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    平均応答時間
                  </span>
                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                    45ms
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div className="bg-green-600 h-2 rounded-full" style={{ width: '75%' }} />
                </div>
              </div>

              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    スループット
                  </span>
                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                    1,250 req/s
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div className="bg-blue-600 h-2 rounded-full" style={{ width: '85%' }} />
                </div>
              </div>

              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    エラー率
                  </span>
                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                    0.12%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div className="bg-green-600 h-2 rounded-full" style={{ width: '12%' }} />
                </div>
              </div>

              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    可用性
                  </span>
                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                    99.95%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div className="bg-green-600 h-2 rounded-full" style={{ width: '99.95%' }} />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 最近のイベント */}
        <Card>
          <CardHeader>
            <CardTitle>最近のイベント</CardTitle>
            <CardDescription>
              システムの最近のイベントとログ
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-green-500 rounded-full mt-2"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    システムバックアップが正常に完了しました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    2時間前
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-blue-500 rounded-full mt-2"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    新しいメトリクスコレクターが接続されました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    4時間前
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-yellow-500 rounded-full mt-2"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    ML Engineのメモリ使用率が警告レベルに達しました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    6時間前
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3">
                <div className="w-2 h-2 bg-green-500 rounded-full mt-2"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    システムアップデートが正常に完了しました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    1日前
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
} 