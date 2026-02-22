import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/console/dashboard'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('@/views/auth/ForgotPasswordView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/authorize',
    name: 'Authorize',
    component: () => import('@/views/auth/AuthorizeView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/console',
    name: 'Console',
    component: () => import('@/components/layout/ConsoleLayout.vue'),
    meta: { requiresAuth: true },
    redirect: '/console/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/console/UserDashboardView.vue')
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/console/ProfileView.vue')
      },
      {
        path: 'realname',
        name: 'RealName',
        component: () => import('@/views/console/RealNameView.vue')
      },
      {
        path: 'authorizations',
        name: 'Authorizations',
        component: () => import('@/views/console/AuthorizationsView.vue')
      }
    ]
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/components/layout/ConsoleLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: 'dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/console/DashboardView.vue')
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: () => import('@/views/admin/UsersView.vue')
      },
      {
        path: 'applications',
        name: 'ApplicationManagement',
        component: () => import('@/views/admin/ApplicationsView.vue')
      },
      {
        path: 'tokens',
        name: 'TokenManagement',
        component: () => import('@/views/admin/TokensView.vue')
      }
    ]
  },
  {
    path: '/unauthorized',
    name: 'Unauthorized',
    component: () => import('@/views/UnauthorizedView.vue')
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFoundView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.meta.requiresAuth !== false
  const requiresAdmin = to.meta.requiresAdmin === true

  // 需要认证的路由
  if (requiresAuth) {
    // 检查是否有 token
    if (!authStore.isAuthenticated) {
      console.log('No authentication token, redirecting to login')
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }

    // 检查是否有用户信息
    if (!authStore.userInfo) {
      console.log('No user info, redirecting to login')
      authStore.logout()
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }

    // 需要管理员权限的路由
    if (requiresAdmin && !authStore.isAdmin) {
      console.log('User is not admin, redirecting to unauthorized')
      next({ name: 'Unauthorized' })
      return
    }
  }

  // 已登录用户访问登录/注册页面，重定向到控制台
  if ((to.name === 'Login' || to.name === 'Register' || to.name === 'ForgotPassword') && authStore.isAuthenticated) {
    next({ name: 'Dashboard' })
    return
  }

  next()
})

export default router
