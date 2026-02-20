<template>
  <div class="auth-layout">
    <div class="auth-form-card">
      <div class="auth-form-content">
        <slot></slot>
      </div>
      <div class="auth-footer">
        <p class="footer-text">XianLinNet 沪ICP备2025144886号</p>
        <p class="footer-text">© 2026 XianLinNet. All right reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.auth-layout {
  height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 40px 80px;
  background-image: url('https://t.alcy.cc/moe');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
}

.auth-layout::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(2px);
}

.auth-form-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  max-height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
  background: rgba(30, 30, 30, 0.25);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
}

.auth-form-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 36px 36px 20px;
  /* 自定义滚动条样式 */
  scrollbar-width: thin;
  scrollbar-color: rgba(246, 51, 154, 0.5) rgba(255, 255, 255, 0.1);
}

/* Webkit 浏览器滚动条样式 */
.auth-form-content::-webkit-scrollbar {
  width: 6px;
}

.auth-form-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.auth-form-content::-webkit-scrollbar-thumb {
  background: rgba(246, 51, 154, 0.5);
  border-radius: 3px;
  transition: background 0.3s ease;
}

.auth-form-content::-webkit-scrollbar-thumb:hover {
  background: rgba(246, 51, 154, 0.7);
}

.auth-footer {
  flex-shrink: 0;
  padding: 16px 36px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  text-align: center;
  background: rgba(0, 0, 0, 0.15);
}

.footer-text {
  margin: 0;
  padding: 3px 0;
  color: rgba(255, 255, 255, 0.65);
  font-size: 11px;
  line-height: 1.4;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

@media (max-width: 1200px) {
  .auth-layout {
    padding: 40px 60px;
  }
  
  .auth-form-card {
    max-height: calc(100vh - 80px);
  }
}

@media (max-width: 992px) {
  .auth-layout {
    padding: 40px;
    justify-content: center;
  }

  .auth-form-card {
    max-width: 480px;
    max-height: calc(100vh - 80px);
  }
}

@media (max-width: 768px) {
  .auth-layout {
    padding: 24px;
  }

  .auth-form-card {
    max-width: 100%;
    max-height: calc(100vh - 48px);
  }

  .auth-form-content {
    padding: 32px 28px 16px;
  }

  .auth-footer {
    padding: 14px 28px 16px;
  }

  .footer-text {
    font-size: 10px;
  }
}

@media (max-width: 480px) {
  .auth-layout {
    padding: 16px;
  }

  .auth-form-card {
    max-height: calc(100vh - 32px);
    border-radius: 8px;
  }

  .auth-form-content {
    padding: 28px 24px 14px;
  }

  .auth-footer {
    padding: 12px 24px 14px;
  }

  .footer-text {
    font-size: 9px;
  }
}

/* 平滑滚动 */
.auth-form-content {
  scroll-behavior: smooth;
}

/* 滚动时添加渐变遮罩效果 */
.auth-form-card::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 100px;
  background: linear-gradient(to top, rgba(30, 30, 30, 0.25), transparent);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 2;
}

.auth-form-card:has(.auth-form-content:not(:hover)):has(.auth-form-content[scrollTop]) .auth-form-card::after {
  opacity: 1;
}
</style>
