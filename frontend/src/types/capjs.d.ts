declare global {
  interface Window {
    Cap?: any
    CAP_CUSTOM_FETCH?: (...args: any[]) => Promise<any>
    CAP_CUSTOM_WASM_URL?: string
    CAP_CSS_NONCE?: string
    CAP_DONT_SKIP_REDEFINE?: boolean
  }

  namespace JSX {
    interface IntrinsicElements {
      'cap-widget': CapWidgetElement
    }
  }
}

interface CapWidgetElement extends HTMLElement {
  'data-cap-api-endpoint'?: string
  'data-cap-site-key'?: string
  'data-cap-worker-count'?: string
  'data-cap-i18n-initial-state'?: string
  'data-cap-i18n-verifying'?: string
  'data-cap-i18n-verified'?: string
  'data-cap-i18n-error'?: string
  token?: string
  reset?: () => void
}

export {}
