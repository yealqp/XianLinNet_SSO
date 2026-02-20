<template>
  <div class="captcha-wrapper">
    <!-- SVG 渐变定义 -->
    <svg width="0" height="0" style="position: absolute;">
      <defs>
        <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" style="stop-color:#f6339a;stop-opacity:1" />
          <stop offset="100%" style="stop-color:#ff4db3;stop-opacity:1" />
        </linearGradient>
      </defs>
    </svg>
    
    <div 
      class="captcha-box"
      :class="{ 'is-verifying': isVerifying, 'is-verified': isVerified, 'is-error': hasError }"
      @click="handleClick"
      role="button"
      tabindex="0"
      :aria-label="ariaLabel"
      @keydown.enter.prevent="handleClick"
      @keydown.space.prevent="handleClick"
    >
      <div class="checkbox">
        <svg v-if="isVerifying" class="progress-ring" viewBox="0 0 32 32">
          <circle class="progress-ring-bg" cx="16" cy="16" r="14"></circle>
          <circle 
            class="progress-ring-circle" 
            cx="16" cy="16" 
            r="14"
            :style="{ strokeDashoffset: progressOffset }"
          ></circle>
        </svg>
        <svg v-else-if="isVerified" class="checkmark" viewBox="0 0 24 24">
          <path 
            fill="none" 
            stroke="#00a67d" 
            stroke-linecap="round" 
            stroke-linejoin="round" 
            stroke-width="2" 
            d="m5 12 5 5L20 7"
          />
        </svg>
        <svg v-else-if="hasError" class="error-icon" viewBox="0 0 24 24">
          <path 
            fill="#f55b50" 
            d="M11 15h2v2h-2zm0-8h2v6h-2zm1-5C6.47 2 2 6.5 2 12a10 10 0 0 0 10 10a10 10 0 0 0 10-10A10 10 0 0 0 12 2m0 18a8 8 0 0 1-8-8a8 8 0 0 1 8-8a8 8 0 0 1 8 8a8 8 0 0 1-8 8"
          />
        </svg>
      </div>
      <p class="label">{{ statusText }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps<{
  siteKey: string
  apiEndpoint?: string
}>()

const emit = defineEmits<{
  (e: 'success', token: string): void
  (e: 'error', error: any): void
}>()

// 状态
const isVerifying = ref(false)
const isVerified = ref(false)
const hasError = ref(false)
const progress = ref(0)
const workerCount = ref(navigator.hardwareConcurrency || 8)

// 计算属性
const statusText = computed(() => {
  if (isVerifying.value) return `验证中... ${progress.value}%`
  if (isVerified.value) return '验证成功'
  if (hasError.value) return '验证失败，请重试'
  return '我是人类'
})

const ariaLabel = computed(() => {
  if (isVerifying.value) return '正在验证您是人类，请稍候'
  if (isVerified.value) return '我们已验证您是人类，您现在可以继续'
  if (hasError.value) return '发生错误，请重试'
  return '点击验证您是人类'
})

const progressOffset = computed(() => {
  const circumference = 2 * Math.PI * 14
  return circumference - (progress.value / 100) * circumference
})

// 生成伪随机数（用于生成 salt 和 target）
function seededRandom(seed: string, length: number): string {
  let hash = 2166136261
  for (let i = 0; i < seed.length; i++) {
    hash ^= seed.charCodeAt(i)
    hash += (hash << 1) + (hash << 4) + (hash << 7) + (hash << 8) + (hash << 24)
  }
  hash = hash >>> 0

  let result = ''
  while (result.length < length) {
    hash ^= hash << 13
    hash ^= hash >>> 17
    hash ^= hash << 5
    hash = hash >>> 0
    result += hash.toString(16).padStart(8, '0')
  }
  return result.substring(0, length)
}

// 全局 WASM 模块缓存
let globalWasmModule: WebAssembly.Module | null = null
let globalWasmLoading: Promise<WebAssembly.Module> | null = null

// 全局 Worker Blob URL 缓存
let globalWorkerBlobUrl: string | null = null

// 创建 Worker 代码的 Blob URL（只创建一次）
const getWorkerBlobUrl = (): string => {
  if (globalWorkerBlobUrl) {
    return globalWorkerBlobUrl
  }
  
  const workerCode = `
    let wasmInstance = null;
    
    // 初始化 WASM 实例
    async function initWasm(wasmModule) {
      if (wasmInstance) return wasmInstance;
      
      if (!wasmModule) {
        console.warn('[Cap Worker] 没有 WASM 模块，使用 JS 降级');
        return null;
      }
      
      try {
        // WASM 导入对象
        const imports = {
          wbg: {
            __wbindgen_init_externref_table: function() {}
          }
        };
        
        // 实例化预编译的 WASM 模块
        const instance = await WebAssembly.instantiate(wasmModule, imports);
        
        // 创建 solve_pow 函数包装器
        const memory = instance.exports.memory;
        const solve_pow_raw = instance.exports.solve_pow;
        const malloc = instance.exports.__wbindgen_malloc;
        const start = instance.exports.__wbindgen_start;
        
        // 初始化
        if (start) start();
        
        // 文本编码器
        const encoder = new TextEncoder();
        
        // 字符串传递到 WASM
        function passStringToWasm(str) {
          const encoded = encoder.encode(str);
          const ptr = malloc(encoded.length, 1);
          const mem = new Uint8Array(memory.buffer);
          mem.set(encoded, ptr);
          return [ptr, encoded.length];
        }
        
        // solve_pow 包装函数
        function solve_pow(salt, target) {
          const [saltPtr, saltLen] = passStringToWasm(salt);
          const [targetPtr, targetLen] = passStringToWasm(target);
          const result = solve_pow_raw(saltPtr, saltLen, targetPtr, targetLen);
          return result;
        }
        
        wasmInstance = { solve_pow };
        return wasmInstance;
      } catch (e) {
        console.warn('[Cap Worker] WASM 实例化失败:', e);
        return null;
      }
    }
    
    self.onmessage = async ({ data: { salt, target, wasmModule } }) => {
      try {
        // 初始化 WASM 实例
        const wasm = await initWasm(wasmModule);
        
        if (wasm && wasm.solve_pow) {
          // 使用 WASM 求解
          const startTime = performance.now();
          const nonce = wasm.solve_pow(salt, target);
          const duration = performance.now() - startTime;
          self.postMessage({ 
            nonce: Number(nonce), 
            found: true,
            durationMs: duration.toFixed(2)
          });
        } else {
          // JS 降级实现
          const encoder = new TextEncoder();
          const targetBytes = new Uint8Array(target.length / 2);
          for (let i = 0; i < targetBytes.length; i++) {
            targetBytes[i] = parseInt(target.substring(i * 2, i * 2 + 2), 16);
          }
          
          let nonce = 0;
          const targetLen = targetBytes.length;
          
          while (true) {
            const input = salt + nonce;
            const inputBytes = encoder.encode(input);
            const hashBuffer = await crypto.subtle.digest('SHA-256', inputBytes);
            const hashBytes = new Uint8Array(hashBuffer, 0, targetLen);
            
            let match = true;
            for (let i = 0; i < targetLen; i++) {
              if (hashBytes[i] !== targetBytes[i]) {
                match = false;
                break;
              }
            }
            
            if (match) {
              self.postMessage({ nonce, found: true });
              return;
            }
            
            nonce++;
            
            // 每 50000 次检查一次，避免阻塞
            if (nonce % 50000 === 0) {
              await new Promise(resolve => setTimeout(resolve, 0));
            }
          }
        }
      } catch (error) {
        self.postMessage({ 
          found: false, 
          error: error.message || String(error) 
        });
      }
    };
  `
  
  const blob = new Blob([workerCode], { type: 'application/javascript' })
  globalWorkerBlobUrl = URL.createObjectURL(blob)
  console.log('[Cap] Worker Blob URL 已创建（将被所有 Worker 复用）')
  
  return globalWorkerBlobUrl
}

// 加载并编译 WASM 模块（只加载一次）
const loadWasmModule = async (): Promise<WebAssembly.Module> => {
  // 如果已经加载，直接返回
  if (globalWasmModule) {
    console.log('[Cap] 使用缓存的 WASM 模块')
    return globalWasmModule
  }
  
  // 如果正在加载，等待加载完成
  if (globalWasmLoading) {
    console.log('[Cap] 等待 WASM 模块加载完成')
    return globalWasmLoading
  }

  // 开始加载
  console.log('[Cap] 开始加载 WASM 模块')
  globalWasmLoading = (async () => {
    try {
      const wasmUrl = new URL('/cap_wasm_bg.wasm', window.location.origin).href
      const wasmResponse = await fetch(wasmUrl)
      const wasmBuffer = await wasmResponse.arrayBuffer()
      
      // 编译 WASM 模块（只编译一次）
      const module = await WebAssembly.compile(wasmBuffer)
      globalWasmModule = module
      console.log('[Cap] WASM 模块编译成功')
      return module
    } catch (error) {
      console.error('[Cap] WASM 模块加载失败:', error)
      globalWasmLoading = null
      throw error
    }
  })()
  
  return globalWasmLoading
}

// 解决多个挑战
const solveChallenges = async (challenges: Array<[string, string]>) => {
  const totalChallenges = challenges.length
  let completedChallenges = 0
  const solutions: number[] = []
  
  // 预加载并编译 WASM 模块（只加载一次）
  let wasm: WebAssembly.Module | null = null
  try {
    wasm = await loadWasmModule()
    console.log('[Cap] WASM 模块已准备好，将被所有 Worker 共享')
  } catch (error) {
    console.warn('[Cap] WASM 加载失败，将使用 JS 降级:', error)
  }
  
  // 一次性创建所有 Worker（复用 Blob URL）
  const workerUrl = getWorkerBlobUrl()
  const workers = Array(workerCount.value)
    .fill(null)
    .map(() => {
      try {
        return new Worker(workerUrl)
      } catch (e) {
        console.error('[Cap] Worker 创建失败:', e)
        throw new Error('Worker 创建失败')
      }
    })
  
  console.log(`[Cap] 创建了 ${workers.length} 个 Worker`)
  
  try {
    // 处理每个挑战
    const solveChallenge = ([salt, target]: [string, string], workerIndex: number): Promise<number> => {
      return new Promise((resolve, reject) => {
        const worker = workers[workerIndex]
        if (!worker) {
          reject(new Error('Worker 不可用'))
          return
        }
        
        worker.onmessage = ({ data }) => {
          if (data.found) {
            completedChallenges++
            progress.value = Math.round((completedChallenges / totalChallenges) * 100)
            if (data.durationMs) {
              console.log(`[Cap Worker ${workerIndex}] 求解完成，耗时: ${data.durationMs}ms`)
            }
            resolve(data.nonce)
          } else {
            reject(new Error(data.error || 'Worker 求解失败'))
          }
        }
        
        worker.onerror = (error) => {
          reject(error)
        }
        
        // 发送任务到 Worker
        worker.postMessage({ salt, target, wasmModule: wasm })
      })
    }
    
    // 分批处理挑战
    for (let i = 0; i < challenges.length; i += workerCount.value) {
      const batch = challenges.slice(i, Math.min(i + workerCount.value, challenges.length))
      console.log(`[Cap] 处理批次 ${Math.floor(i / workerCount.value) + 1}，包含 ${batch.length} 个挑战`)
      
      const batchSolutions = await Promise.all(
        batch.map((challenge, index) => solveChallenge(challenge, index))
      )
      solutions.push(...batchSolutions)
    }
  } finally {
    // 清理所有 Worker
    workers.forEach((worker, index) => {
      if (worker) {
        try {
          worker.terminate()
          console.log(`[Cap] Worker ${index} 已终止`)
        } catch (e) {
          console.error(`[Cap] Worker ${index} 终止失败:`, e)
        }
      }
    })
  }
  
  return solutions
}

// 获取挑战
const fetchChallenge = async () => {
  let endpoint = props.apiEndpoint
  if (!endpoint) {
    throw new Error('缺少 API 端点配置')
  }
  
  if (!endpoint.endsWith('/')) {
    endpoint += '/'
  }
  
  const response = await fetch(`${endpoint}challenge`, {
    method: 'POST'
  })
  
  if (!response.ok) {
    throw new Error(`获取挑战失败: ${response.status}`)
  }
  
  return await response.json()
}

// 兑换 token
const redeemToken = async (token: string, solutions: number[]) => {
  let endpoint = props.apiEndpoint
  if (!endpoint) {
    throw new Error('缺少 API 端点配置')
  }
  
  if (!endpoint.endsWith('/')) {
    endpoint += '/'
  }
  
  const response = await fetch(`${endpoint}redeem`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ token, solutions })
  })
  
  if (!response.ok) {
    throw new Error(`兑换 token 失败: ${response.status}`)
  }
  
  return await response.json()
}

