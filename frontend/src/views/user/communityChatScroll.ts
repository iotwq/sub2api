export const communityChatBottomThreshold = 96

export function isCommunityChatNearBottom(
  metrics: Pick<HTMLElement, 'scrollTop' | 'scrollHeight' | 'clientHeight'>,
  threshold = communityChatBottomThreshold,
): boolean {
  return metrics.scrollHeight - metrics.scrollTop - metrics.clientHeight <= threshold
}

export function shouldFollowCommunityChatMessage(isPinnedToBottom: boolean, isOwnMessage: boolean): boolean {
  return isPinnedToBottom || isOwnMessage
}
