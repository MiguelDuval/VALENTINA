import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const viewsDir = join(dirname(fileURLToPath(import.meta.url)), '..')
const leadsView = readFileSync(join(viewsDir, 'LeadsView.vue'), 'utf-8')

describe('VALENTINA Vue source regressions', () => {
  it('LeadsView.vue does not contain an extra script closing tag after style', () => {
    const styleEndIndex = leadsView.indexOf('</style>')
    expect(styleEndIndex).toBeGreaterThan(-1)
    const afterStyle = leadsView.slice(styleEndIndex + '</style>'.length)
    expect(afterStyle.trim()).not.toMatch(/^<\/script>/)
  })

  it('LeadsView pagination serializes the numeric ref value', () => {
    expect(leadsView).toContain("'&offset=' + offset.value")
    expect(leadsView).not.toContain("'&offset=' + offset)
  })
})