// 主验证流程
const solve = async () => {
  if (isVerifying.value || isVerified.value) return
  
  try {
    isVerifying.value = true
    hasError.value = false
    progress.value = 0
    
    // 获取挑战
    const challengeData = await fetchChallenge()
    const { challenge, token } = challengeData
    
    // 生成挑战数组
    let challenges: Array<[string, string]>
    if (Array.isArray(challenge)) {
      challenges = challenge
    } else {
      // 根据挑战配置生成
      const { c: count, s: saltLen, d: targetLen } = challenge
      challenges = Array.from({ length: count }, (_, i) => {
        const index = i + 1
        return [
          seededRandom(`${token}${index}`, saltLen),
          seededRandom(`${token}${index}d`, targetLen)
        ]
      })
    }
    
    console.log(`[Cap] 开始求解 ${challenges.length} 个挑战`)
    
    // 求解所有挑战（WASM 模块会在这里加载一次）
    const solutions = await solveChallenges(challenges)
    
    console.log('[Cap] 所有挑战已求解，正在兑换 token')
    
    // 兑换 token
    const result = await redeemToken(token, solutions)
    
    if (!result.success) {
      throw new Error('Token 兑换失败')
    }
    
    // 验证成功
    isVerified.value = true
    isVerifying.value = false
    progress.value = 100
    
    console.log('[Cap] 验证成功，token:', result.token)
    emit('success', result.token)
    
  } catch (error: any) {
    console.error('[Cap] 验证失败:', error)
    hasError.value = true
    isVerifying.value = false
    emit('error', error)
  }
}

