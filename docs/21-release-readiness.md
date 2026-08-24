# Release Readiness

This checklist separates repository work from console, legal, and physical-device
work. A checked item must have evidence; configuration must never be inferred.

## Repository complete

- [x] Email/password, Google OAuth, and Apple OAuth client flows use Supabase Auth.
- [x] Notes and senior-scoped care-circle messaging, including read state.
- [x] OpenAPI route contract in `docs/openapi.yaml`.
- [x] Inter typography and bundled font license.
- [x] Notification inbox, scheduler, local fallback, and Expo provider implementation.
- [x] Bundle identifiers: `app.meracare.mobile`.

## Product approval required

- [ ] Approve or reject `apps/mobile/assets/images/brand-mark-v2.png`.
- [ ] Produce deterministic icon, adaptive-icon, favicon, splash, and social-card
      exports from the approved source artwork.
- [ ] Approve screenshots and store copy.

## Console configuration required

- [ ] Enable and verify Google in Google Cloud and Supabase.
- [ ] Enable and verify Apple in Apple Developer and Supabase.
- [ ] Create the EAS project and record `extra.eas.projectId` in `app.json`.
- [ ] Configure APNs and FCM credentials in EAS; keep them out of this repository.
- [ ] Keep `PUSH_ENABLED=false` until both platforms successfully register tokens.
- [ ] Configure production API/CORS URLs and Supabase redirect allow-lists.

## Legal and store-owner input required

- [ ] Have counsel/product ownership approve `docs/privacy-policy-draft.md` and
      supply the controller/operator name, contact address, jurisdiction,
      retention periods, and account-deletion process.
- [ ] Publish the approved policy and insert its real HTTPS URL into store metadata.
- [ ] Create App Store Connect and Google Play listings; record their real URLs
      only after the stores assign them.

## Physical-device acceptance

- [ ] iPhone development/release build: email, Google, Apple, local reminders,
      push registration, delivery, and every deep link.
- [ ] Android development/release build: the same checks; do not use Expo Go for push.
- [ ] Sign out with an empty queue, a queued offline action, and a deactivated device.
- [ ] Accessibility: large text, VoiceOver/TalkBack labels, focus order, contrast,
      and 48dp targets.
