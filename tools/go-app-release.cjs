#!/usr/bin/env node
// Written by mnci: 'mnci add go-app --release' creates it and 'mnci upgrade'
// rewrites it, so local edits do not survive an upgrade.
//
// A releasable Go app (tag release:go) is versioned from its git tag, with no
// manifest, and its per-platform zips are attached to its GitHub Release:
//   release.version.versionActions   this file, loaded by 'nx release'
//   node tools/go-app-release.cjs assets            the apps built for all six platforms
//   node tools/go-app-release.cjs assets --native   the apps that need a C toolchain, this OS only
'use strict'
const { spawnSync } = require('node:child_process')
const { existsSync, readdirSync, readFileSync } = require('node:fs')
const { dirname, join } = require('node:path')
const { VersionActions } = require('nx/release')

const FIRST_RELEASE_BASE = '0.0.0'

class GoAppVersionActions extends VersionActions {
  validManifestFilenames = null

  async validate () {
    // Nothing on disk to check: there is no manifest.
  }

  async readCurrentVersionFromSourceManifest () {
    // Reached only as the fallback of the git-tag resolver, i.e. before the first tag.
    return { currentVersion: FIRST_RELEASE_BASE, manifestPath: 'none (a Go app is versioned by its git tag)' }
  }

  async readCurrentVersionFromRegistry () {
    return null
  }

  async readCurrentVersionOfDependency () {
    return { currentVersion: null, dependencyCollection: null }
  }

  async updateProjectVersion () {
    return []
  }

  async updateProjectDependencies () {
    return []
  }
}

module.exports = GoAppVersionActions

function fail (message) {
  console.error(message)
  process.exit(1)
}

function run (command, args, env) {
  const result = spawnSync(command, args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'inherit'], env: { ...process.env, ...env }, shell: false })
  if (result.status !== 0) fail(command + ' ' + args.join(' ') + ' failed (exit ' + result.status + ')')

  return result.stdout
}

function nx (args, env) {
  const manifest = require.resolve('nx/package.json')
  const bin = require(manifest).bin
  const entry = typeof bin === 'string' ? bin : bin.nx

  return run(process.execPath, [join(dirname(manifest), entry), ...args], env)
}

function releasableApps (native) {
  const names = existsSync('apps') ? readdirSync('apps') : []

  return names.filter(name => {
    try {
      const tags = JSON.parse(readFileSync(join('apps', name, 'project.json'), 'utf8')).tags || []

      return tags.includes('release:go') && tags.includes('build:cgo') === native
    } catch {
      return false
    }
  })
}

function attachAssets (native) {
  const apps = releasableApps(native)
  if (apps.length === 0) {
    console.log('No releasable ' + (native ? 'native ' : '') + 'Go app - nothing to attach.')

    return
  }
  const target = native ? 'package-native' : 'package-all'
  const tags = run('git', ['tag', '--points-at', 'HEAD']).split(/\r?\n/).filter(Boolean)
  let attached = 0
  for (const app of apps) {
    const tag = tags.find(each => each.startsWith(app + '@'))
    if (!tag) {
      console.log(app + ': not released by this run - skipping.')
      continue
    }
    const version = tag.slice(app.length + 1)
    console.log(app + ': ' + (native ? 'building for this OS' : 'building the six platforms') + ' as ' + version)
    nx(['run', app + ':' + target], { VERSION: version })
    const prefix = 'go-app-' + app + '-'
    const zips = existsSync('dist/drop') ? readdirSync('dist/drop').filter(each => each.startsWith(prefix) && each.endsWith('.zip')) : []
    if (zips.length === 0) fail(app + ': ' + target + ' produced no ' + prefix + '*.zip in dist/drop.')
    const upload = spawnSync('gh', ['release', 'upload', tag, ...zips.map(each => join('dist/drop', each)), '--clobber'], { stdio: 'inherit', shell: process.platform === 'win32' })
    if (upload.status !== 0) fail(app + ': could not attach the zips to the ' + tag + ' release (exit ' + upload.status + ').')
    console.log(app + ': attached ' + zips.length + ' zips to ' + tag)
    attached += 1
  }
  console.log(attached + ' release(s) carry their platform zips.')
}

if (require.main === module) {
  if (process.argv[2] !== 'assets' || (process.argv[3] !== undefined && process.argv[3] !== '--native')) fail('Usage: node tools/go-app-release.cjs assets [--native]')
  attachAssets(process.argv[3] === '--native')
}
