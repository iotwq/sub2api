<template>
  <AuthLayout>
    <div class="login-view">
      <div class="login-heading">
        <p class="login-kicker">{{ t('auth.access.workspace') }}</p>
        <h2>
          {{ t('auth.welcomeBack') }}
        </h2>
        <p class="login-description">
          {{ t('auth.signInToAccount') }}
        </p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div>
          <label for="email" class="input-label login-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="login-field">
            <div class="login-field-icon">
              <Icon name="mail" size="md" />
            </div>
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="authActionDisabled"
              class="input login-input pl-11"
              :class="{ 'input-error': errors.email }"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
        </div>

        <div>
          <label for="password" class="input-label login-label">
            {{ t('auth.passwordLabel') }}
          </label>
          <div class="login-field">
            <div class="login-field-icon">
              <Icon name="lock" size="md" />
            </div>
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :disabled="authActionDisabled"
              class="input login-input pl-11 pr-11"
              :class="{ 'input-error': errors.password }"
              :placeholder="t('auth.passwordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              :disabled="authActionDisabled"
              class="login-password-toggle"
              :title="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
              :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
          <div class="login-password-meta">
            <span></span>
            <router-link
              v-if="passwordResetEnabled && !backendModeEnabled"
              to="/forgot-password"
              class="login-text-link"
            >
              {{ t('auth.forgotPassword') }}
            </router-link>
          </div>
        </div>

        <!-- Turnstile Widget -->
        <div v-if="captchaEnabled">
          <TurnstileWidget
            ref="turnstileRef"
            :turnstile-enabled="turnstileEnabled"
            :turnstile-site-key="turnstileSiteKey"
            :tencent-enabled="tencentCaptchaEnabled"
            :tencent-app-id="tencentCaptchaAppId"
            :tencent-region="tencentCaptchaRegion"
            :aliyun-enabled="aliyunCaptchaEnabled"
            :aliyun-scene-id="aliyunCaptchaSceneId"
            :aliyun-prefix="aliyunCaptchaPrefix"
            :aliyun-region="aliyunCaptchaRegion"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <button
          type="submit"
          :disabled="authActionDisabled || (turnstileEnabled && !turnstileToken)"
          class="btn login-submit w-full"
        >
          <span v-if="isLoading" class="login-spinner" aria-hidden="true"></span>
          <Icon v-else name="login" size="md" />
          {{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}
        </button>

        <LoginAgreementPrompt
          v-if="loginAgreementEnabled"
          :accepted="agreementAccepted"
          :documents="loginAgreementDocuments"
          :mode="loginAgreementMode"
          :updated-at="loginAgreementUpdatedAt"
          :visible="showAgreementModal"
          @accept="acceptLoginAgreement"
          @reject="rejectLoginAgreement"
          @open="showAgreementModal = true"
        />

        <div v-if="showPasskeyLogin || showOAuthLogin" class="login-oauth">
          <div class="login-divider">
            <div></div>
            <span>
              {{ t('auth.oauthOrContinue') }}
            </span>
            <div></div>
          </div>

          <button
            v-if="showPasskeyLogin"
            type="button"
            class="btn btn-secondary w-full"
            :disabled="authActionDisabled"
            @click="handlePasskeyLogin"
          >
            <Icon name="key" size="md" class="mr-2" />
            {{ passkeyLoading ? t('auth.passkeySigningIn') : t('auth.passkeySignIn') }}
          </button>

          <EmailOAuthButtons
            :disabled="authActionDisabled"
            :github-enabled="githubOAuthEnabled"
            :google-enabled="googleOAuthEnabled"
            :show-divider="false"
            @start="handleOAuthStart"
          />

          <LinuxDoOAuthSection
            v-if="linuxdoOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
            @start="handleOAuthStart"
          />
          <DingTalkOAuthSection
            v-if="dingtalkOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
            @start="handleOAuthStart"
          />
          <WechatOAuthSection
            v-if="wechatOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
            @start="handleOAuthStart"
          />
          <OidcOAuthSection
            v-if="oidcOAuthEnabled"
            :disabled="authActionDisabled"
            :provider-name="oidcOAuthProviderName"
            :show-divider="false"
            @start="handleOAuthStart"
          />
        </div>
      </form>
    </div>

    <!-- Footer -->
    <template v-if="!backendModeEnabled && publicSettingsLoaded && registrationEnabled" #footer>
      <p class="login-footer">
        {{ t('auth.dontHaveAccount') }}
        <router-link
          to="/register"
          class="login-footer-link"
        >
          {{ t('auth.signUp') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>

  <!-- 2FA Modal -->
  <TotpLoginModal
    v-if="show2FAModal"
    ref="totpModalRef"
    :temp-token="totpTempToken"
    :user-email-masked="totpUserEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import DingTalkOAuthSection from '@/components/auth/DingTalkOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import EmailOAuthButtons from '@/components/auth/EmailOAuthButtons.vue'
import LoginAgreementPrompt from '@/components/auth/LoginAgreementPrompt.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/CaptchaChallenge.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  buildOAuthLoginStartURL,
  getPublicSettings,
  isTotp2FARequired,
  isWeChatWebOAuthEnabled,
  startOAuthLogin,
  type OAuthLoginStart
} from '@/api/auth'
import type {
  ActionCaptchaRequestProof,
  LoginAgreementDocument,
  TotpLoginResponse
} from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { clearAllAffiliateReferralCodes } from '@/utils/oauthAffiliate'

const { t } = useI18n()
const LOGIN_AGREEMENT_STORAGE_KEY = 'sub2api_login_agreement_consent'

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const passkeyLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const publicSettingsLoaded = ref<boolean>(false)

// Public settings
const registrationEnabled = ref<boolean>(false)
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const tencentCaptchaEnabled = ref<boolean>(false)
const tencentCaptchaAppId = ref<string>('')
const tencentCaptchaRegion = ref<string>('cn')
const aliyunCaptchaEnabled = ref<boolean>(false)
const aliyunCaptchaSceneId = ref<string>('')
const aliyunCaptchaPrefix = ref<string>('')
const aliyunCaptchaRegion = ref<string>('cn')
const linuxdoOAuthEnabled = ref<boolean>(false)
const dingtalkOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const backendModeEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const githubOAuthEnabled = ref<boolean>(false)
const googleOAuthEnabled = ref<boolean>(false)
const passwordResetEnabled = ref<boolean>(false)
const passkeyEnabled = ref<boolean>(false)
const loginAgreementEnabled = ref<boolean>(false)
const loginAgreementMode = ref<'modal' | 'checkbox' | string>('modal')
const loginAgreementUpdatedAt = ref<string>('')
const loginAgreementRevision = ref<string>('')
const loginAgreementDocuments = ref<LoginAgreementDocument[]>([])
const agreementAccepted = ref<boolean>(false)
const showAgreementModal = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')
const tencentCaptchaRandstr = ref<string>('')
const aliyunCaptchaReady = computed(
  () =>
    aliyunCaptchaEnabled.value &&
    Boolean(aliyunCaptchaSceneId.value) &&
    Boolean(aliyunCaptchaPrefix.value)
)
// 动作触发式验证码（腾讯/阿里云）：提交、OAuth 启动、passkey 时弹窗验证
const actionCaptchaEnabled = computed(
  () =>
    (tencentCaptchaEnabled.value && Boolean(tencentCaptchaAppId.value)) ||
    aliyunCaptchaReady.value
)
const captchaEnabled = computed(
  () =>
    (turnstileEnabled.value && Boolean(turnstileSiteKey.value)) || actionCaptchaEnabled.value
)

// 2FA state
const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

const validationToastMessage = computed(
  () => errors.email || errors.password || errors.turnstile || ''
)

const agreementGateActive = computed(
  () => loginAgreementEnabled.value && !agreementAccepted.value
)

const authActionDisabled = computed(
  () => isLoading.value || passkeyLoading.value || !publicSettingsLoaded.value || agreementGateActive.value
)

const showPasskeyLogin = computed(
  () => passkeyEnabled.value && typeof window.PublicKeyCredential !== 'undefined'
)

const showOAuthLogin = computed(
  () =>
    !backendModeEnabled.value &&
    (linuxdoOAuthEnabled.value ||
      dingtalkOAuthEnabled.value ||
      wechatOAuthEnabled.value ||
      oidcOAuthEnabled.value ||
      githubOAuthEnabled.value ||
      googleOAuthEnabled.value)
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// ==================== Lifecycle ====================

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await getPublicSettings()
    registrationEnabled.value = settings.registration_enabled === true
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    tencentCaptchaEnabled.value = settings.tencent_captcha_enabled === true
    tencentCaptchaAppId.value = settings.tencent_captcha_app_id || ''
    tencentCaptchaRegion.value = settings.tencent_captcha_region || 'cn'
    aliyunCaptchaEnabled.value = settings.aliyun_captcha_enabled === true
    aliyunCaptchaSceneId.value = settings.aliyun_captcha_scene_id || ''
    aliyunCaptchaPrefix.value = settings.aliyun_captcha_prefix || ''
    aliyunCaptchaRegion.value = settings.aliyun_captcha_region || 'cn'
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
    wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
    backendModeEnabled.value = settings.backend_mode_enabled
    oidcOAuthEnabled.value = settings.oidc_oauth_enabled
    oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    githubOAuthEnabled.value = settings.github_oauth_enabled
    googleOAuthEnabled.value = settings.google_oauth_enabled
    backendModeEnabled.value = settings.backend_mode_enabled
    passwordResetEnabled.value = settings.password_reset_enabled
    passkeyEnabled.value = settings.passkey_enabled === true
    applyLoginAgreementSettings(settings)
  } catch (error) {
    console.error('Failed to load public settings:', error)
    loginAgreementEnabled.value = false
    agreementAccepted.value = true
  } finally {
    publicSettingsLoaded.value = true
  }
})

