# Authentication Experience

The authentication routes continue the visual language established by the `/home` Agent research experience without changing authentication behavior.

## Structure

- Desktop uses a two-band layout: a full-height Agent orchestration scene and a quiet authentication workspace.
- The visual band contains the site brand, secure-access context, a four-stage Agent trace, and live identity-channel status.
- The workspace contains a back-to-home action, locale control, the route-specific authentication content, footer actions, and copyright.
- At widths up to 900px, the visual band becomes a compact scene above the form. At widths up to 560px, secondary brand and scene copy are removed before they can overlap.

`AuthLayout.vue` is shared by login, registration, password recovery, email verification, and provider callback routes. Its slots and route behavior remain unchanged.

## Login Form

- Email and password fields use stable 48px controls with explicit labels, restrained borders, and visible focus/error states.
- The password visibility control has localized title and accessible-label text.
- The primary sign-in action uses the homepage mint signal color and retains loading, disabled, Turnstile, agreement, and 2FA behavior.
- OAuth buttons keep their provider actions while adopting the same shape, border, and workspace color system.
- Registration and password recovery links remain in their existing routes.

## Verification

From `frontend/` run:

```bash
pnpm typecheck
pnpm exec eslint src/components/layout/AuthLayout.vue src/views/auth/LoginView.vue src/i18n/locales/zh/common.ts src/i18n/locales/en/common.ts
pnpm test:run src/views/auth/__tests__ src/components/auth/__tests__
pnpm build
```

Visual verification should cover `/login` at desktop, tablet, and mobile widths, both themes, password visibility, locale switching, OAuth-enabled and OAuth-disabled states, and a long registration form using the same layout.
