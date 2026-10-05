#!/usr/bin/env node
// Owned by this repository, not by mnci. It is the one place that decides how the
// files attached to a release are named:
//
//   <product>_<major>.<minor>.<patch>_<os>_<processor>.<file type>
//   mvd_0.0.9_windows_amd64.zip   mvd_0.0.9_macos_universal.dmg   mvd-tui_0.1.0_linux_arm64.zip
//
//   node tools/release-assets.cjs zip <app>                      zips dist/platforms/<app>/<os>-<arch>
//   node tools/release-assets.cjs name <app> <os> <arch> <ext>   prints one file name
//   node tools/release-assets.cjs check                          fails if a file in dist/drop breaks the rule
//
// The version is $VERSION (the release tag's version); it is "dev" when not set, and
// "dev" is not a version, so a local build is never mistaken for a release file.
'use strict'
const { existsSync, mkdirSync, readdirSync } = require('node:fs')
const { join } = require('node:path')

const OS_NAMES = { darwin: 'macos' }
const FILE_NAME = /^(?:mvd|mvd-tui)_\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?_(?:windows|macos|linux)_(?:amd64|arm64|universal)\.(?:zip|dmg)$/

// A release tag is <app>@<version>. The app name is matched up to the @, so mvd never picks up an mvd-tui tag.
function releaseTag (tags, app) {
  return tags.find(each => each.startsWith(app + '@'))
}

function assetName ({ app, version, os, arch, ext }) {
  return app + '_' + version + '_' + (OS_NAMES[os] || os) + '_' + arch + '.' + ext
}

function isReleaseZip (file, app, version) {
  return file.startsWith(app + '_' + version + '_') && file.endsWith('.zip')
}

function isReleaseFileName (file) {
  return FILE_NAME.test(file)
}

function fail (message) {
  console.error(message)
  process.exit(1)
}

function zipPlatforms (app, version) {
  const AdmZip = require('adm-zip')
  const root = join('dist', 'platforms', app)
  if (!existsSync(root)) fail(app + ': nothing built in ' + root)
  mkdirSync(join('dist', 'drop'), { recursive: true })
  for (const dir of readdirSync(root)) {
    const [os, arch] = dir.split('-', 2)
    const zip = new AdmZip()
    zip.addLocalFolder(join(root, dir))
    const name = assetName({ app, version, os, arch, ext: 'zip' })
    zip.writeZip(join('dist', 'drop', name))
    console.log(name)
  }
}

function checkDrop () {
  const drop = join('dist', 'drop')
  // go-app-<app>.zip files are the per-app packs mnci makes for its own artifact; they are never attached to a release.
  const files = existsSync(drop) ? readdirSync(drop).filter(file => !file.startsWith('go-app-')) : []
  if (files.length === 0) fail('No release files in ' + drop + '.')
  const bad = files.filter(file => !isReleaseFileName(file))
  for (const file of files) console.log((bad.includes(file) ? 'BAD  ' : 'ok   ') + file)
  if (bad.length > 0) fail(bad.length + ' file(s) in ' + drop + ' do not match <product>_<version>_<os>_<processor>.<type>.')
}

// The desktop app was released as mvd-tray@x.y.z until 0.0.10 and is mvd@x.y.z from then on. Until the
// first mvd@ tag exists, its version continues from the highest mvd-tray@ tag instead of restarting at 0.0.0.
function legacyVersion (tags, app) {
  if (app !== 'mvd') return undefined
  const versions = tags
    .map(tag => /^mvd-tray@(\d+)\.(\d+)\.(\d+)$/.exec(tag))
    .filter(Boolean)
    .map(match => match.slice(1).map(Number))
    .sort((a, b) => a[0] - b[0] || a[1] - b[1] || a[2] - b[2])
  const latest = versions.at(-1)

  return latest ? latest.join('.') : undefined
}

module.exports = { assetName, isReleaseFileName, isReleaseZip, legacyVersion, releaseTag }

if (require.main === module) {
  const [command, app, os, arch, ext] = process.argv.slice(2)
  const version = process.env.VERSION || 'dev'
  if (command === 'zip' && app) zipPlatforms(app, version)
  else if (command === 'name' && app && os && arch && ext) console.log(assetName({ app, version, os, arch, ext }))
  else if (command === 'check') checkDrop()
  else fail('Usage: node tools/release-assets.cjs zip <app> | name <app> <os> <arch> <ext> | check')
}
