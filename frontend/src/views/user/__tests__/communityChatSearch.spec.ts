import { describe, expect, it } from 'vitest'
import {
  centeredCommunityChatSearchScrollTop,
  hasCommunityChatSearchMatch,
  splitCommunityChatSearchText,
} from '../communityChatSearch'

describe('community chat search presentation', () => {
  it('splits every case-insensitive literal match without changing the source text', () => {
    expect(splitCommunityChatSearchText('Claude and CLAUDE use $5.00?', 'claude')).toEqual([
      { text: 'Claude', matched: true },
      { text: ' and ', matched: false },
      { text: 'CLAUDE', matched: true },
      { text: ' use $5.00?', matched: false },
    ])
    expect(splitCommunityChatSearchText('Price is $5.00?', '$5.00?')).toEqual([
      { text: 'Price is ', matched: false },
      { text: '$5.00?', matched: true },
    ])
  })

  it('reports literal matches and ignores an empty search query', () => {
    expect(hasCommunityChatSearchMatch('模型_价格 100%', '_价格')).toBe(true)
    expect(hasCommunityChatSearchMatch('模型_价格 100%', '%')).toBe(true)
    expect(hasCommunityChatSearchMatch('模型_价格 100%', '   ')).toBe(false)
  })

  it('centers the keyword inside the message list instead of the whole page', () => {
    expect(centeredCommunityChatSearchScrollTop({
      containerScrollTop: 300,
      containerTop: 200,
      containerHeight: 600,
      targetTop: 500,
      targetHeight: 40,
    })).toBe(320)

    expect(centeredCommunityChatSearchScrollTop({
      containerScrollTop: 0,
      containerTop: 200,
      containerHeight: 600,
      targetTop: 220,
      targetHeight: 20,
    })).toBe(0)
  })
})