// ==================== Login Agreement ====================

function applyLoginAgreementSettings(settings: {
  login_agreement_enabled?: boolean
  login_agreement_mode?: string
  login_agreement_updated_at?: string
  login_agreement_revision?: string
  login_agreement_documents?: LoginAgreementDocument[]
}): void {
  const documents = Array.isArray(settings.login_agreement_documents)
    ? settings.login_agreement_documents.filter((doc) => doc.title?.trim())
    : []
  loginAgreementDocuments.value = documents
  loginAgreementEnabled.value = settings.login_agreement_enabled === true && documents.length > 0
  loginAgreementMode.value = settings.login_agreement_mode === 'checkbox' ? 'checkbox' : 'modal'
  loginAgreementUpdatedAt.value = settings.login_agreement_updated_at || ''
  loginAgreementRevision.value =
    settings.login_agreement_revision ||
    `${loginAgreementUpdatedAt.value}:${documents.map((doc) => `${doc.id}:${doc.title}`).join('|')}`

  agreementAccepted.value = !loginAgreementEnabled.value || hasAcceptedLoginAgreement(loginAgreementRevision.value)
  showAgreementModal.value =
    loginAgreementEnabled.value && !agreementAccepted.value && loginAgreementMode.value !== 'checkbox'
}

function hasAcceptedLoginAgreement(revision: string): boolean {
  if (!revision) {
    return false
  }
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return false
    }
    const parsed = JSON.parse(raw) as { revision?: string }
    return parsed.revision === revision
  } catch {
    return false
  }
}

