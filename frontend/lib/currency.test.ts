import { describe, expect, it } from 'vitest'
import { formatYen } from './currency'

describe('formatYen', () => {
  it('3桁区切りと円記号付きの文字列を返す', () => {
    expect(formatYen(1000)).toBe('￥1,000')
  })

  it('0円も正しくフォーマットする', () => {
    expect(formatYen(0)).toBe('￥0')
  })
})
