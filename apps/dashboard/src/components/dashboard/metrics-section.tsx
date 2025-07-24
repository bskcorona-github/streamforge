'use client'

import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/components/ui/card'

interface MetricsSectionProps {
  data: any
}

export function MetricsSection({ data }: MetricsSectionProps) {
  if (!data) return null

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          メトリクス
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          システムメトリクスの詳細な分析とグラフ
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>CPU使用率の推移</CardTitle>
            <CardDescription>
              過去24時間のCPU使用率の変化
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-900 rounded-lg">
              <p className="text-gray-500 dark:text-gray-400">
                グラフコンポーネントがここに表示されます
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>メモリ使用率の推移</CardTitle>
            <CardDescription>
              過去24時間のメモリ使用率の変化
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-900 rounded-lg">
              <p className="text-gray-500 dark:text-gray-400">
                グラフコンポーネントがここに表示されます
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>ネットワークトラフィック</CardTitle>
            <CardDescription>
              入出力トラフィックの監視
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-900 rounded-lg">
              <p className="text-gray-500 dark:text-gray-400">
                グラフコンポーネントがここに表示されます
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>ディスクI/O</CardTitle>
            <CardDescription>
              ディスクの読み書き速度
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-900 rounded-lg">
              <p className="text-gray-500 dark:text-gray-400">
                グラフコンポーネントがここに表示されます
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
} 