// 点击处理
const handleClick = () => {
  if (isVerifying.value || isVerified.value) return
  solve()
}

// 重置
const reset = () => {
  isVerifying.value = false
  isVerified.value = false
  hasError.value = false
  progress.value = 0
}

// 暴露方法
defineExpose({
  reset,
  solve
})

onMounted(() => {
  console.log('[Cap] Captcha 组件已挂载')
})

onBeforeUnmount(() => {
  reset()
  
  // 清理全局 Blob URL（如果这是最后一个使用它的组件）
  // 注意：在实际应用中，如果有多个 Captcha 组件，可能需要引用计数
  if (globalWorkerBlobUrl) {
    console.log('[Cap] 清理 Worker Blob URL')
    URL.revokeObjectURL(globalWorkerBlobUrl)
    globalWorkerBlobUrl = null
  }
})
</script>

<style scoped>
.captcha-wrapper {
  margin: 16px 0;
  display: flex;
  justify-content: center;
  min-height: 80px;
}

.captcha-box {
  background: linear-gradient(135deg, #ffffff 0%, #fef5f9 100%);
  border: 2px solid #ffe0ed;
  border-radius: 12px;
  user-select: none;
  height: 64px;
  width: 100%;
  max-width: 320px;
  display: flex;
  align-items: center;
  padding: 16px 20px;
  gap: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  color: #262626;
  box-shadow: 0 2px 8px rgba(246, 51, 154, 0.08);
}

.captcha-box::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(246, 51, 154, 0.05) 0%, transparent 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.captcha-box:hover {
  border-color: #ffb3d9;
  box-shadow: 0 4px 16px rgba(246, 51, 154, 0.15);
  transform: translateY(-1px);
}

.captcha-box:hover::before {
  opacity: 1;
}

.captcha-box:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(246, 51, 154, 0.12);
}

