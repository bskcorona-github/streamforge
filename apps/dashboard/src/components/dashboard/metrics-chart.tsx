'use client'

import { useEffect, useRef, useState } from 'react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ChartOptions,
  ChartData,
} from 'chart.js'
import { Line } from 'react-chartjs-2'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

interface MetricsChartProps {
  title: string
  data: {
    labels: string[]
    datasets: {
      label: string
      data: number[]
      borderColor?: string
      backgroundColor?: string
    }[]
  }
  height?: number
  showLegend?: boolean
  yAxisLabel?: string
  xAxisLabel?: string
}

export function MetricsChart({
  title,
  data,
  height = 300,
  showLegend = true,
  yAxisLabel,
  xAxisLabel,
}: MetricsChartProps) {
  const chartRef = useRef<ChartJS>(null)
  const [chartData, setChartData] = useState<ChartData<'line'>>(data)

  useEffect(() => {
    setChartData(data)
  }, [data])

  const options: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: showLegend,
        position: 'top' as const,
      },
      title: {
        display: false,
      },
      tooltip: {
        mode: 'index' as const,
        intersect: false,
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        titleColor: '#fff',
        bodyColor: '#fff',
        borderColor: '#374151',
        borderWidth: 1,
      },
    },
    scales: {
      x: {
        display: true,
        title: {
          display: !!xAxisLabel,
          text: xAxisLabel,
          color: '#6b7280',
        },
        grid: {
          color: '#374151',
          drawBorder: false,
        },
        ticks: {
          color: '#9ca3af',
        },
      },
      y: {
        display: true,
        title: {
          display: !!yAxisLabel,
          text: yAxisLabel,
          color: '#6b7280',
        },
        grid: {
          color: '#374151',
          drawBorder: false,
        },
        ticks: {
          color: '#9ca3af',
        },
      },
    },
    interaction: {
      mode: 'nearest' as const,
      axis: 'x' as const,
      intersect: false,
    },
  }

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-gray-900 dark:text-white">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ height }}>
          <Line ref={chartRef} data={chartData} options={options} />
        </div>
      </CardContent>
    </Card>
  )
}

// リアルタイムメトリクスチャート
interface RealtimeMetricsChartProps {
  title: string
  metricName: string
  updateInterval?: number
  maxDataPoints?: number
}

export function RealtimeMetricsChart({
  title,
  metricName,
  updateInterval = 5000,
  maxDataPoints = 50,
}: RealtimeMetricsChartProps) {
  const [data, setData] = useState<ChartData<'line'>>({
    labels: [],
    datasets: [
      {
        label: metricName,
        data: [],
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.4,
        fill: true,
      },
    ],
  })

  useEffect(() => {
    const interval = setInterval(() => {
      // 実際のAPIからデータを取得する代わりに、シミュレーションデータを使用
      const now = new Date()
      const timestamp = now.toLocaleTimeString()
      const value = Math.random() * 100

      setData((prevData) => {
        const newLabels = [...prevData.labels, timestamp]
        const newData = [...prevData.datasets[0].data, value]

        // 最大データポイント数を超えた場合、古いデータを削除
        if (newLabels.length > maxDataPoints) {
          newLabels.shift()
          newData.shift()
        }

        return {
          labels: newLabels,
          datasets: [
            {
              ...prevData.datasets[0],
              data: newData,
            },
          ],
        }
      })
    }, updateInterval)

    return () => clearInterval(interval)
  }, [metricName, updateInterval, maxDataPoints])

  return (
    <MetricsChart
      title={title}
      data={data}
      height={250}
      yAxisLabel="Value"
      xAxisLabel="Time"
    />
  )
}