function acceptLoginAgreement(): void {
  if (loginAgreementRevision.value) {
    localStorage.setItem(
      LOGIN_AGREEMENT_STORAGE_KEY,
      JSON.stringify({
        revision: loginAgreementRevision.value,
        accepted_at: new Date().toISOString()
      })
    )
  }
  agreementAccepted.value = true
  showAgreementModal.value = false
}

function rejectLoginAgreement(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
  agreementAccepted.value = false
  showAgreementModal.value = false
  appStore.showWarning(t('legal.loginAgreementPrompt.loginRejectedWarning'))
}

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string, randstr = ''): void {
  turnstileToken.value = token
  tencentCaptchaRandstr.value = randstr
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  tencentCaptchaRandstr.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  tencentCaptchaRandstr.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

function resetCaptchaProof(): void {
  turnstileRef.value?.reset()
  turnstileToken.value = ''
  tencentCaptchaRandstr.value = ''
  errors.turnstile = ''
}

async function acquireActionProof(): Promise<boolean> {
  if (!actionCaptchaEnabled.value) return true

  const proof = await turnstileRef.value?.verifyAction()
  if (!proof) return false

  turnstileToken.value = proof.token
  tencentCaptchaRandstr.value = proof.randstr
  return true
}

// ==================== Validation ====================

function validateForm(): boolean {
  // Reset errors
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (agreementGateActive.value) {
    appStore.showWarning(t('legal.loginAgreementPrompt.loginRequiredWarning'))
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return false
  }

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Turnstile validation
  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleLogin(): Promise<void> {
  // Clear previous error
  errorMessage.value = ''

  // Validate form
  if (!validateForm()) {
    return
  }

  if (!(await acquireActionProof())) {
    return
  }

  isLoading.value = true

  try {
    // Call auth store login（阿里云 captchaVerifyParam 复用 turnstile_token 字段）
    const response = await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token:
        turnstileEnabled.value || aliyunCaptchaEnabled.value ? turnstileToken.value : undefined,
      tencent_captcha_ticket: tencentCaptchaEnabled.value ? turnstileToken.value : undefined,
      tencent_captcha_randstr: tencentCaptchaEnabled.value
        ? tencentCaptchaRandstr.value
        : undefined
    })

    // Check if 2FA is required
    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    // Show success toast
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    errorMessage.value = extractI18nErrorMessage(error, t, 'auth.errors', t('auth.loginFailed'))

    // Also show error toast
    appStore.showError(errorMessage.value)
  } finally {
    if (captchaEnabled.value) {
      resetCaptchaProof()
    }
    isLoading.value = false
  }
}

