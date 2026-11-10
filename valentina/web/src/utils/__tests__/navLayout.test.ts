import { describe, it, expect } from 'vitest'
import {
  sectionFromPath,
  showChatSidebar,
  showContextPanel,
  TOP_LEVEL_SECTIONS,
  NAV_GROUPS,
  NAV_ITEMS,
} from '../navLayout'

describe('sectionFromPath', () => {
  it('maps chat, space and the root landing onto the chat section', () => {
    expect(sectionFromPath('/')).toBe('chat')
    expect(sectionFromPath('/chat')).toBe('chat')
    expect(sectionFromPath('/chat/sess-1')).toBe('chat')
    expect(sectionFromPath('/space/sp-1')).toBe('chat')
    expect(sectionFromPath('/space/sp-1/extra')).toBe('chat')
  })

  it('folds workflow aliases onto automation', () => {
    expect(sectionFromPath('/workflows')).toBe('automation')
    expect(sectionFromPath('/workflows/wf-1')).toBe('automation')
    expect(sectionFromPath('/workflows/wf-1/runs/run-9')).toBe('automation')
    expect(sectionFromPath('/routines')).toBe('automation')
  })

  it('maps every CRM route to its own active navigation section', () => {
    expect(sectionFromPath('/valentina/opportunities')).toBe('crm-opportunities')
    expect(sectionFromPath('/valentina/opportunities?status=qualified')).toBe('crm-opportunities')
    expect(sectionFromPath('/valentina/leads')).toBe('crm-leads')
    expect(sectionFromPath('/valentina/companies')).toBe('crm-companies')
    expect(sectionFromPath('/valentina/sources')).toBe('crm-sources')
    expect(sectionFromPath('/valentina/research')).toBe('crm-research')
    expect(sectionFromPath('/valentina')).toBe('crm-opportunities')
  })

  it('keeps other upstream first segments as their own section', () => {
    expect(sectionFromPath('/overview')).toBe('overview')
    expect(sectionFromPath('/stats')).toBe('stats')
    expect(sectionFromPath('/logs')).toBe('logs')
    expect(sectionFromPath('/settings')).toBe('settings')
    expect(sectionFromPath('/agents')).toBe('agents')
    expect(sectionFromPath('/agents/atlas')).toBe('agents')
    expect(sectionFromPath('/connections')).toBe('connections')
    expect(sectionFromPath('/skills/browse')).toBe('skills')
    expect(sectionFromPath('/inbox')).toBe('inbox')
    expect(sectionFromPath('/models')).toBe('models')
    expect(sectionFromPath('/memory')).toBe('memory')
    expect(sectionFromPath('/cloud')).toBe('cloud')
  })
})

describe('showChatSidebar / showContextPanel', () => {
  const chatRoutes = ['/', '/chat', '/chat/sess-1', '/space/abc', '/space/abc/x']
  const topLevelRoutes = [
    '/overview',
    '/stats',
    '/logs',
    '/settings',
    '/agents',
    '/agents/atlas',
    '/connections',
    '/skills',
    '/skills/browse',
    '/workflows',
    '/workflows/wf-1',
    '/routines',
    '/inbox',
    '/models',
    '/memory',
    '/cloud',
    '/valentina/opportunities',
    '/valentina/leads',
    '/valentina/companies',
    '/valentina/sources',
    '/valentina/research',
  ]

  it('shows the company/channel/DM rail on chat and space routes only', () => {
    for (const path of chatRoutes) {
      expect(showChatSidebar(path), path).toBe(true)
      expect(showContextPanel(path), path).toBe(true)
    }
  })

  it('hides the rail on every top-level app and CRM surface', () => {
    for (const path of topLevelRoutes) {
      expect(showChatSidebar(path), path).toBe(false)
      expect(showContextPanel(path), path).toBe(false)
    }
  })

  it('covers all documented top-level section names', () => {
    expect(TOP_LEVEL_SECTIONS).toEqual(expect.arrayContaining([
      'overview',
      'stats',
      'logs',
      'settings',
      'agents',
      'connections',
      'skills',
      'automation',
      'inbox',
      'models',
      'memory',
      'cloud',
      'crm-opportunities',
      'crm-leads',
      'crm-companies',
      'crm-sources',
      'crm-research',
    ]))
    for (const section of TOP_LEVEL_SECTIONS) {
      expect(showChatSidebar('/' + section)).toBe(false)
    }
  })
})

describe('NAV_GROUPS', () => {
  it('preserves upstream group order and appends CRM as its own group', () => {
    expect(NAV_GROUPS.map((group) => group.id)).toEqual(['chat', 'admin', 'logs', 'crm'])
  })

  it('keeps Overview and Chat in the chat group', () => {
    expect(NAV_GROUPS.find((group) => group.id === 'chat')!.items.map((item) => item.section))
      .toEqual(['overview', 'chat'])
  })

  it('preserves the upstream admin and logs groups', () => {
    const admin = NAV_GROUPS.find((group) => group.id === 'admin')!.items.map((item) => item.section)
    const logs = NAV_GROUPS.find((group) => group.id === 'logs')!.items.map((item) => item.section)

    expect(admin).toEqual(['agents', 'memory', 'models', 'automation', 'connections', 'skills', 'settings'])
    expect(logs).toEqual(['stats', 'logs', 'inbox'])
    expect(admin).not.toEqual(expect.arrayContaining(['stats', 'logs', 'inbox']))
    expect(logs).not.toContain('settings')
  })

  it('gives each CRM destination a unique section and matching route', () => {
    const crmItems = NAV_GROUPS.find((group) => group.id === 'crm')!.items
    expect(crmItems.map((item) => item.section)).toEqual([
      'crm-opportunities',
      'crm-leads',
      'crm-companies',
      'crm-sources',
      'crm-research',
    ])
    expect(crmItems.map((item) => item.path)).toEqual([
      '/valentina/opportunities',
      '/valentina/leads',
      '/valentina/companies',
      '/valentina/sources',
      '/valentina/research',
    ])
    expect(new Set(crmItems.map((item) => item.section)).size).toBe(5)
  })

  it('keeps NAV_ITEMS flattened in group order', () => {
    expect(NAV_ITEMS.map((item) => item.section)).toEqual([
      'overview',
      'chat',
      'agents',
      'memory',
      'models',
      'automation',
      'connections',
      'skills',
      'settings',
      'stats',
      'logs',
      'inbox',
      'crm-opportunities',
      'crm-leads',
      'crm-companies',
      'crm-sources',
      'crm-research',
    ])
  })
})
