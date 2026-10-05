'use strict'
const test = require('node:test')
const assert = require('node:assert/strict')
const { assetName, isReleaseFileName, isReleaseZip, legacyVersion, releaseTag } = require('./release-assets.cjs')

test('names a file <product>_<version>_<os>_<processor>.<type>', () => {
  const cases = [
    [{ app: 'mvd', version: '0.0.9', os: 'windows', arch: 'amd64', ext: 'zip' }, 'mvd_0.0.9_windows_amd64.zip'],
    [{ app: 'mvd', version: '0.0.9', os: 'linux', arch: 'amd64', ext: 'zip' }, 'mvd_0.0.9_linux_amd64.zip'],
    [{ app: 'mvd', version: '0.0.9', os: 'darwin', arch: 'arm64', ext: 'zip' }, 'mvd_0.0.9_macos_arm64.zip'],
    [{ app: 'mvd', version: '0.0.9', os: 'darwin', arch: 'universal', ext: 'dmg' }, 'mvd_0.0.9_macos_universal.dmg'],
    [{ app: 'mvd-tui', version: '0.1.0', os: 'darwin', arch: 'amd64', ext: 'zip' }, 'mvd-tui_0.1.0_macos_amd64.zip'],
    [{ app: 'mvd-tui', version: '1.2.3-rc.1', os: 'windows', arch: 'arm64', ext: 'zip' }, 'mvd-tui_1.2.3-rc.1_windows_arm64.zip'],
  ]
  for (const [input, want] of cases) assert.equal(assetName(input), want)
})

test('never writes darwin in a file name', () => {
  assert.doesNotMatch(assetName({ app: 'mvd-tui', version: '1.0.0', os: 'darwin', arch: 'arm64', ext: 'zip' }), /darwin/)
})

test('the file name starts with the app name, so mvd and mvd-tui files never mix', () => {
  assert.ok(assetName({ app: 'mvd', version: '0.0.9', os: 'linux', arch: 'amd64', ext: 'zip' }).startsWith('mvd_'))
  assert.ok(assetName({ app: 'mvd-tui', version: '0.0.9', os: 'linux', arch: 'amd64', ext: 'zip' }).startsWith('mvd-tui_'))
})

test('finds the release tag of each app and never confuses mvd with mvd-tui', () => {
  const tags = ['mvd-tui@0.0.1', 'mvd@0.0.11']
  assert.equal(releaseTag(tags, 'mvd'), 'mvd@0.0.11')
  assert.equal(releaseTag(tags, 'mvd-tui'), 'mvd-tui@0.0.1')
  assert.equal(releaseTag(['mvd-tui@0.0.1'], 'mvd'), undefined)
  assert.equal(releaseTag(['mvd@0.0.11'], 'mvd-tui'), undefined)
  assert.equal(releaseTag(['mvd-tray@0.0.10'], 'mvd'), undefined)
})

test('accepts the convention and rejects everything else', () => {
  for (const good of ['mvd_0.0.9_windows_amd64.zip', 'mvd_10.20.30_macos_universal.dmg', 'mvd-tui_0.1.0_linux_arm64.zip', 'mvd_1.0.0-rc.1_linux_amd64.zip']) {
    assert.ok(isReleaseFileName(good), good)
  }
  for (const bad of ['go-app-mvd-windows-amd64.zip', 'MVD.dmg', 'mvd_dev_linux_amd64.zip', 'mvd_0.0.9_darwin_arm64.zip', 'mvd_0.0.9_linux_x86.zip', 'mvd_0.0.9_linux_amd64.tar.gz', 'mvd-cli_0.0.9_linux_amd64.zip', 'mvd_0.9_linux_amd64.zip']) {
    assert.ok(!isReleaseFileName(bad), bad)
  }
})

test('picks only the zips of that app and that version to attach', () => {
  assert.ok(isReleaseZip('mvd_0.0.9_windows_amd64.zip', 'mvd', '0.0.9'))
  assert.ok(!isReleaseZip('mvd_0.0.8_windows_amd64.zip', 'mvd', '0.0.9'))
  assert.ok(!isReleaseZip('mvd-tui_0.0.9_windows_amd64.zip', 'mvd', '0.0.9'))
  assert.ok(!isReleaseZip('go-app-mvd-windows-amd64.zip', 'mvd', '0.0.9'))
  assert.ok(!isReleaseZip('mvd_0.0.9_macos_universal.dmg', 'mvd', '0.0.9'))
  assert.ok(!isReleaseZip('mvd_0.0.10_linux_amd64.zip', 'mvd', '0.0.1'))
  assert.ok(isReleaseZip('mvd-tui_0.0.1_linux_amd64.zip', 'mvd-tui', '0.0.1'))
  assert.ok(!isReleaseZip('mvd-tui_0.0.1_linux_amd64.zip', 'mvd', '0.0.1'))
})

test('mvd continues from the highest old mvd-tray tag and compares numbers, not text', () => {
  assert.equal(legacyVersion(['mvd-tray@0.0.9', 'mvd-tray@0.0.10', 'mvd-tray@0.0.2'], 'mvd'), '0.0.10')
  assert.equal(legacyVersion(['mvd-tray@0.1.0', 'mvd-tray@0.0.10'], 'mvd'), '0.1.0')
})

test('only mvd uses the old tags, and nothing is invented when there are none', () => {
  assert.equal(legacyVersion(['mvd-tray@0.0.10'], 'mvd-tui'), undefined)
  assert.equal(legacyVersion([], 'mvd'), undefined)
  assert.equal(legacyVersion(['mvd@0.0.11', 'mvd-tui@0.0.1', 'mvd-tray@1.0'], 'mvd'), undefined)
})
