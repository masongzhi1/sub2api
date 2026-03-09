<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.welcomeBack') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ isAPIKeyMode ? t('auth.apiKeyLoginHint') : t('auth.passwordLoginHint') }}
        </p>
      </div>

      <div class="rounded-2xl border border-gray-200 bg-gray-50 p-1 dark:border-dark-700 dark:bg-dark-800/80">
        <div class="grid grid-cols-2 gap-1">
          <button
            type="button"
            class="inline-flex min-h-[44px] items-center justify-center rounded-xl px-4 py-2 text-sm font-medium transition-colors"
            :class="isAPIKeyMode
              ? 'bg-primary-600 text-white shadow-sm'
              : 'text-gray-600 hover:bg-white hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white'"
            :aria-pressed="isAPIKeyMode"
            @click="switchLoginMode('api_key')"
          >
            <Icon name="key" size="md" class="mr-2" />
            {{ t('auth.apiKeyTab') }}
          </button>
          <button
            type="button"
            class="inline-flex min-h-[44px] items-center justify-center rounded-xl px-4 py-2 text-sm font-medium transition-colors"
            :class="!isAPIKeyMode
              ? 'bg-primary-600 text-white shadow-sm'
              : 'text-gray-600 hover:bg-white hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white'"
            :aria-pressed="!isAPIKeyMode"
            @click="switchLoginMode('password')"
          >
            <Icon name="mail" size="md" class="mr-2" />
            {{ t('auth.passwordTab') }}
          </button>
        </div>
      </div>

      <LinuxDoOAuthSection
        v-if="!isAPIKeyMode && linuxdoOAuthEnabled"
        :disabled="isLoading"
      />

      <form @submit.prevent="handleLogin" class="space-y-5">
        <div v-if="isAPIKeyMode">
          <label for="api_key" class="input-label">
            {{ t('auth.apiKeyLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="key" size="md" class="text-gray-400 dark:text-dark-500" />
            </div>
            <input
              id="api_key"
              v-model="formData.apiKey"
              type="text"
              autocomplete="off"
              :disabled="isLoading"
              class="input pl-11 font-mono"
              :class="{ 'input-error': errors.apiKey }"
              :placeholder="t('auth.apiKeyPlaceholder')"
            />
          </div>
          <p v-if="errors.apiKey" class="input-error-text">
            {{ errors.apiKey }}
          </p>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            {{ t('auth.apiKeyLoginNote') }}
          </p>
        </div>

        <template v-else>
          <div>
            <label for="email" class="input-label">
              {{ t('auth.emailLabel') }}
            </label>
            <div class="relative">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
                <Icon name="mail" size="md" class="text-gray-400 dark:text-dark-500" />
              </div>
              <input
                id="email"
                v-model="formData.email"
                type="email"
                autocomplete="email"
                :disabled="isLoading"
                class="input pl-11"
                :class="{ 'input-error': errors.email }"
                :placeholder="t('auth.emailPlaceholder')"
              />
            </div>
            <p v-if="errors.email" class="input-error-text">
              {{ errors.email }}
            </p>
          </div>

          <div>
            <label for="password" class="input-label">
              {{ t('auth.passwordLabel') }}
            </label>
            <div class="relative">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
                <Icon name="lock" size="md" class="text-gray-400 dark:text-dark-500" />
              </div>
              <input
                id="password"
                v-model="formData.password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                :disabled="isLoading"
                class="input pl-11 pr-11"
                :class="{ 'input-error': errors.password }"
                :placeholder="t('auth.passwordPlaceholder')"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-dark-300"
              >
                <Icon v-if="showPassword" name="eyeOff" size="md" />
                <Icon v-else name="eye" size="md" />
              </button>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <p v-if="errors.password" class="input-error-text">
                {{ errors.password }}
              </p>
              <span v-else></span>
              <router-link
                v-if="passwordResetEnabled"
                to="/forgot-password"
                class="text-sm font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
              >
                {{ t('auth.forgotPassword') }}
              </router-link>
            </div>
          </div>
        </template>

        <div v-if="turnstileEnabled && turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
          <p v-if="errors.turnstile" class="input-error-text mt-2 text-center">
            {{ errors.turnstile }}
          </p>
        </div>

        <transition name="fade">
          <div
            v-if="errorMessage"
            class="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-800/50 dark:bg-red-900/20"
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="text-red-500" />
              </div>
              <p class="text-sm text-red-700 dark:text-red-400">
                {{ errorMessage }}
              </p>
            </div>
          </div>
        </transition>

        <button
          type="submit"
          :disabled="isLoading || (turnstileEnabled && !turnstileToken)"
          class="btn btn-primary w-full"
        >
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-white"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else :name="isAPIKeyMode ? 'key' : 'login'" size="md" class="mr-2" />
          {{ isLoading ? t('auth.signingIn') : (isAPIKeyMode ? t('auth.signInWithApiKey') : t('auth.signIn')) }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('auth.dontHaveAccount') }}
        <router-link
          to="/register"
          class="font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
        >
          {{ t('auth.signUp') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>

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
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings, isTotp2FARequired } from '@/api/auth'
import type { LoginRequest, TotpLoginResponse } from '@/types'

const { t } = useI18n()

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

type LoginMode = 'api_key' | 'password'

const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const loginMode = ref<LoginMode>('api_key')

const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)
const passwordResetEnabled = ref<boolean>(false)

const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  apiKey: '',
  email: '',
  password: ''
})

const errors = reactive({
  apiKey: '',
  email: '',
  password: '',
  turnstile: ''
})

const isAPIKeyMode = computed(() => loginMode.value === 'api_key')

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
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    passwordResetEnabled.value = settings.password_reset_enabled
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

function switchLoginMode(mode: LoginMode): void {
  if (loginMode.value === mode) {
    return
  }

  loginMode.value = mode
  errorMessage.value = ''
  errors.apiKey = ''
  errors.email = ''
  errors.password = ''
  showPassword.value = false
}

function validateForm(): boolean {
  errors.apiKey = ''
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (isAPIKeyMode.value) {
    if (!formData.apiKey.trim()) {
      errors.apiKey = t('auth.apiKeyRequired')
      isValid = false
    }
  } else {
    if (!formData.email.trim()) {
      errors.email = t('auth.emailRequired')
      isValid = false
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = t('auth.invalidEmail')
      isValid = false
    }

    if (!formData.password) {
      errors.password = t('auth.passwordRequired')
      isValid = false
    } else if (formData.password.length < 6) {
      errors.password = t('auth.passwordMinLength')
      isValid = false
    }
  }

  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

async function handleLogin(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    const payload: LoginRequest = isAPIKeyMode.value
      ? {
          api_key: formData.apiKey.trim(),
          turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
        }
      : {
          email: formData.email.trim(),
          password: formData.password,
          turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
        }

    const response = await authStore.login(payload)

    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    appStore.showSuccess(t('auth.loginSuccess'))

    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    const err = error as { message?: string; response?: { data?: { detail?: string } } }

    if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.loginFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)
    show2FAModal.value = false
    appStore.showSuccess(t('auth.loginSuccess'))

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
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
