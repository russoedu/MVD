#!/usr/bin/env node
// Downloads VideoCraw's music-fixer for this machine's GOOS/GOARCH into
// libs/mvd-core/musicfixer/bundle, where the Go build embeds it. Without a token or a
// release the build goes on without it, and the feature is off in that build.
'use strict'
const { spawnSync } = require('node:child_process')
const { mkdirSync } = require('node:fs')
const { join } = require('node:path')

const REPO = 'Open-Craw/VideoCraw'
const goenv = k => spawnSync('go', ['env', k], { encoding: 'utf8' }).stdout.trim()
const name = `music-fixer-${goenv('GOOS')}-${goenv('GOARCH')}.zip`
const dir = join(__dirname, '..', 'libs', 'mvd-core', 'musicfixer', 'bundle')

if (!process.env.VIDEOCRAW_TOKEN) {
  console.log(`fetch-music-fixer: no VIDEOCRAW_TOKEN, building without ${name}`)
  process.exit(0)
}
mkdirSync(dir, { recursive: true })
const tag = process.env.MUSIC_FIXER_TAG ? [process.env.MUSIC_FIXER_TAG] : []
const r = spawnSync('gh', ['release', 'download', ...tag, '-R', REPO, '-p', name, '-D', dir, '--clobber'], {
  stdio: 'inherit', env: { ...process.env, GH_TOKEN: process.env.VIDEOCRAW_TOKEN }
})
console.log(r.status === 0 ? `fetch-music-fixer: ${name} is in the bundle` : `fetch-music-fixer: no ${name} in ${REPO}'s latest release, building without it`)
process.exit(0)
