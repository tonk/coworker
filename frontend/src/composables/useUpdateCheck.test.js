import { describe, it, expect, beforeEach } from 'vitest'
import { isNewer, pickAsset } from './useUpdateCheck'

describe('isNewer', () => {
  it('returns true when latest is a higher major version', () => {
    expect(isNewer('1.0.0', '2.0.0')).toBe(true)
  })

  it('returns false when current is a higher major version', () => {
    expect(isNewer('3.0.0', '2.0.0')).toBe(false)
  })

  it('returns true when latest is a higher minor version', () => {
    expect(isNewer('1.3.0', '1.4.0')).toBe(true)
  })

  it('returns true when latest is a higher patch version', () => {
    expect(isNewer('1.0.4', '1.0.5')).toBe(true)
  })

  it('returns false when versions are equal', () => {
    expect(isNewer('1.2.3', '1.2.3')).toBe(false)
  })

  it('handles leading v prefix', () => {
    expect(isNewer('v1.0.0', 'v2.0.0')).toBe(true)
    expect(isNewer('v2.0.0', 'v1.0.0')).toBe(false)
  })

  it('handles mixed v prefix', () => {
    expect(isNewer('1.0.0', 'v2.0.0')).toBe(true)
    // 'v1.0.0' vs '2.0.0' — after stripping: '1.0.0' vs '2.0.0', so 2.0.0 is newer
    expect(isNewer('v2.0.0', '1.0.0')).toBe(false)
  })

  it('returns false when current is dev', () => {
    expect(isNewer('dev', '1.0.0')).toBe(false)
  })
})

describe('pickAsset', () => {
  const assets = [
    { name: 'WarmDesk-v1.2.3-x86_64.AppImage', browser_download_url: 'https://example.com/appimage' },
    { name: 'WarmDesk-v1.2.3-amd64.deb', browser_download_url: 'https://example.com/deb' },
    { name: 'WarmDesk-v1.2.3-x86_64.rpm', browser_download_url: 'https://example.com/rpm' },
    { name: 'WarmDesk-v1.2.3-universal.dmg', browser_download_url: 'https://example.com/dmg' },
    { name: 'WarmDesk-v1.2.3-x64-portable.zip', browser_download_url: 'https://example.com/zip' },
    { name: 'warmdesk-v1.2.3-linux-amd64.tar.gz', browser_download_url: 'https://example.com/tar' },
  ]

  beforeEach(() => {
    window.__TAURI_INTERNALS__ = {}
  })

  it('returns null when not in Tauri', () => {
    window.__TAURI_INTERNALS__ = undefined
    expect(pickAsset(assets, 'v1.2.3')).toBeNull()
  })

  it('returns null when assets is empty', () => {
    expect(pickAsset([], 'v1.2.3')).toBeNull()
  })

  it('returns null when no match found', () => {
    expect(pickAsset(assets, 'v9.9.9')).toBeNull()
  })
})
