import { css } from '@emotion/css'
import { theme } from '@/styles/theme'

class TopLoader {
  private element: HTMLDivElement | null = null
  private timer: number | null = null
  private progress: number = 0
  private isLoading: boolean = false

  private readonly loaderClass = css`
    position: fixed;
    top: 0;
    left: 0;
    height: 3px;
    background: ${theme.colors.primary};
    box-shadow: 0 0 10px ${theme.colors.primary}, 0 0 5px ${theme.colors.primary};
    transition: width 0.2s ease, opacity 0.4s ease;
    z-index: 9999;
    width: 0%;
    opacity: 1;
  `

  start() {
    if (this.isLoading) return
    
    this.isLoading = true
    this.progress = 0
    
    // 创建加载条元素
    if (!this.element) {
      this.element = document.createElement('div')
      this.element.className = this.loaderClass
      document.body.appendChild(this.element)
    }
    
    this.element.style.width = '0%'
    this.element.style.opacity = '1'
    
    // 模拟加载进度
    this.timer = window.setInterval(() => {
      this.progress += Math.random() * 10
      
      // 限制最大进度为 90%，等待实际完成
      if (this.progress >= 90) {
        this.progress = 90
        if (this.timer) {
          clearInterval(this.timer)
          this.timer = null
        }
      }
      
      if (this.element) {
        this.element.style.width = `${this.progress}%`
      }
    }, 200)
  }

  finish() {
    if (!this.isLoading) return
    
    this.isLoading = false
    
    if (this.timer) {
      clearInterval(this.timer)
      this.timer = null
    }
    
    // 完成动画
    if (this.element) {
      this.element.style.width = '100%'
      
      // 延迟移除
      setTimeout(() => {
        if (this.element) {
          this.element.style.opacity = '0'
          
          setTimeout(() => {
            if (this.element && this.element.parentNode) {
              this.element.parentNode.removeChild(this.element)
              this.element = null
            }
            this.progress = 0
          }, 400)
        }
      }, 200)
    }
  }

  fail() {
    if (!this.isLoading) return
    
    this.isLoading = false
    
    if (this.timer) {
      clearInterval(this.timer)
      this.timer = null
    }
    
    // 失败时变红
    if (this.element) {
      this.element.style.background = theme.colors.error
      this.element.style.boxShadow = `0 0 10px ${theme.colors.error}, 0 0 5px ${theme.colors.error}`
      
      setTimeout(() => {
        if (this.element) {
          this.element.style.opacity = '0'
          
          setTimeout(() => {
            if (this.element && this.element.parentNode) {
              this.element.parentNode.removeChild(this.element)
              this.element = null
            }
            this.progress = 0
          }, 400)
        }
      }, 800)
    }
  }
}

export const topLoader = new TopLoader()
