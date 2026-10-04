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

const PRODUCTS = { 'mvd-tray': 'mvd' }
const OS_NAMES = { darwin: 'macos' }
const FILE_NAME = /^(mvd|mvd-tui)_\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?_(windows|macos|linux)_(amd64|arm64|universal)\.(zip|dmg)$/

function productName (app) {
  return PRODUCTS[app] || app
}

function assetName ({ app, version, os, arch, ext }) {
  return productName(app) + '_' + version + '_' + (OS_NAMES[os] || os) + '_' + arch + '.' + ext
}

function isReleaseZip (file, app, version) {
  return file.startsWith(productName(app) + '_' + version + '_') && file.endsWith('.zip')
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
    const [os, arch] = dir.split('-')
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

module.exports = { assetName, isReleaseFileName, isReleaseZip, productName }

if (require.main === module) {
  const [command, app, os, arch, ext] = process.argv.slice(2)
  const version = process.env.VERSION || 'dev'
  if (command === 'zip' && app) zipPlatforms(app, version)
  else if (command === 'name' && app && os && arch && ext) console.log(assetName({ app, version, os, arch, ext }))
  else if (command === 'check') checkDrop()
  else fail('Usage: node tools/release-assets.cjs zip <app> | name <app> <os> <arch> <ext> | check')
}
