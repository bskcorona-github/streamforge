'use client'

import { Card } from '@/components/ui/card'
import { MetricCard } from '@/components/ui/metric-card'
import { SystemHealthIndicator } from '@/components/ui/system-health-indicator'

interface OverviewSectionProps {
  data: any
}

export function OverviewSection({ data }: OverviewSectionProps) {
  if (!data) return null

  const { overview, metrics } = data

  return (
    <div className="space-y-6">
      {/* ヘッダー */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          システムオーバービュー
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          システム全体の状況と主要メトリクス
        </p>
      </div>

      {/* 主要メトリクス */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard
          title="総メトリクス数"
          value={overview.totalMetrics.toLocaleString()}
          change="+12%"
          changeType="positive"
          icon="📊"
        />
        <MetricCard
          title="アクティブアラート"
          value={overview.activeAlerts}
          change="+2"
          changeType="negative"
          icon="🚨"
        />
        <MetricCard
          title="システム健全性"
          value={<SystemHealthIndicator status={overview.systemHealth} />}
          change=""
          changeType="neutral"
          icon="❤️"
        />
        <MetricCard
          title="稼働時間"
          value={`${overview.uptime}%`}
          change="+0.1%"
          changeType="positive"
          icon="⏱️"
        />
      </div>

      {/* システムメトリクス */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="card-header">
            <h3 className="card-title">リソース使用率</h3>
            <p className="card-description">
              リアルタイムのシステムリソース使用状況
            </p>
          </div>
          <div className="card-content">
            <div className="space-y-4">
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    CPU使用率
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {metrics.cpu}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${metrics.cpu}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    メモリ使用率
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {metrics.memory}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-green-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${metrics.memory}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    ディスク使用率
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {metrics.disk}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-yellow-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${metrics.disk}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    ネットワーク使用率
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {metrics.network}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-purple-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${metrics.network}%` }}
                  />
                </div>
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="card-header">
            <h3 className="card-title">最近のアクティビティ</h3>
            <p className="card-description">
              システムの最近のイベントとアクティビティ
            </p>
          </div>
          <div className="card-content">
            <div className="space-y-4">
              <div className="flex items-center space-x-3">
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    新しいメトリクスが収集されました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    2分前
                  </p>
                </div>
              </div>
              <div className="flex items-center space-x-3">
                <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    CPU使用率が警告レベルに達しました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    5分前
                  </p>
                </div>
              </div>
              <div className="flex items-center space-x-3">
                <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    システムバックアップが完了しました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    10分前
                  </p>
                </div>
              </div>
              <div className="flex items-center space-x-3">
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900 dark:text-white">
                    新しいサービスが起動しました
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    15分前
                  </p>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
} 