import { useI18n } from 'vue-i18n'
import type { MonitorIntelligenceResult } from '@/api/admin/channelMonitor'

export function useMonitorIntelligence() {
  const { t } = useI18n()
  function label(result?: MonitorIntelligenceResult | null): string {
    const status = result?.status || 'untested'
    return t(`monitorCommon.intelligence.${status}`)
  }
  function color(result?: MonitorIntelligenceResult | null): string {
    if (result?.status === 'passed') return 'bg-emerald-500'
    if (result?.status === 'failed') return 'bg-red-500'
    return 'bg-gray-300 dark:bg-dark-600'
  }
  function tooltip(result?: MonitorIntelligenceResult | null, checkedAt?: string): string {
    const parts = [label(result)]
    if (result?.status === 'passed' || result?.status === 'failed') {
      parts.push(t('monitorCommon.intelligence.candyAnswer', {
        answer: result.answer || t('monitorCommon.intelligence.answerNotRecorded'),
      }))
    } else if (result?.reason) {
      const reasons: Record<string, string> = {
        response_incomplete: 'responseIncomplete',
        empty_response: 'emptyResponse', missing_final_answer: 'missingFinalAnswer',
        request_failed: 'requestFailed', incompatible_configuration: 'incompatible',
      }
      const key = reasons[result.reason]
      if (key) parts.push(t(`monitorCommon.intelligence.${key}`))
    }
    if (checkedAt) parts.push(new Date(checkedAt).toLocaleString())
    return parts.join(' · ')
  }
  return { label, color, tooltip }
}
