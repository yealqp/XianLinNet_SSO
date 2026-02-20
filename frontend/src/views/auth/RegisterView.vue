<template>
  <AuthLayout>
    <div class="register-view">
      <h2 class="title">注册 XianlinNet ID</h2>
      <a-form
        :model="formState"
        name="register"
        @finish="handleRegister"
        layout="vertical"
      >
        <a-form-item
          name="username"
          :rules="[{ required: true, message: '请输入用户名' }]"
        >
          <a-input
            v-model:value="formState.username"
            placeholder="用户名"
            size="large"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item
          name="email"
          :rules="[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '请输入有效的邮箱地址' },
          ]"
        >
          <a-input
            v-model:value="formState.email"
            placeholder="邮箱"
            size="large"
          >
            <template #prefix>
              <MailOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="verificationCode"
          :rules="[{ required: true, message: '请输入验证码' }]"
        >
          <a-input
            v-model:value="formState.verificationCode"
            placeholder="邮箱验证码"
            size="large"
          >
            <template #prefix>
              <SafetyOutlined />
            </template>
            <template #suffix>
              <a-button
                type="link"
                size="small"
                :disabled="countdown > 0 || !captchaToken"
                :loading="sendingCode"
                @click="handleSendCode"
              >
                {{ countdown > 0 ? `${countdown}秒后重试` : "发送验证码" }}
              </a-button>
            </template>
          </a-input>
        </a-form-item>

        <!-- Captcha Widget -->
        <a-form-item>
          <Captcha
            :site-key="captchaSiteKey"
            :api-endpoint="captchaApiEndpoint"
            @success="handleCaptchaSuccess"
            @error="handleCaptchaError"
            ref="captchaRef"
          />
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[
            { required: true, message: '请输入密码' },
            { min: 6, message: '密码至少 6 个字符' },
          ]"
        >
          <a-input-password
            v-model:value="formState.password"
            placeholder="密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item
          name="confirmPassword"
          :rules="[
            { required: true, message: '请确认密码' },
            { validator: validateConfirmPassword },
          ]"
        >
          <a-input-password
            v-model:value="formState.confirmPassword"
            placeholder="确认密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            block
          >
            注册
          </a-button>
        </a-form-item>

        <div class="extra-links">
          已有账号？<router-link to="/login">立即登录</router-link>
        </div>
      </a-form>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import {
  UserOutlined,
  LockOutlined,
  MailOutlined,
  SafetyOutlined,
} from "@ant-design/icons-vue";
import AuthLayout from "@/components/layout/AuthLayout.vue";
import Captcha from "@/components/Captcha.vue";
import { message } from "ant-design-vue";
import { authApi } from "@/api/auth";

const router = useRouter();
const captchaSiteKey = import.meta.env.VITE_CAPTCHA_SITE_KEY || '1cbf106b94';
const captchaApiEndpoint = import.meta.env.VITE_CAPTCHA_API_ENDPOINT || 'https://captcha.yealqp.cn/1cbf106b94';

const formState = reactive({
  email: "",
  verificationCode: "",
  username: "",
  password: "",
  confirmPassword: "",
});

const loading = ref(false);
const sendingCode = ref(false);
const countdown = ref(0);
const captchaToken = ref("");
const captchaRef = ref<InstanceType<typeof Captcha> | null>(null);

const validateConfirmPassword = () => {
  if (
    formState.password &&
    formState.confirmPassword &&
    formState.password !== formState.confirmPassword
  ) {
    return Promise.reject("两次输入的密码不一致");
  }
  return Promise.resolve();
};

const handleCaptchaSuccess = (token: string) => {
  captchaToken.value = token;
};

const handleCaptchaError = (error: any) => {
  console.error('Captcha error:', error);
  message.error('人机验证失败，请刷新页面重试');
};

