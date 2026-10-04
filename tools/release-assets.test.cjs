'use strict'
const test = require('node:test')
const assert = require('node:assert/strict')
const { assetName, isReleaseFileName, isReleaseZip, productName } = require('./release-assets.cjs')

test('names a file <product>_<version>_<os>_<processor>.<type>', () => {
  const cases = [
    [{ app: 'mvd-tray', version: '0.0.9', os: 'windows', arch: 'amd64', ext: 'zip' }, 'mvd_0.0.9_windows_amd64.zip'],
    [{ app: 'mvd-tray', version: '0.0.9', os: 'linux', arch: 'amd64', ext: 'zip' }, 'mvd_0.0.9_linux_amd64.zip'],
    [{ app: 'mvd-tray', version: '0.0.9', os: 'darwin', arch: 'arm64', ext: 'zip' }, 'mvd_0.0.9_macos_arm64.zip'],
    [{ app: 'mvd-tray', version: '0.0.9', os: 'darwin', arch: 'universal', ext: 'dmg' }, 'mvd_0.0.9_macos_universal.dmg'],
    [{ app: 'mvd-tui', version: '0.1.0', os: 'darwin', arch: 'amd64', ext: 'zip' }, 'mvd-tui_0.1.0_macos_amd64.zip'],
    [{ app: 'mvd-tui', version: '1.2.3-rc.1', os: 'windows', arch: 'arm64', ext: 'zip' }, 'mvd-tui_1.2.3-rc.1_windows_arm64.zip'],
  ]
  for (const [input, want] of cases) assert.equal(assetName(input), want)
})

test('never writes darwin in a file name', () => {
  assert.doesNotMatch(assetName({ app: 'mvd-tui', version: '1.0.0', os: 'darwin', arch: 'arm64', ext: 'zip' }), /darwin/)
})

test('the desktop app is called mvd and the terminal app keeps its name', () => {
  assert.equal(productName('mvd-tray'), 'mvd')
  assert.equal(productName('mvd-tui'), 'mvd-tui')
})

test('accepts the convention and rejects everything else', () => {
  for (const good of ['mvd_0.0.9_windows_amd64.zip', 'mvd_10.20.30_macos_universal.dmg', 'mvd-tui_0.1.0_linux_arm64.zip', 'mvd_1.0.0-rc.1_linux_amd64.zip']) {
    assert.ok(isReleaseFileName(good), good)
  }
  for (const bad of ['go-app-mvd-tray-windows-amd64.zip', 'MVD.dmg', 'mvd_dev_linux_amd64.zip', 'mvd_0.0.9_darwin_arm64.zip', 'mvd_0.0.9_linux_x86.zip', 'mvd_0.0.9_linux_amd64.tar.gz', 'mvd-cli_0.0.9_linux_amd64.zip', 'mvd_0.9_linux_amd64.zip']) {
    assert.ok(!isReleaseFileName(bad), bad)
  }
})

test('picks only the zips of that app and that version to attach', () => {
  assert.ok(isReleaseZip('mvd_0.0.9_windows_amd64.zip', 'mvd-tray', '0.0.9'))
  assert.ok(!isReleaseZip('mvd_0.0.8_windows_amd64.zip', 'mvd-tray', '0.0.9'))
  assert.ok(!isReleaseZip('mvd-tui_0.0.9_windows_amd64.zip', 'mvd-tray', '0.0.9'))
  assert.ok(!isReleaseZip('go-app-mvd-tray-windows-amd64.zip', 'mvd-tray', '0.0.9'))
  assert.ok(!isReleaseZip('mvd_0.0.9_macos_universal.dmg', 'mvd-tray', '0.0.9'))
})