async function handlePasskeyLogin(): Promise<void> {
  if (agreementGateActive.value) {
    appStore.showWarning(t('legal.loginAgreementPrompt.loginRequiredWarning'))
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return
  }

  passkeyLoading.value = true
  try {
    let proof: ActionCaptchaRequestProof | undefined
    if (actionCaptchaEnabled.value) {
      const result = await turnstileRef.value?.verifyAction()
      if (!result) return
      proof = tencentCaptchaEnabled.value
        ? {
            tencent_captcha_ticket: result.token,
            tencent_captcha_randstr: result.randstr
          }
        : { turnstile_token: result.token }
    }

    await authStore.loginWithPasskey(proof)
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const fallback = error instanceof DOMException && error.name === 'NotAllowedError'
      ? t('auth.passkeyCancelled')
      : t('auth.passkeyFailed')
    errorMessage.value = extractI18nErrorMessage(error, t, 'auth.errors', fallback)
    appStore.showError(errorMessage.value)
  } finally {
    if (actionCaptchaEnabled.value) {
      resetCaptchaProof()
    }
    passkeyLoading.value = false
  }
}

async function handleOAuthStart(request: OAuthLoginStart): Promise<void> {
  if (authActionDisabled.value) return

  if (!actionCaptchaEnabled.value) {
    window.location.href = buildOAuthLoginStartURL(request)
    return
  }

  isLoading.value = true
  try {
    const proof = await turnstileRef.value?.verifyAction()
    if (!proof) return

    const result = await startOAuthLogin(
      request,
      tencentCaptchaEnabled.value
        ? {
            tencent_captcha_ticket: proof.token,
            tencent_captcha_randstr: proof.randstr
          }
        : { turnstile_token: proof.token }
    )
    window.location.href = result.authorize_url
  } catch (error: unknown) {
    errorMessage.value = extractI18nErrorMessage(
      error,
      t,
      'auth.errors',
      t('auth.turnstileFailed')
    )
    appStore.showError(errorMessage.value)
  } finally {
    resetCaptchaProof()
    isLoading.value = false
  }
}

// ==================== 2FA Handlers ====================

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)

    // Close modal and show success
    show2FAModal.value = false
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('profile.totp.loginFailed')

    if (totpModalRef.value) {
      totpModalRef.value.setError(message)
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  show2FAModal.value = false
  totpTempToken.value = ''
  totpUserEmailMasked.value = ''
}
</script>

