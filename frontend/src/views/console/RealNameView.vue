<template>
  <div class="realname-view">
    <h2 class="page-title">实名认证</h2>

    <a-card v-if="!userInfo?.isRealName">
      <a-alert
        message="实名认证说明"
        description="完成实名认证后，您将获得更多功能权限。请确保填写的信息真实有效。"
        type="info"
        show-icon
        style="margin-bottom: 24px;"
      />

      <a-form
        :model="formState"
        name="realname"
        @finish="handleSubmit"
        layout="vertical"
      >
        <a-form-item
          label="真实姓名"
          name="name"
          :rules="[{ required: true, message: '请输入真实姓名' }]"
        >
          <a-input
            v-model:value="formState.name"
            placeholder="请输入真实姓名"
            size="large"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          label="身份证号"
          name="idcard"
          :rules="[
            { required: true, message: '请输入身份证号' },
            { pattern: /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/, message: '请输入有效的身份证号' }
          ]"
        >
          <a-input
            v-model:value="formState.idcard"
            placeholder="请输入18位身份证号"
            size="large"
            maxlength="18"
          >
            <template #prefix>
              <IdcardOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item>
          <a-space>
            <a-button
              type="primary"
              html-type="submit"
              size="large"
              :loading="loading"
            >
              提交认证
            </a-button>
            <a-button size="large" @click="resetForm">
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>

      <a-divider />

      <div class="tips">
        <h4>温馨提示：</h4>
        <ul>
          <li>请确保填写的姓名和身份证号真实有效</li>
          <li>身份证信息将用于实名认证，不会泄露给第三方</li>
          <li>实名认证成功后，您的账号将升级为实名用户</li>
          <li>如有疑问，请联系客服</li>
        </ul>
      </div>
    </a-card>

    <a-card v-else>
      <a-result
        status="success"
        title="您已完成实名认证"
        sub-title="您的账号已通过实名认证，可以使用所有功能。"
      >
        <template #extra>
          <a-button type="primary" @click="$router.push('/console/dashboard')">
            返回首页
          </a-button>
        </template>
      </a-result>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { storage } from '@/utils/storage'
import { UserOutlined, IdcardOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const authStore = useAuthStore()
const userInfo = computed(() => authStore.userInfo)

const formState = reactive({
  name: '',
  idcard: ''
})

const loading = ref(false)

const handleSubmit = async () => {
  if (!userInfo.value?.id) {
    message.error('用户信息获取失败')
    return
  }

  loading.value = true
  try {
    const response = await authApi.submitRealName({
      userId: userInfo.value.id,
      name: formState.name,
      idcard: formState.idcard
    })

    if (response.status === 'ok') {
      message.success('实名认证成功！')
      
      // 更新本地用户信息
      if (authStore.userInfo) {
        authStore.userInfo.isRealName = true
        // 保存到 localStorage
        storage.setUserInfo(authStore.userInfo)
      }

      // 重置表单
      resetForm()
    } else {
      message.error(response.msg || '实名认证失败')
    }
  } catch (error: any) {
    console.error('Real name verification failed:', error)
    message.error(error.response?.data?.msg || '实名认证失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  formState.name = ''
  formState.idcard = ''
}
</script>

<style scoped>
.realname-view {
  padding: 0;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 24px;
  color: #1f1f1f;
}

.tips {
  background: #f6f8fa;
  padding: 16px;
  border-radius: 6px;
}

.tips h4 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 15px;
}

.tips ul {
  margin: 0;
  padding-left: 20px;
  color: #666;
  font-size: 14px;
  line-height: 1.8;
}

.tips li {
  margin-bottom: 4px;
}

:deep(.ant-form-item-label > label) {
  color: #595959;
  font-size: 14px;
}

:deep(.ant-input-affix-wrapper) {
  background: #ffffff;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
}

:deep(.ant-input-affix-wrapper:hover) {
  border-color: #ff4db3;
}

:deep(.ant-input-affix-wrapper:focus),
:deep(.ant-input-affix-wrapper-focused) {
  border-color: #f6339a;
  box-shadow: 0 0 0 2px rgba(246, 51, 154, 0.1);
}

:deep(.ant-btn-primary) {
  background: #f6339a;
  border: none;
}

:deep(.ant-btn-primary:hover) {
  background: #ff4db3;
}

:deep(.ant-btn-primary:active) {
  background: #e02987;
}
</style>