.captcha-box.is-verifying {
  cursor: progress;
  border-color: #f6339a;
  box-shadow: 0 4px 16px rgba(246, 51, 154, 0.2);
}

.captcha-box.is-verifying::before {
  opacity: 1;
}

.captcha-box.is-verified {
  cursor: default;
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  border-color: #86efac;
  box-shadow: 0 4px 16px rgba(34, 197, 94, 0.15);
}

.captcha-box.is-error {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border-color: #fca5a5;
  box-shadow: 0 4px 16px rgba(239, 68, 68, 0.15);
}

.checkbox {
  width: 32px;
  height: 32px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  background: #ffffff;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.captcha-box:hover .checkbox {
  border-color: #f6339a;
  box-shadow: 0 2px 6px rgba(246, 51, 154, 0.15);
}

.is-verifying .checkbox {
  background: linear-gradient(135deg, #fef5f9 0%, #ffffff 100%);
  transform: scale(1.05);
  border-color: #f6339a;
  border-radius: 50%;
  box-shadow: 0 0 0 4px rgba(246, 51, 154, 0.1);
}

.is-verified .checkbox {
  border-color: #10b981;
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.1);
}

.is-error .checkbox {
  border-color: #ef4444;
  background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
  box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.1);
}

.progress-ring {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.progress-ring-bg {
  fill: none;
  stroke: #f3f4f6;
  stroke-width: 3;
}

.progress-ring-circle {
  fill: none;
  stroke: url(#gradient);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-dasharray: 87.96;
  transition: stroke-dashoffset 0.3s ease;
  filter: drop-shadow(0 0 2px rgba(246, 51, 154, 0.3));
}

.checkmark {
  width: 100%;
  height: 100%;
  animation: checkmark-appear 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes checkmark-appear {
  0% {
    opacity: 0;
    transform: scale(0.5) rotate(-45deg);
  }
  50% {
    transform: scale(1.1) rotate(5deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}

.checkmark path {
  stroke-dasharray: 23.21320343017578px;
  stroke-dashoffset: 23.21320343017578px;
  animation: checkmark-draw 0.6s cubic-bezier(0.4, 0, 0.2, 1) forwards;
  filter: drop-shadow(0 1px 2px rgba(16, 185, 129, 0.3));
}

@keyframes checkmark-draw {
  to {
    stroke-dashoffset: 0;
  }
}

.error-icon {
  width: 100%;
  height: 100%;
  animation: error-appear 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes error-appear {
  0% {
    opacity: 0;
    transform: scale(0.3);
  }
  50% {
    transform: scale(1.15);
  }
  70% {
    transform: scale(0.95);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.error-icon path {
  filter: drop-shadow(0 1px 2px rgba(239, 68, 68, 0.3));
}

.label {
  margin: 0;
  font-weight: 600;
  font-size: 15px;
  user-select: none;
  transition: all 0.3s ease;
  flex: 1;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  letter-spacing: 0.01em;
  color: #1f2937;
}

.is-verifying .label {
  color: #f6339a;
  font-weight: 700;
}

.is-verified .label {
  color: #059669;
  font-weight: 700;
}

.is-error .label {
  color: #dc2626;
  font-weight: 700;
}

/* 添加渐变定义 */
.captcha-box::after {
  content: '';
  position: absolute;
  width: 0;
  height: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .captcha-box {
    max-width: 100%;
    height: 60px;
    padding: 14px 18px;
  }
  
  .checkbox {
    width: 28px;
    height: 28px;
  }
  
  .label {
    font-size: 14px;
  }
}

/* 深色模式支持 */
@media (prefers-color-scheme: dark) {
  .captcha-box {
    background: linear-gradient(135deg, #1f2937 0%, #111827 100%);
    border-color: #374151;
    color: #f3f4f6;
  }
  
  .checkbox {
    background: #374151;
    border-color: #4b5563;
  }
  
  .label {
    color: #f3f4f6;
  }
  
  .is-verifying .checkbox {
    background: linear-gradient(135deg, #1f2937 0%, #111827 100%);
  }
}
</style>
