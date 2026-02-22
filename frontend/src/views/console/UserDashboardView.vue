<template>
  <div :class="styles.container">
    <!-- 页面标题 -->
    <div :class="styles.pageHeader">
      <h1 :class="styles.pageTitle">首页</h1>
    </div>

    <!-- 通知卡片 - 仅显示未实名认证提示 -->
    <div :class="styles.noticeCards" v-if="!userInfo?.isRealName">
      <div :class="[styles.noticeCard, styles.noticeWarning]">
        <div :class="[styles.noticeIcon, styles.noticeIconWarning]">
          <ExclamationCircleOutlined />
        </div>
        <div :class="styles.noticeContent">
          <div :class="styles.noticeTitle">注意！</div>
          <div :class="styles.noticeText">实名认证后可使用所有功能，请尽快完成实名认证</div>
        </div>
        <div :class="styles.noticeAction">
          <a-button type="link" size="small" @click="$router.push('/console/realname')">去认证</a-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片网格 -->
    <div :class="styles.statsGrid">
      <div :class="styles.statCard">
        <div :class="styles.statIconWrapper">
          <div :class="[styles.statIcon, styles.statIconPrimary]">
            <UserOutlined />
          </div>
        </div>
        <div :class="styles.statInfo">
          <div :class="styles.statLabel">用户信息</div>
          <div :class="styles.statValue">{{ userInfo?.username }}</div>
        </div>
        <div :class="styles.statBadge">
          <a-tag color="pink">正常</a-tag>
        </div>
      </div>

      <div :class="styles.statCard">
        <div :class="styles.statIconWrapper">
          <div :class="[styles.statIcon, styles.statIconSuccess]">
            <ThunderboltOutlined />
          </div>
        </div>
        <div :class="styles.statInfo">
          <div :class="styles.statLabel">账户状态</div>
          <div :class="styles.statValue">
            <a-badge status="success" text="活跃" />
          </div>
        </div>
      </div>

      <div :class="styles.statCard">
        <div :class="styles.statIconWrapper">
          <div :class="[styles.statIcon, styles.statIconWarning]">
            <SafetyOutlined />
          </div>
        </div>
        <div :class="styles.statInfo">
          <div :class="styles.statLabel">实名认证 / 可用功能</div>
          <div :class="styles.statValue">{{ userInfo?.isRealName ? '已认证' : '未认证' }}</div>
        </div>
      </div>

      <div :class="styles.statCard">
        <div :class="styles.statIconWrapper">
          <div :class="[styles.statIcon, styles.statIconInfo]">
            <LinkOutlined />
          </div>
        </div>
        <div :class="styles.statInfo">
          <div :class="styles.statLabel">关联账号</div>
          <div :class="styles.statValue">0</div>
        </div>
      </div>
    </div>

    <!-- 快速操作区域 -->
    <div :class="styles.quickActionsSection">
      <h3 :class="styles.sectionTitle">快速操作</h3>
      <div :class="styles.quickActionsGrid">
        <button :class="[styles.quickActionItem, 'quick-action-item']" @click="$router.push('/console/profile')">
          <div :class="styles.actionIconWrapper">
            <UserOutlined />
          </div>
          <span :class="styles.actionLabel">个人资料</span>
        </button>
        <button :class="[styles.quickActionItem, 'quick-action-item']" @click="handleChangePassword">
          <div :class="styles.actionIconWrapper">
            <LockOutlined />
          </div>
          <span :class="styles.actionLabel">修改密码</span>
        </button>
        <button :class="[styles.quickActionItem, 'quick-action-item']" @click="$router.push('/console/realname')">
          <div :class="styles.actionIconWrapper">
            <SafetyOutlined />
          </div>
          <span :class="styles.actionLabel">实名认证</span>
        </button>
        <button :class="[styles.quickActionItem, 'quick-action-item']" @click="handleLogout">
          <div :class="styles.actionIconWrapper">
            <LogoutOutlined />
          </div>
          <span :class="styles.actionLabel">退出登录</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { css } from '@emotion/css'
