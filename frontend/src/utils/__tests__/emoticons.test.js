import { describe, it, expect } from 'vitest'
import { detectEmoticon, detectEmojiShortcode, EMOTICONS, QUICK_REACTION_EMOJIS, EMOJI_SHORTCODES } from '../emoticons'

describe('EMOTICONS', () => {
  it('has entries sorted longest-first', () => {
    for (let i = 1; i < EMOTICONS.length; i++) {
      expect(EMOTICONS[i - 1][0].length >= EMOTICONS[i][0].length).toBe(true)
    }
  })
})

describe('QUICK_REACTION_EMOJIS', () => {
  it('has 7 default reactions', () => {
    expect(QUICK_REACTION_EMOJIS).toHaveLength(7)
  })
})

describe('EMOJI_SHORTCODES', () => {
  it('creates underscore-less aliases for all underscore names', () => {
    let found = 0
    for (const key of Object.keys(EMOJI_SHORTCODES)) {
      if (key.includes('_')) {
        const alias = key.replace(/_/g, '')
        expect(EMOJI_SHORTCODES[alias]).toBeDefined()
        found++
      }
    }
    expect(found).toBeGreaterThan(0)
  })
})

describe('detectEmoticon', () => {
  it('detects emoticon at end of text', () => {
    expect(detectEmoticon('Hello :)')).toEqual({ pattern: ':)', emoji: '😊' })
    expect(detectEmoticon('😎 B-)')).toEqual({ pattern: 'B-)', emoji: '😎' })
  })

  it('returns null when no emoticon found', () => {
    expect(detectEmoticon('Hello world')).toBeNull()
  })

  it('returns null when emoticon is mid-word', () => {
    expect(detectEmoticon('Hello:)')).toBeNull()
  })

  it('detects multi-char emoticons greedily', () => {
    expect(detectEmoticon('O:-)')).toEqual({ pattern: 'O:-)', emoji: '😇' })
  })

  it('detects heart emoticons', () => {
    expect(detectEmoticon('I love you <3')).toEqual({ pattern: '<3', emoji: '❤️' })
    expect(detectEmoticon('broken </3')).toEqual({ pattern: '</3', emoji: '💔' })
  })
})

describe('detectEmojiShortcode', () => {
  it('detects shortcode at end of text', () => {
    expect(detectEmojiShortcode('Hello :smile:')).toEqual({ pattern: ':smile:', emoji: '😄' })
  })

  it('returns null when no shortcode found', () => {
    expect(detectEmojiShortcode('Hello world')).toBeNull()
    expect(detectEmojiShortcode('Hello :unknown:')).toBeNull()
  })

  it('returns null when shortcode is mid-word', () => {
    expect(detectEmojiShortcode('Hello:smile:')).toBeNull()
  })
})
