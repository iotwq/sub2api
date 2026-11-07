import type { ModelPlazaGroup } from '@/api/modelPlaza'

/** A group with one shared media rate uses it in the heading/filter as well as its rows. */
export function plazaGroupDisplayRate(group: ModelPlazaGroup, rechargeRatio: number, userRate = true): number {
  const base = (userRate ? group.user_rate_multiplier ?? group.rate_multiplier : group.rate_multiplier) / rechargeRatio
  const rates = group.models.map(model => {
    const mode = model.pricing?.billing_mode || 'token'
    const media = model.media_type || mode
    if (mode !== 'token' && media === 'image' && group.image_rate_independent) {
      return group.image_rate_multiplier ?? 1
    }
    if (mode !== 'token' && media === 'video' && group.video_rate_independent) {
      return Math.max(0, group.video_rate_multiplier ?? 1)
    }
    return base
  })
  const rate = rates.length && rates.every(value => value === rates[0]) ? rates[0] : base
  return Number(rate.toPrecision(10))
}
