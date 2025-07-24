interface SystemHealthIndicatorProps {
  status: 'healthy' | 'warning' | 'critical'
  className?: string
}

export function SystemHealthIndicator({ status, className = '' }: SystemHealthIndicatorProps) {
  const getStatusConfig = () => {
    switch (status) {
      case 'healthy':
        return {
          color: 'text-green-600 dark:text-green-400',
          bgColor: 'bg-green-100 dark:bg-green-900',
          text: '健全',
          icon: '🟢'
        }
      case 'warning':
        return {
          color: 'text-yellow-600 dark:text-yellow-400',
          bgColor: 'bg-yellow-100 dark:bg-yellow-900',
          text: '警告',
          icon: '🟡'
        }
      case 'critical':
        return {
          color: 'text-red-600 dark:text-red-400',
          bgColor: 'bg-red-100 dark:bg-red-900',
          text: '危険',
          icon: '🔴'
        }
      default:
        return {
          color: 'text-gray-600 dark:text-gray-400',
          bgColor: 'bg-gray-100 dark:bg-gray-900',
          text: '不明',
          icon: '⚪'
        }
    }
  }

  const config = getStatusConfig()

  return (
    <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bgColor} ${config.color} ${className}`}>
      <span className="mr-1">{config.icon}</span>
      {config.text}
    </div>
  )
} 