'use client'

import { useState, useEffect } from 'react'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { OverviewSection } from '@/components/dashboard/overview-section'
import { MetricsSection } from '@/components/dashboard/metrics-section'
import { AlertsSection } from '@/components/dashboard/alerts-section'
import { ServicesSection } from '@/components/dashboard/services-section'
import { SystemHealthSection } from '@/components/dashboard/system-health-section'
import { useDashboardData } from '@/hooks/use-dashboard-data'
import { LoadingSpinner } from '@/components/ui/loading-spinner'
import { ErrorBoundary } from '@/components/ui/error-boundary'

export default function DashboardPage() {
  const [activeTab, setActiveTab] = useState('overview')
  const { data, isLoading, error, refetch } = useDashboardData()

  useEffect(() => {
    // リアルタイム更新のためのポーリング
    const interval = setInterval(() => {
      refetch()
    }, 30000) // 30秒ごとに更新

    return () => clearInterval(interval)
  }, [refetch])

  if (isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-screen">
          <LoadingSpinner size="lg" />
        </div>
      </DashboardLayout>
    )
  }

  if (error) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-screen">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-red-600 mb-4">エラーが発生しました</h2>
            <p className="text-gray-600 mb-4">{error.message}</p>
            <button
              onClick={() => refetch()}
              className="btn btn-primary"
            >
              再試行
            </button>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  const tabs = [
    { id: 'overview', label: 'オーバービュー', icon: '📊' },
    { id: 'metrics', label: 'メトリクス', icon: '📈' },
    { id: 'alerts', label: 'アラート', icon: '🚨' },
    { id: 'services', label: 'サービス', icon: '🔧' },
    { id: 'health', label: 'システム健全性', icon: '❤️' },
  ]

  return (
    <DashboardLayout>
      <ErrorBoundary>
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
          {/* ヘッダー */}
          <div className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="flex justify-between items-center py-4">
                <div>
                  <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                    StreamForge ダッシュボード
                  </h1>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    リアルタイム監視と分析
                  </p>
                </div>
                <div className="flex items-center space-x-4">
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse-green"></div>
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      リアルタイム
                    </span>
                  </div>
                  <button
                    onClick={() => refetch()}
                    className="btn btn-outline btn-sm"
                  >
                    更新
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* タブナビゲーション */}
          <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <nav className="flex space-x-8">
                {tabs.map((tab) => (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                      activeTab === tab.id
                        ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                        : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                    }`}
                  >
                    <span className="mr-2">{tab.icon}</span>
                    {tab.label}
                  </button>
                ))}
              </nav>
            </div>
          </div>

          {/* メインコンテンツ */}
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            {activeTab === 'overview' && (
              <OverviewSection data={data} />
            )}
            {activeTab === 'metrics' && (
              <MetricsSection data={data} />
            )}
            {activeTab === 'alerts' && (
              <AlertsSection data={data} />
            )}
            {activeTab === 'services' && (
              <ServicesSection data={data} />
            )}
            {activeTab === 'health' && (
              <SystemHealthSection data={data} />
            )}
          </div>
        </div>
      </ErrorBoundary>
    </DashboardLayout>
  )
} 