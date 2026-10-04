import { describe, expect, it } from 'vitest'

import {
  communityChatBottomThreshold,
  isCommunityChatNearBottom,
  shouldFollowCommunityChatMessage,
} from '../communityChatScroll'

describe('community chat scroll behavior', () => {
  it('treats the configured bottom threshold as pinned', () => {
    expect(isCommunityChatNearBottom({
      scrollTop: 404,
      scrollHeight: 1000,
      clientHeight: 500,
    })).toBe(true)
    expect(communityChatBottomThreshold).toBe(96)
  })

  it('does not treat older message reading positions as pinned', () => {
    expect(isCommunityChatNearBottom({
      scrollTop: 200,
      scrollHeight: 1000,
      clientHeight: 500,
    })).toBe(false)
  })

  it('follows incoming messages only while pinned', () => {
    expect(shouldFollowCommunityChatMessage(true, false)).toBe(true)
    expect(shouldFollowCommunityChatMessage(false, false)).toBe(false)
  })

  it('always follows the current user message', () => {
    expect(shouldFollowCommunityChatMessage(false, true)).toBe(true)
  })
})
