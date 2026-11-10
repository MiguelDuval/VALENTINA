/** Route → layout mapping for the Vue shell.
 *
 * Chat / space / DM stay a three-column layout:
 *   icon rail + company/channel/DM sidebar + main pane.
 *
 * Every other top-level surface (stats, logs, settings, people, plugins,
 * workflows, activity, analytics, …) is two-column: icon rail + full-width main.
 * The company/channel/DM rail is chat-only.
 */

export type AppSection =
  | 'overview'
  | 'chat'
  | 'agents'
  | 'memory'
  | 'models'
  | 'automation'
  | 'connections'
  | 'skills'
  | 'stats'
  | 'settings'
  | 'logs'
  | 'inbox'
  | 'cloud'
  | 'crm-opportunities'
  | 'crm-leads'
  | 'crm-companies'
  | 'crm-sources'
  | 'crm-research'
  | string

/** First path segment, with aliases folded onto the icon-rail section. */
export function sectionFromPath(path: string): AppSection {
  const raw = (path.split('?')[0] || '').trim()
  const seg = (raw.replace(/^#/, '').split('/').filter(Boolean)[0] || 'chat').toLowerCase()
  if (seg === 'routines' || seg === 'workflows') return 'automation'
  if (seg === 'space') return 'chat'
  if (seg === 'valentina') {
    const second = (raw.replace(/^#/, '').split('/').filter(Boolean)[1] || '').toLowerCase()
    switch (second) {
      case 'opportunities': return 'crm-opportunities'
      case 'leads': return 'crm-leads'
      case 'companies': return 'crm-companies'
      case 'sources': return 'crm-sources'
      case 'research': return 'crm-research'
    }
    return 'crm-opportunities'
  }
  return seg
}

/** Company / channel / DM sidebar — chat, space, and the / chat landing only. */
export function showChatSidebar(path: string): boolean {
  return sectionFromPath(path) === 'chat'
}

/** Context column (Column 2). Same as the chat sidebar: chat-only. */
export function showContextPanel(path: string): boolean {
  return showChatSidebar(path)
}

/** Top-level app surfaces that must not show the company/channel/DM rail. */
export const TOP_LEVEL_SECTIONS: readonly string[] = [
  'overview',
  'stats',        // analytics
  'logs',
  'settings',
  'agents',       // people
  'connections',  // plugins
  'skills',       // plugins / skill library
  'automation',   // workflows
  'inbox',        // activity
  'models',
  'memory',
  'cloud',
  'crm-opportunities',
  'crm-leads',
  'crm-companies',
  'crm-sources',
  'crm-research',
]

export type NavGroupId = 'chat' | 'admin' | 'logs' | 'crm'

export interface RailNavItem {
  section: string
  label: string
  path: string
  icon: string
}

export interface RailNavGroup {
  id: NavGroupId
  items: readonly RailNavItem[]
}

/** Slack-class icon-rail groups: chat, then admin, then logs, then crm. Settings is admin. */
export const NAV_GROUPS: readonly RailNavGroup[] = [
  {
    id: 'chat',
    items: [
      { section: 'overview',    label: 'Обзор',        path: '/overview',       icon: 'overview' },
      { section: 'chat',        label: 'Ассистент',    path: '/chat',           icon: 'chat' },
    ],
  },
  {
    id: 'admin',
    items: [
      { section: 'agents',      label: 'Агенты',       path: '/agents',         icon: 'agents' },
      { section: 'memory',      label: 'Память',       path: '/memory',         icon: 'memory' },
      { section: 'models',      label: 'Модели',       path: '/models',         icon: 'models' },
      { section: 'automation',  label: 'Процессы',     path: '/workflows',      icon: 'automation' },
      { section: 'connections', label: 'Подключения',  path: '/connections',    icon: 'connections' },
      { section: 'skills',      label: 'Навыки',       path: '/skills/browse',  icon: 'skills' },
      { section: 'settings',    label: 'Настройки',    path: '/settings',       icon: 'settings' },
    ],
  },
  {
    id: 'logs',
    items: [
      { section: 'stats',       label: 'Статистика',   path: '/stats',          icon: 'stats' },
      { section: 'logs',        label: 'Журнал',       path: '/logs',           icon: 'logs' },
      { section: 'inbox',       label: 'Уведомления',  path: '/inbox',          icon: 'inbox' },
    ],
  },
  {
    id: 'crm',
    items: [
      { section: 'crm-opportunities', label: 'Возможности', path: '/valentina/opportunities', icon: 'crm' },
      { section: 'crm-leads',         label: 'Заявки',      path: '/valentina/leads',         icon: 'crm' },
      { section: 'crm-companies',     label: 'Компании',    path: '/valentina/companies',   icon: 'crm' },
      { section: 'crm-sources',       label: 'Источники',   path: '/valentina/sources',     icon: 'crm' },
      { section: 'crm-research',      label: 'Исследования',path: '/valentina/research',    icon: 'crm' },
    ],
  },
]

export const NAV_ITEMS = NAV_GROUPS.flatMap((g) => [...g.items])