<style scoped>
.login-view {
  display: grid;
  gap: 28px;
  animation: login-view-in 560ms 180ms both cubic-bezier(0.22, 1, 0.36, 1);
}

.login-view,
.login-view * {
  letter-spacing: 0;
}

.login-heading {
  text-align: left;
}

.login-heading > * {
  animation: login-content-in 520ms both ease-out;
}

.login-heading h2 { animation-delay: 70ms; }
.login-heading .login-description { animation-delay: 120ms; }

.login-kicker {
  margin: 0 0 10px;
  color: #238c73;
  font-size: 9px;
  font-weight: 900;
}

.login-heading h2 {
  margin: 0;
  color: #111613;
  font-size: 34px;
  font-weight: 850;
  line-height: 1.1;
}

.login-description {
  margin: 10px 0 0;
  color: #69746d;
  font-size: 14px;
  line-height: 1.6;
}

.login-form {
  display: grid;
  gap: 20px;
}

.login-form > * {
  animation: login-content-in 520ms both ease-out;
}

.login-form > :nth-child(1) { animation-delay: 150ms; }
.login-form > :nth-child(2) { animation-delay: 200ms; }
.login-form > :nth-child(3) { animation-delay: 250ms; }
.login-form > :nth-child(4) { animation-delay: 300ms; }
.login-form > :nth-child(5) { animation-delay: 350ms; }

.login-label {
  margin-bottom: 7px;
  color: #3f4a43;
  font-size: 12px;
  font-weight: 750;
}

.login-field {
  position: relative;
}

.login-field::before {
  position: absolute;
  z-index: 1;
  top: 8px;
  bottom: 8px;
  left: 0;
  width: 2px;
  border-radius: 2px;
  background: #238c73;
  content: '';
  opacity: 0;
  transform: scaleY(0.5);
  transition: opacity 160ms ease, transform 160ms ease;
}

.login-field:focus-within::before {
  opacity: 1;
  transform: scaleY(1);
}

.login-field-icon {
  position: absolute;
  z-index: 2;
  top: 0;
  bottom: 0;
  left: 0;
  display: flex;
  align-items: center;
  padding-left: 14px;
  pointer-events: none;
  color: #718078;
  transition: color 160ms ease, transform 160ms ease;
}

.login-field:focus-within .login-field-icon {
  color: #238c73;
  transform: scale(1.05);
}

.login-input {
  height: 48px;
  border: 1px solid #c9d2cc;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.72);
  color: #111613;
  font-size: 14px;
  box-shadow: none;
}

.login-input::placeholder {
  color: #98a29c;
}

.login-input:hover:not(:disabled) {
  border-color: #9ba9a1;
}

.login-input:focus {
  border-color: #238c73;
  outline: none;
  box-shadow: 0 0 0 3px rgba(35, 140, 115, 0.13);
}

.login-input.input-error {
  border-color: #d95c52;
  box-shadow: 0 0 0 3px rgba(217, 92, 82, 0.1);
}

.login-input:disabled {
  background: rgba(220, 226, 222, 0.7);
  color: #77817b;
}

.login-password-toggle {
  position: absolute;
  z-index: 2;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  width: 44px;
  align-items: center;
  justify-content: center;
  color: #718078;
  transition: color 160ms ease;
}

.login-password-toggle:hover:not(:disabled),
.login-password-toggle:focus-visible {
  color: #1a5f4f;
  outline: none;
}

.login-password-meta {
  display: flex;
  min-height: 22px;
  align-items: center;
  justify-content: space-between;
  margin-top: 5px;
}

.login-text-link,
.login-footer-link {
  color: #16765f;
  font-size: 12px;
  font-weight: 800;
  text-decoration: none;
}

.login-text-link:hover,
.login-text-link:focus-visible,
.login-footer-link:hover,
.login-footer-link:focus-visible {
  color: #0b4f3e;
  text-decoration: underline;
  text-underline-offset: 3px;
  outline: none;
}

.login-submit {
  min-height: 48px;
  border: 1px solid #72f5d5;
  border-radius: 4px;
  background: #72f5d5;
  background-image: none;
  color: #07110d;
  font-size: 13px;
  font-weight: 850;
  box-shadow: 0 8px 20px rgba(35, 140, 115, 0.16);
  transition: transform 160ms ease, box-shadow 160ms ease, background 160ms ease;
}

