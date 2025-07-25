'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Progress } from '@/components/ui/progress'
import { 
  Activity, 
  AlertTriangle, 
  BarChart3, 
  Cpu, 
  Database, 
  Globe, 
  Server, 
  Settings,
  TrendingUp,
  Users
} from 'lucide-react'

interface Metric {
  name: string
  value: number
  unit: string
  trend: 'up' | 'down' | 'stable'
  change: number
}

interface Alert {
  id: string
  severity: 'critical' | 'warning' | 'info'
  message: string
  timestamp: string
  service: string
}

interface Service {
  name: string
  status: 'healthy' | 'degraded' | 'down'
  uptime: number
  responseTime: number
  requestsPerSecond: number
}

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState('overview')
  const [metrics, setMetrics] = useState<Metric[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // モックデータの読み込み
    loadMockData()
  }, [])

  const loadMockData = () => {
    // メトリクスデータ
    const mockMetrics: Metric[] = [
      {
        name: 'CPU Usage',
        value: 45.2,
        unit: '%',
        trend: 'up',
        change: 2.1
      },
      {
        name: 'Memory Usage',
        value: 78.5,
        unit: '%',
        trend: 'stable',
        change: 0.0
      },
      {
        name: 'Network Throughput',
        value: 1250,
        unit: 'MB/s',
        trend: 'up',
        change: 15.3
      },
      {
        name: 'Active Connections',
        value: 1247,
        unit: '',
        trend: 'down',
        change: -5.2
      }
    ]

    // アラートデータ
    const mockAlerts: Alert[] = [
      {
        id: '1',
        severity: 'warning',
        message: 'High memory usage detected on collector-01',
        timestamp: '2024-01-15T10:30:00Z',
        service: 'collector'
      },
      {
        id: '2',
        severity: 'info',
        message: 'New ML model deployed successfully',
        timestamp: '2024-01-15T09:45:00Z',
        service: 'ml-engine'
      }
    ]

    // サービスデータ
    const mockServices: Service[] = [
      {
        name: 'API Gateway',
        status: 'healthy',
        uptime: 99.98,
        responseTime: 45,
        requestsPerSecond: 1250
      },
      {
        name: 'Collector',
        status: 'degraded',
        uptime: 99.85,
        responseTime: 120,
        requestsPerSecond: 890
      },
      {
        name: 'Stream Processor',
        status: 'healthy',
        uptime: 99.99,
        responseTime: 12,
        requestsPerSecond: 2100
      },
      {
        name: 'ML Engine',
        status: 'healthy',
        uptime: 99.92,
        responseTime: 85,
        requestsPerSecond: 450
      }
    ]

    setMetrics(mockMetrics)
    setAlerts(mockAlerts)
    setServices(mockServices)
    setIsLoading(false)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy': return 'bg-green-500'
      case 'degraded': return 'bg-yellow-500'
      case 'down': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-100 text-red-800 border-red-200'
      case 'warning': return 'bg-yellow-100 text-yellow-800 border-yellow-200'
      case 'info': return 'bg-blue-100 text-blue-800 border-blue-200'
      default: return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        {/* ヘッダー */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            StreamForge Dashboard
          </h1>
          <p className="text-gray-600">
            Real-time monitoring and analytics for distributed streaming systems
          </p>
        </div>

        {/* タブナビゲーション */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="overview" className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              Overview
            </TabsTrigger>
            <TabsTrigger value="metrics" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Metrics
            </TabsTrigger>
            <TabsTrigger value="services" className="flex items-center gap-2">
              <Server className="h-4 w-4" />
              Services
            </TabsTrigger>
            <TabsTrigger value="alerts" className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              Alerts
            </TabsTrigger>
            <TabsTrigger value="settings" className="flex items-center gap-2">
              <Settings className="h-4 w-4" />
              Settings
            </TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-6">
            {/* システム概要カード */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Services</CardTitle>
                  <Server className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{services.length}</div>
                  <p className="text-xs text-muted-foreground">
                    Active services
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Active Alerts</CardTitle>
                  <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{alerts.length}</div>
                  <p className="text-xs text-muted-foreground">
                    Current alerts
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Avg Response Time</CardTitle>
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {Math.round(services.reduce((acc, s) => acc + s.responseTime, 0) / services.length)}ms
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Across all services
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
                  <Globe className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {services.reduce((acc, s) => acc + s.requestsPerSecond, 0).toLocaleString()}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Requests per second
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* メトリクス概要 */}
            <Card>
              <CardHeader>
                <CardTitle>System Metrics</CardTitle>
                <CardDescription>
                  Real-time system performance metrics
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {metrics.map((metric, index) => (
                    <div key={index} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">{metric.name}</span>
                        <Badge variant={metric.trend === 'up' ? 'default' : metric.trend === 'down' ? 'destructive' : 'secondary'}>
                          {metric.trend === 'up' ? '+' : metric.trend === 'down' ? '-' : ''}{metric.change}%
                        </Badge>
                      </div>
                      <div className="flex items-center space-x-2">
                        <Progress value={metric.value} className="flex-1" />
                        <span className="text-sm font-mono">
                          {metric.value}{metric.unit}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Metrics Tab */}
          <TabsContent value="metrics" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Detailed Metrics</CardTitle>
                <CardDescription>
                  Comprehensive system metrics and analytics
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {metrics.map((metric, index) => (
                    <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center space-x-4">
                        <div className="p-2 bg-blue-100 rounded-lg">
                          <Activity className="h-4 w-4 text-blue-600" />
                        </div>
                        <div>
                          <h3 className="font-medium">{metric.name}</h3>
                          <p className="text-sm text-gray-500">
                            Current: {metric.value}{metric.unit}
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-lg font-bold">{metric.value}{metric.unit}</div>
                        <div className={`text-sm ${metric.trend === 'up' ? 'text-green-600' : metric.trend === 'down' ? 'text-red-600' : 'text-gray-600'}`}>
                          {metric.trend === 'up' ? '↗' : metric.trend === 'down' ? '↘' : '→'} {Math.abs(metric.change)}%
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Services Tab */}
          <TabsContent value="services" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Service Status</CardTitle>
                <CardDescription>
                  Real-time status of all StreamForge services
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {services.map((service, index) => (
                    <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center space-x-4">
                        <div className={`w-3 h-3 rounded-full ${getStatusColor(service.status)}`}></div>
                        <div>
                          <h3 className="font-medium">{service.name}</h3>
                          <p className="text-sm text-gray-500">
                            Uptime: {service.uptime}%
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-sm font-medium">{service.responseTime}ms</div>
                        <div className="text-sm text-gray-500">
                          {service.requestsPerSecond} req/s
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Alerts Tab */}
          <TabsContent value="alerts" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>System Alerts</CardTitle>
                <CardDescription>
                  Current alerts and notifications
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {alerts.length === 0 ? (
                    <div className="text-center py-8 text-gray-500">
                      No active alerts
                    </div>
                  ) : (
                    alerts.map((alert) => (
                      <Alert key={alert.id} className={getSeverityColor(alert.severity)}>
                        <AlertTriangle className="h-4 w-4" />
                        <AlertDescription>
                          <div className="flex items-center justify-between">
                            <span>{alert.message}</span>
                            <Badge variant="outline" className="ml-2">
                              {alert.service}
                            </Badge>
                          </div>
                          <div className="text-xs mt-1">
                            {new Date(alert.timestamp).toLocaleString()}
                          </div>
                        </AlertDescription>
                      </Alert>
                    ))
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Settings Tab */}
          <TabsContent value="settings" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Dashboard Settings</CardTitle>
                <CardDescription>
                  Configure dashboard preferences and notifications
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-medium">Auto-refresh</h3>
                      <p className="text-sm text-gray-500">
                        Automatically refresh dashboard data
                      </p>
                    </div>
                    <Button variant="outline">Configure</Button>
                  </div>
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-medium">Alert Notifications</h3>
                      <p className="text-sm text-gray-500">
                        Configure alert notification preferences
                      </p>
                    </div>
                    <Button variant="outline">Configure</Button>
                  </div>
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-medium">Theme</h3>
                      <p className="text-sm text-gray-500">
                        Choose dashboard theme
                      </p>
                    </div>
                    <Button variant="outline">Configure</Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
} 