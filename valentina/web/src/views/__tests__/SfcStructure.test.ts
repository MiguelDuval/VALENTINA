import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { join } from 'path'

describe('Single-file component structure', () => {
  const viewsDir = join(__dirname, '..')

  it('LeadsView.vue must not contain an extra </script> after </style>', () => {
    const leadsPath = join(viewsDir, 'LeadsView.vue')
    const content = readFileSync(leadsPath, 'utf-8')
    const styleEndIndex = content.indexOf('</style>')
    expect(styleEndIndex).toBeGreaterThan(-1)
    const afterStyle = content.substring(styleEndIndex + '</style>'.length)
    expect(afterStyle.trim()).not.toStartWith('</script>')
  })

  it('LeadsView.vue must not concatenate a ref object directly in a URL', () => {
    const leadsPath = join(viewsDir, 'LeadsView.vue')
    const content = readFileSync(leadsPath, 'utf-8')
    // offset is a ref; the API URL must use offset.value, not the ref object itself
    expect(content).not.toContain("' + offset')")
    expect(content).not.toContain("+ offset')")
  })

  it('every Vue SFC must have </style> as its last top-level closing tag', () => {
    const fs = require('fs')
    const files = fs.readdirSync(viewsDir).filter(f => f.endsWith('.vue'))
    for (const file of files) {
      const content = readFileSync(join(viewsDir, file), 'utf-8')
      const lastNonEmpty = content.trim().split('\n').pop()
      expect(lastNonEmpty).toBe('</style>')
    }
  })
})