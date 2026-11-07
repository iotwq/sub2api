import { describe, expect, it } from 'vitest'
import { plazaGroupDisplayRate } from '../modelPlaza'
import type { ModelPlazaGroup, PlazaModel } from '@/api/modelPlaza'

const image = { media_type: 'image', pricing: { billing_mode: 'per_request' } } as PlazaModel
const token = { pricing: { billing_mode: 'token' } } as PlazaModel
const group = { rate_multiplier: 1, image_rate_independent: true, image_rate_multiplier: 1, models: [image] } as ModelPlazaGroup

describe('plazaGroupDisplayRate', () => {
  it('uses the explicit independent rate for a media-only group', () => {
    expect(plazaGroupDisplayRate(group, 6)).toBe(1)
    expect(plazaGroupDisplayRate({ ...group, rate_multiplier: 0.36, user_rate_multiplier: 0.12 }, 6)).toBe(1)
  })
  it('does not infer an independent rate from a group multiplier of one', () => {
    expect(plazaGroupDisplayRate({ ...group, image_rate_independent: false }, 6)).toBe(0.1666666667)
  })
  it('keeps the group rate for mixed models and applies user rates', () => {
    expect(plazaGroupDisplayRate({ ...group, rate_multiplier: 0.36, models: [image, token] }, 6)).toBe(0.06)
    expect(plazaGroupDisplayRate({ ...group, rate_multiplier: 0.36, user_rate_multiplier: 0.12, models: [token] }, 6)).toBe(0.02)
  })
})
