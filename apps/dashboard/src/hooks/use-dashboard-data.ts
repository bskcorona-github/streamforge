'use client'

import { useState, useEffect } from 'react'

interface DashboardData {
  overview: {
    totalMetrics: number
    activeAlerts: number
    systemHealth: 'healthy' | 'warning' | 'critical'
    uptime: number
  }
  metrics: {
    cpu: number
    memory: number
    disk: number
    network: number
  }
  alerts: Array<{
    id: string
    severity: 'low' | 'medium' | 'high' | 'critical'
    message: string
    timestamp: string
    status: 'active' | 'resolved'
  }>
  services: Array<{
    name: string
    status: 'running' | 'stopped' | 'error'
    uptime: number
    metrics: {
      requests: number
      errors: number
      latency: number
    }
  }>
}

export function useDashboardData() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = async () => {
    try {
      setIsLoading(true)
      setError(null)

      // 実際のAPIエンドポイントに置き換える
      const response = await fetch('/api/dashboard')
      
      if (!response.ok) {
        throw new Error('Failed to fetch dashboard data')
      }

      const result = await response.json()
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'))
      
      // 開発用のモックデータ
      setData({
        overview: {
          totalMetrics: 1250,
          activeAlerts: 3,
          systemHealth: 'healthy',
          uptime: 99.8,
        },
        metrics: {
          cpu: 45.2,
          memory: 67.8,
          disk: 23.1,
          network: 12.5,
        },
        alerts: [
          {
            id: '1',
            severity: 'medium',
            message: 'CPU使用率が80%を超えています',
            timestamp: new Date().toISOString(),
            status: 'active',
          },
          {
            id: '2',
            severity: 'low',
            message: 'メモリ使用率が警告レベルに達しています',
            timestamp: new Date(Date.now() - 300000).toISOString(),
            status: 'active',
          },
        ],
        services: [
          {
            name: 'API Gateway',
            status: 'running',
            uptime: 99.9,
            metrics: {
              requests: 1250,
              errors: 5,
              latency: 45,
            },
          },
          {
            name: 'Stream Processor',
            status: 'running',
            uptime: 99.8,
            metrics: {
              requests: 890,
              errors: 2,
              latency: 32,
            },
          },
          {
            name: 'ML Engine',
            status: 'running',
            uptime: 99.7,
            metrics: {
              requests: 456,
              errors: 1,
              latency: 78,
            },
          },
        ],
      })
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  return {
    data,
    isLoading,
    error,
    refetch: fetchData,
  }
} 