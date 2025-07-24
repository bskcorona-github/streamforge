import { ReactNode } from 'react'

interface MetricCardProps {
  title: string
  value: ReactNode
  change?: string
  changeType?: 'positive' | 'negative' | 'neutral'
  icon?: string
  className?: string
}

export function MetricCard({ 
  title, 
  value, 
  change, 
  changeType = 'neutral', 
  icon,
  className = '' 
}: MetricCardProps) {
  const getChangeColor = () => {
    switch (changeType) {
      case 'positive':
        return 'text-green-600 dark:text-green-400'
      case 'negative':
        return 'text-red-600 dark:text-red-400'
      default:
        return 'text-gray-600 dark:text-gray-400'
    }
  }

  const getChangeIcon = () => {
    switch (changeType) {
      case 'positive':
        return '↗'
      case 'negative':
        return '↘'
      default:
        return ''
    }
  }

  return (
    <div className={`bg-white dark:bg-gray-800 shadow-sm border border-gray-200 dark:border-gray-700 rounded-lg p-6 ${className}`}>
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
            {title}
          </p>
          <div className="flex items-baseline">
            <p className="text-2xl font-semibold text-gray-900 dark:text-white">
              {value}
            </p>
            {change && (
              <span className={`ml-2 text-sm font-medium ${getChangeColor()}`}>
                {getChangeIcon()} {change}
              </span>
            )}
          </div>
        </div>
        {icon && (
          <div className="text-2xl">
            {icon}
          </div>
        )}
      </div>
    </div>
  )
} 