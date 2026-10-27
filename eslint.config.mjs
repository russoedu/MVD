// This file is YOURS. mnci writes it once and never touches it again, so
// anything you add here survives `mnci upgrade`.
//
// The rules live in ./eslint.config.mnci.mjs, which mnci DOES rewrite on every
// upgrade — so put your changes here, not there.
import mnci from './eslint.config.mnci.mjs'

export default [
  ...mnci(),
  {
    // A yt-dlp metadata dump saved as UTF-16 by PowerShell: ESLint cannot parse it as JSON.
    // Delete it and this block when nothing needs the file.
    name:    'local/utf16-metadata-dump',
    ignores: ['test_meta.json'],
  },
  {
    // The classic themes set their tokens (custom properties) on :root[data-theme=...] in
    // apps/mvd-web/src/classic-themes; the components' stylesheets in other files read them, so
    // the linter cannot see where they are defined.
    name:  'local/classic-theme-tokens',
    files: [
      'apps/mvd-web/src/classic-*/**/*.css',
      'apps/mvd-web/src/add-music-window/**/*.css',
      'apps/mvd-web/src/theme-preview/**/*.css',
      'apps/mvd-web/src/classic-app/**/*.css',
      'apps/mvd-web/src/preferences-window/**/*.css',
    ],
    rules: {
      'css/no-invalid-properties': ['error', { allowUnknownVariables: true }],
    },
  },
  {
    ignores: [
      '**/vite.config.*.timestamp*',
    ],
  },
  {
    // The React build that stage-web copies in for go:embed. Generated, and git-ignored.
    name:    'local/staged-web-build',
    ignores: ['apps/mvd/localserver/web/**'],
  },
]