const handleSendCode = async () => {
  if (!formState.email) {
    message.error("请先输入邮箱");
    return;
  }

  if (!captchaToken.value) {
    message.error("请先完成人机验证");
    return;
  }

  sendingCode.value = true;
  try {
    await authApi.sendVerificationCode(
      formState.email,
      "register",
      captchaToken.value
    );
    message.success("验证码已发送到您的邮箱");

    // Reset captcha after successful send
    captchaToken.value = "";
    if (captchaRef.value) {
      captchaRef.value.reset();
    }

    // Start countdown
    countdown.value = 60;
    const timer = setInterval(() => {
      countdown.value--;
      if (countdown.value <= 0) {
        clearInterval(timer);
      }
    }, 1000);
  } catch (error: any) {
    console.error("Send code failed:", error);
    message.error(error.response?.data?.msg || "发送验证码失败");
    // Reset captcha on error
    captchaToken.value = "";
    if (captchaRef.value) {
      captchaRef.value.reset();
    }
  } finally {
    sendingCode.value = false;
  }
};

const handleRegister = async () => {
  loading.value = true;
  try {
    await authApi.register({
      email: formState.email,
      password: formState.password,
      username: formState.username,
      verificationCode: formState.verificationCode,
    });
    message.success("注册成功，请登录");
    router.push("/login");
  } catch (error: any) {
    console.error("Register failed:", error);
    message.error(error.response?.data?.msg || "注册失败，请稍后重试");
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.register-view {
  width: 100%;
}

.title {
  font-size: 26px;
  font-weight: 600;
  text-align: left;
  margin-bottom: 20px;
  color: #ffffff;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.extra-links {
  text-align: center;
  margin-top: 12px;
  color: rgba(255, 255, 255, 0.8);
  font-size: 13px;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.extra-links a {
  color: #ffd6ed;
  text-decoration: none;
  transition: all 0.3s;
  margin-left: 4px;
}

.extra-links a:hover {
  color: #ffffff;
  text-shadow: 0 2px 6px rgba(246, 51, 154, 0.6);
}

:deep(.ant-form-item-label > label) {
  color: rgba(255, 255, 255, 0.9);
  font-size: 13px;
  font-weight: 500;
}

:deep(.ant-input-affix-wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  padding: 10px 14px;
  transition: all 0.3s;
}

:deep(.ant-input-affix-wrapper:hover) {
  background: rgba(255, 255, 255, 0.98);
  border-color: #ff4db3;
  box-shadow: 0 2px 8px rgba(246, 51, 154, 0.15);
}

:deep(.ant-input-affix-wrapper:focus),
:deep(.ant-input-affix-wrapper-focused) {
  background: #ffffff;
  border-color: #f6339a;
  box-shadow: 0 0 0 3px rgba(246, 51, 154, 0.15);
}

:deep(.ant-input) {
  background: transparent;
  color: #1f1f1f;
  font-size: 14px;
}

:deep(.ant-input::placeholder) {
  color: #999999;
}

:deep(.ant-input-prefix) {
  color: #666666;
}

:deep(.ant-btn-primary) {
  height: 44px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  background: linear-gradient(135deg, #f6339a 0%, #ff4db3 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(246, 51, 154, 0.4);
  transition: all 0.3s;
  margin-top: 4px;
}

:deep(.ant-btn-primary:hover:not(:disabled)) {
  background: linear-gradient(135deg, #ff4db3 0%, #ff66c4 100%);
  box-shadow: 0 6px 16px rgba(246, 51, 154, 0.5);
  transform: translateY(-2px);
}

:deep(.ant-btn-primary:active) {
  background: linear-gradient(135deg, #e02987 0%, #f6339a 100%);
  transform: translateY(0);
}

:deep(.ant-btn-primary:disabled) {
  background: rgba(255, 255, 255, 0.3);
  box-shadow: none;
}

:deep(.ant-btn-link) {
  color: #f6339a;
  padding: 0;
  height: auto;
  font-size: 12px;
  font-weight: 500;
}

:deep(.ant-btn-link:hover:not(:disabled)) {
  color: #ff4db3;
}

:deep(.ant-btn-link:disabled) {
  color: #999999;
}

:deep(.ant-form-item) {
  margin-bottom: 14px;
}

:deep(.ant-form-item-explain-error) {
  color: #ffccc7;
  background: rgba(255, 77, 79, 0.15);
  padding: 4px 8px;
  border-radius: 4px;
  margin-top: 4px;
  font-size: 12px;
}

@media (max-width: 768px) {
  .title {
    font-size: 22px;
    margin-bottom: 16px;
  }

  :deep(.ant-btn-primary) {
    height: 42px;
    font-size: 14px;
  }

  .extra-links {
    font-size: 12px;
  }
}
</style>