import { theme, mq, gradients } from '@/styles/theme'
import { 
  UserOutlined, 
  LockOutlined,
  ThunderboltOutlined,
  SafetyOutlined,
  LogoutOutlined,
  ExclamationCircleOutlined,
  LinkOutlined
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()
const userInfo = computed(() => authStore.userInfo)

const handleChangePassword = () => {
  message.info('密码修改功能开发中')
}

const handleLogout = () => {
  authStore.logout()
  router.push('/auth/login')
  message.success('已退出登录')
}

// Emotion CSS 样式
const styles = {
  container: css({
    padding: 0,
    maxWidth: '1200px',
  }),
  
  pageHeader: css({
    marginBottom: theme.spacing['2xl'],
  }),
  
  pageTitle: css({
    fontSize: theme.fontSize['3xl'],
    fontWeight: theme.fontWeight.semibold,
    color: theme.colors.textPrimary,
    margin: 0,
  }),
  
  noticeCards: css({
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing.md,
    marginBottom: theme.spacing['2xl'],
  }),
  
  noticeCard: css({
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing.md,
    padding: theme.spacing.lg,
    borderRadius: theme.radius.md,
    background: theme.colors.white,
    border: `1px solid`,
    transition: `all ${theme.transition.base}`,
    '&:hover': {
      boxShadow: theme.shadows.md,
    },
  }),
  
  noticeInfo: css({
    background: theme.colors.noticeInfoBg,
    borderColor: theme.colors.noticeInfoBorder,
  }),
  
  noticeWarning: css({
    background: theme.colors.noticeWarningBg,
    borderColor: theme.colors.noticeWarningBorder,
  }),
  
  noticeIcon: css({
    width: '40px',
    height: '40px',
    borderRadius: theme.radius.full,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: theme.fontSize['3xl'],
    flexShrink: 0,
  }),
  
  noticeIconInfo: css({
    background: '#e0f2fe',
    color: theme.colors.noticeInfoIcon,
  }),
  
  noticeIconWarning: css({
    background: '#fee2e2',
    color: theme.colors.noticeWarningIcon,
  }),
  
  noticeContent: css({
    flex: 1,
  }),
  
  noticeTitle: css({
    fontSize: theme.fontSize.base,
    fontWeight: theme.fontWeight.semibold,
    color: theme.colors.textPrimary,
    marginBottom: '4px',
  }),
  
  noticeText: css({
    fontSize: theme.fontSize.sm,
    color: theme.colors.textSecondary,
    lineHeight: 1.5,
  }),
  
  noticeAction: css({
    flexShrink: 0,
  }),
  
  statsGrid: css({
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
    gap: theme.spacing.lg,
    marginBottom: theme.spacing['3xl'],
    [mq.md]: {
      gridTemplateColumns: 'repeat(2, 1fr)',
      gap: theme.spacing.md,
    },
    [mq.xs]: {
      gridTemplateColumns: '1fr',
    },
  }),
  
  statCard: css({
    background: theme.colors.white,
    border: `1px solid ${theme.colors.border}`,
    borderRadius: theme.radius.md,
    padding: theme.spacing.xl,
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing.lg,
    transition: `all ${theme.transition.base}`,
    position: 'relative',
    overflow: 'hidden',
    '&:hover': {
      boxShadow: theme.shadows.lg,
      transform: 'translateY(-2px)',
    },
    [mq.md]: {
      padding: theme.spacing.lg,
    },
  }),
  
  statIconWrapper: css({
    flexShrink: 0,
  }),
  
  statIcon: css({
    width: '48px',
    height: '48px',
    borderRadius: theme.radius.full,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: theme.fontSize['3xl'],
    color: theme.colors.white,
    [mq.md]: {
      width: '40px',
      height: '40px',
      fontSize: theme.fontSize['2xl'],
    },
  }),
  
  statIconPrimary: css({
    background: gradients.primary,
  }),
  
  statIconSuccess: css({
    background: gradients.success,
  }),
  
  statIconWarning: css({
    background: gradients.warning,
  }),
  
  statIconInfo: css({
    background: gradients.info,
  }),
  
  statInfo: css({
    flex: 1,
    minWidth: 0,
  }),
  
  statLabel: css({
    fontSize: theme.fontSize.sm,
    color: theme.colors.gray500,
    marginBottom: '6px',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  }),
  
  statValue: css({
    fontSize: theme.fontSize['2xl'],
    fontWeight: theme.fontWeight.semibold,
    color: theme.colors.textPrimary,
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    [mq.md]: {
      fontSize: theme.fontSize.xl,
    },
  }),
  
  statBadge: css({
    flexShrink: 0,
  }),
  
  quickActionsSection: css({
    marginTop: theme.spacing['3xl'],
  }),
  
  sectionTitle: css({
    fontSize: theme.fontSize.xl,
    fontWeight: theme.fontWeight.semibold,
    color: theme.colors.textPrimary,
    margin: `0 0 ${theme.spacing.lg} 0`,
  }),
  
  quickActionsGrid: css({
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))',
    gap: theme.spacing.md,
    [mq.md]: {
      gridTemplateColumns: 'repeat(2, 1fr)',
    },
  }),
  
  quickActionItem: css({
    background: theme.colors.white,
    border: `1px solid ${theme.colors.border}`,
    borderRadius: theme.radius.md,
    padding: `${theme.spacing.xl} ${theme.spacing.lg}`,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: theme.spacing.md,
    cursor: 'pointer',
    transition: `all ${theme.transition.base}`,
    '&:hover': {
      borderColor: theme.colors.primary,
      boxShadow: `0 2px 8px rgba(236, 72, 153, 0.15)`,
      transform: 'translateY(-2px)',
    },
    [mq.md]: {
      padding: `${theme.spacing.lg} ${theme.spacing.md}`,
    },
  }),
  
  actionIconWrapper: css({
    width: '48px',
    height: '48px',
    borderRadius: theme.radius.full,
    background: 'linear-gradient(135deg, #fdf2f8 0%, #fce7f3 100%)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: theme.fontSize['3xl'],
    color: theme.colors.primary,
    transition: `all ${theme.transition.base}`,
    '.quick-action-item:hover &': {
      background: gradients.primary,
      color: theme.colors.white,
    },
    [mq.md]: {
      width: '40px',
      height: '40px',
      fontSize: theme.fontSize['2xl'],
    },
  }),
  
  actionLabel: css({
    fontSize: theme.fontSize.sm,
    fontWeight: theme.fontWeight.medium,
    color: theme.colors.textPrimary,
    textAlign: 'center',
    [mq.md]: {
      fontSize: theme.fontSize.xs,
    },
  }),
}
</script>