// CPU使用率チャート
export function CPUUsageChart() {
  const [data, setData] = useState<ChartData<'line'>>({
    labels: [],
    datasets: [
      {
        label: 'CPU Usage (%)',
        data: [],
        borderColor: 'rgb(239, 68, 68)',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        tension: 0.4,
        fill: true,
      },
    ],
  })

  useEffect(() => {
    const interval = setInterval(() => {
      const now = new Date()
      const timestamp = now.toLocaleTimeString()
      const value = 20 + Math.random() * 60 // 20-80%の範囲

      setData((prevData) => {
        const newLabels = [...prevData.labels, timestamp]
        const newData = [...prevData.datasets[0].data, value]

        if (newLabels.length > 30) {
          newLabels.shift()
          newData.shift()
        }

        return {
          labels: newLabels,
          datasets: [
            {
              ...prevData.datasets[0],
              data: newData,
            },
          ],
        }
      })
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return (
    <MetricsChart
      title="CPU使用率"
      data={data}
      height={250}
      yAxisLabel="Usage (%)"
      xAxisLabel="Time"
    />
  )
}

// メモリ使用率チャート
export function MemoryUsageChart() {
  const [data, setData] = useState<ChartData<'line'>>({
    labels: [],
    datasets: [
      {
        label: 'Memory Usage (%)',
        data: [],
        borderColor: 'rgb(16, 185, 129)',
        backgroundColor: 'rgba(16, 185, 129, 0.1)',
        tension: 0.4,
        fill: true,
      },
    ],
  })

  useEffect(() => {
    const interval = setInterval(() => {
      const now = new Date()
      const timestamp = now.toLocaleTimeString()
      const value = 30 + Math.random() * 40 // 30-70%の範囲

      setData((prevData) => {
        const newLabels = [...prevData.labels, timestamp]
        const newData = [...prevData.datasets[0].data, value]

        if (newLabels.length > 30) {
          newLabels.shift()
          newData.shift()
        }

        return {
          labels: newLabels,
          datasets: [
            {
              ...prevData.datasets[0],
              data: newData,
            },
          ],
        }
      })
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return (
    <MetricsChart
      title="メモリ使用率"
      data={data}
      height={250}
      yAxisLabel="Usage (%)"
      xAxisLabel="Time"
    />
  )
}

// ネットワークトラフィックチャート
export function NetworkTrafficChart() {
  const [data, setData] = useState<ChartData<'line'>>({
    labels: [],
    datasets: [
      {
        label: 'Incoming (MB/s)',
        data: [],
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.4,
        fill: false,
      },
      {
        label: 'Outgoing (MB/s)',
        data: [],
        borderColor: 'rgb(245, 158, 11)',
        backgroundColor: 'rgba(245, 158, 11, 0.1)',
        tension: 0.4,
        fill: false,
      },
    ],
  })

  useEffect(() => {
    const interval = setInterval(() => {
      const now = new Date()
      const timestamp = now.toLocaleTimeString()
      const incoming = Math.random() * 50
      const outgoing = Math.random() * 30

      setData((prevData) => {
        const newLabels = [...prevData.labels, timestamp]
        const newIncoming = [...prevData.datasets[0].data, incoming]
        const newOutgoing = [...prevData.datasets[1].data, outgoing]

        if (newLabels.length > 30) {
          newLabels.shift()
          newIncoming.shift()
          newOutgoing.shift()
        }

        return {
          labels: newLabels,
          datasets: [
            {
              ...prevData.datasets[0],
              data: newIncoming,
            },
            {
              ...prevData.datasets[1],
              data: newOutgoing,
            },
          ],
        }
      })
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return (
    <MetricsChart
      title="ネットワークトラフィック"
      data={data}
      height={250}
      yAxisLabel="Traffic (MB/s)"
      xAxisLabel="Time"
    />
  )
}

// ディスクI/Oチャート
export function DiskIOChart() {
  const [data, setData] = useState<ChartData<'line'>>({
    labels: [],
    datasets: [
      {
        label: 'Read (MB/s)',
        data: [],
        borderColor: 'rgb(139, 92, 246)',
        backgroundColor: 'rgba(139, 92, 246, 0.1)',
        tension: 0.4,
        fill: false,
      },
      {
        label: 'Write (MB/s)',
        data: [],
        borderColor: 'rgb(236, 72, 153)',
        backgroundColor: 'rgba(236, 72, 153, 0.1)',
        tension: 0.4,
        fill: false,
      },
    ],
  })

  useEffect(() => {
    const interval = setInterval(() => {
      const now = new Date()
      const timestamp = now.toLocaleTimeString()
      const read = Math.random() * 20
      const write = Math.random() * 15

      setData((prevData) => {
        const newLabels = [...prevData.labels, timestamp]
        const newRead = [...prevData.datasets[0].data, read]
        const newWrite = [...prevData.datasets[1].data, write]

        if (newLabels.length > 30) {
          newLabels.shift()
          newRead.shift()
          newWrite.shift()
        }

        return {
          labels: newLabels,
          datasets: [
            {
              ...prevData.datasets[0],
              data: newRead,
            },
            {
              ...prevData.datasets[1],
              data: newWrite,
            },
          ],
        }
      })
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return (
    <MetricsChart
      title="ディスクI/O"
      data={data}
      height={250}
      yAxisLabel="I/O (MB/s)"
      xAxisLabel="Time"
    />
  )
} 