.login-submit:hover:not(:disabled),
.login-submit:focus-visible {
  border-color: #58dfbf;
  background: #58dfbf;
  background-image: none;
  box-shadow: 0 10px 24px rgba(35, 140, 115, 0.22);
  outline: none;
  transform: translateY(-1px);
}

.login-submit:focus-visible {
  box-shadow: 0 0 0 3px rgba(35, 140, 115, 0.18);
}

.login-submit:disabled {
  border-color: #b9c8c0;
  background: #cbd6d0;
  color: #6b7770;
}

.login-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(7, 17, 13, 0.24);
  border-top-color: #07110d;
  border-radius: 50%;
  animation: login-spin 0.8s linear infinite;
}

.login-oauth {
  display: grid;
  gap: 12px;
  padding-top: 2px;
}

.login-divider {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login-divider div {
  height: 1px;
  flex: 1;
  background: #cfd7d2;
}

.login-divider span {
  color: #7c8780;
  font-size: 10px;
}

.login-oauth :deep(.btn-secondary) {
  min-height: 46px;
  border: 1px solid #c9d2cc;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.54);
  color: #2e3932;
  box-shadow: 0 4px 12px rgba(17, 22, 19, 0.04);
  transition: transform 160ms ease, border-color 160ms ease, background 160ms ease, box-shadow 160ms ease;
}

.login-oauth :deep(.btn-secondary:hover:not(:disabled)),
.login-oauth :deep(.btn-secondary:focus-visible) {
  border-color: #91a198;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 7px 16px rgba(17, 22, 19, 0.08);
  outline: none;
  transform: translateY(-1px);
}

.login-footer {
  margin: 0;
  color: #69746d;
  font-size: 13px;
}

:global(.dark) .login-kicker {
  color: #72f5d5;
}

:global(.dark) .login-heading h2 {
  color: #f2f6f3;
}

:global(.dark) .login-description,
:global(.dark) .login-footer {
  color: #8f9b94;
}

:global(.dark) .login-label {
  color: #c1cbc5;
}

:global(.dark) .login-input {
  border-color: #303b35;
  background: #111613;
  color: #f0f5f1;
}

:global(.dark) .login-input:hover:not(:disabled) {
  border-color: #506159;
}

:global(.dark) .login-input:focus {
  border-color: #72f5d5;
  box-shadow: 0 0 0 3px rgba(114, 245, 213, 0.1);
}

:global(.dark) .login-field-icon,
:global(.dark) .login-password-toggle {
  color: #77857d;
}

:global(.dark) .login-field:focus-within .login-field-icon {
  color: #72f5d5;
}

:global(.dark) .login-field::before {
  background: #72f5d5;
}

:global(.dark) .login-password-toggle:hover:not(:disabled),
:global(.dark) .login-password-toggle:focus-visible {
  color: #72f5d5;
}

:global(.dark) .login-text-link,
:global(.dark) .login-footer-link {
  color: #72f5d5;
}

:global(.dark) .login-divider div {
  background: #303b35;
}

:global(.dark) .login-divider span {
  color: #748078;
}

:global(.dark) .login-oauth :deep(.btn-secondary) {
  border-color: #303b35;
  background: #111613;
  color: #d9e2dc;
}

:global(.dark) .login-oauth :deep(.btn-secondary:hover:not(:disabled)),
:global(.dark) .login-oauth :deep(.btn-secondary:focus-visible) {
  border-color: #52645b;
  background: #171e1a;
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.24);
}

@keyframes login-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes login-view-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes login-content-in {
  from { opacity: 0; transform: translateY(9px); }
  to { opacity: 1; transform: translateY(0); }
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@media (max-width: 560px) {
  .login-view {
    gap: 24px;
  }

  .login-heading h2 {
    font-size: 29px;
  }

  .login-form {
    gap: 17px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .login-view,
  .login-heading > *,
  .login-form > *,
  .login-field::before {
    animation: none;
    transition: none;
  }

  .login-spinner {
    animation-duration: 1.6s;
  }
}
</style>
