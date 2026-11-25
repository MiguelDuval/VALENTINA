// Decimal input is parsed as text; no binary floating-point multiplication.
export function decimalToInteger(text: string, places: number): number {
  const value = text.trim().replace(',', '.')
  if (!new RegExp(`^\\d+(?:\\.\\d{1,${places}})?$`).test(value)) throw new Error('Введите неотрицательное число с допустимой точностью')
  const [whole, fraction = ''] = value.split('.')
  const integer = BigInt(whole!) * (10n ** BigInt(places)) + BigInt(fraction.padEnd(places, '0'))
  if (integer > BigInt(Number.MAX_SAFE_INTEGER)) throw new Error('Слишком большое значение')
  return Number(integer)
}
export function integerToDecimal(value: number, places: number): string {
  const integer = BigInt(value), scale = 10n ** BigInt(places)
  return `${integer / scale}.${(integer % scale).toString().padStart(places, '0')}`
}
export function euros(cents: number): string { return integerToDecimal(cents, 2).replace('.', ',') + ' €' }
export const projectStatuses: Record<string, string> = { planned: 'Запланирован', active: 'В работе', on_hold: 'Приостановлен', completed: 'Завершён', cancelled: 'Отменён' }
export const projectTransitions: Record<string, string[]> = { planned: ['active', 'cancelled'], active: ['on_hold', 'completed', 'cancelled'], on_hold: ['active', 'cancelled'] }
export interface ProjectInput { title: string; description: string; service_scope: string; company_id: string; source_lead_id: string; location_text: string; start_date: string; end_date: string }
export interface Project extends ProjectInput { project_id: string; status: string; version: number; created_at: string; updated_at: string }
export interface Line { kind: string; name: string; quantity_milli: number; unit_cents: number; vat_bps: number }
export interface Totals { net_cents: number; vat_cents: number; total_cents: number }
export interface Estimate extends Totals { estimate_id: string; revision: number; lines: Line[]; status: string; created_at: string; approved_at?: string }
export interface DocumentRecord { document_id: string; name: string; type: string; annotation: string; estimate_id: string; created_at: string }
