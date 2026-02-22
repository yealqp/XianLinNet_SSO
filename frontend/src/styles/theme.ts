// Emotion 主题配置
export const theme = {
  // 颜色系统
  colors: {
    // 主色调
    primary: '#ec4899',
    primaryLight: '#f472b6',
    primaryDark: '#db2777',
    
    // 辅助色
    success: '#10b981',
    successLight: '#34d399',
    warning: '#f59e0b',
    warningLight: '#fbbf24',
    info: '#3b82f6',
    infoLight: '#60a5fa',
    error: '#ef4444',
    errorLight: '#f87171',
    
    // 中性色
    white: '#ffffff',
    black: '#000000',
    gray50: '#fafafa',
    gray100: '#f3f4f6',
    gray200: '#e5e7eb',
    gray300: '#d1d5db',
    gray400: '#9ca3af',
    gray500: '#6b7280',
    gray600: '#4b5563',
    gray700: '#374151',
    gray800: '#1f2937',
    gray900: '#111827',
    
    // 语义色
    textPrimary: '#1f1f1f',
    textSecondary: '#666666',
    textTertiary: '#999999',
    border: '#e5e7eb',
    borderLight: '#f3f4f6',
    bgBase: '#ffffff',
    bgElevated: '#ffffff',
    
    // 通知色
    noticeInfoBg: '#f0f9ff',
    noticeInfoBorder: '#bae6fd',
    noticeInfoIcon: '#0284c7',
    noticeWarningBg: '#fef3f2',
    noticeWarningBorder: '#fecaca',
    noticeWarningIcon: '#dc2626',
  },
  
  // 圆角
  radius: {
    xs: '4px',
    sm: '6px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    '2xl': '20px',
    '3xl': '24px',
    full: '9999px',
  },
  
  // 阴影
  shadows: {
    none: 'none',
    sm: '0 1px 2px rgba(0, 0, 0, 0.04)',
    md: '0 2px 8px rgba(0, 0, 0, 0.08)',
    lg: '0 4px 12px rgba(0, 0, 0, 0.08)',
    xl: '0 8px 24px rgba(0, 0, 0, 0.12)',
  },
  
  // 间距
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '12px',
    lg: '16px',
    xl: '20px',
    '2xl': '24px',
    '3xl': '32px',
    '4xl': '40px',
    '5xl': '48px',
  },
  
  // 字体大小
  fontSize: {
    xs: '12px',
    sm: '13px',
    base: '14px',
    lg: '15px',
    xl: '16px',
    '2xl': '18px',
    '3xl': '20px',
    '4xl': '24px',
    '5xl': '28px',
    '6xl': '32px',
  },
  
  // 字重
  fontWeight: {
    normal: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
  },
  
  // 过渡
  transition: {
    fast: '0.15s ease',
    base: '0.2s ease',
    slow: '0.3s ease',
  },
  
  // 断点
  breakpoints: {
    xs: '480px',
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
    '2xl': '1536px',
  },
}

export type Theme = typeof theme

// 媒体查询辅助函数
export const mq = {
  xs: `@media (max-width: ${theme.breakpoints.xs})`,
  sm: `@media (max-width: ${theme.breakpoints.sm})`,
  md: `@media (max-width: ${theme.breakpoints.md})`,
  lg: `@media (max-width: ${theme.breakpoints.lg})`,
  xl: `@media (max-width: ${theme.breakpoints.xl})`,
  '2xl': `@media (max-width: ${theme.breakpoints['2xl']})`,
}

// 渐变色辅助函数
export const gradients = {
  primary: `linear-gradient(135deg, ${theme.colors.primary} 0%, ${theme.colors.primaryLight} 100%)`,
  success: `linear-gradient(135deg, ${theme.colors.success} 0%, ${theme.colors.successLight} 100%)`,
  warning: `linear-gradient(135deg, ${theme.colors.warning} 0%, ${theme.colors.warningLight} 100%)`,
  info: `linear-gradient(135deg, ${theme.colors.info} 0%, ${theme.colors.infoLight} 100%)`,
}
