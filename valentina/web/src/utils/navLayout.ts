export const NAV_GROUPS: readonly RailNavGroup[] = [
  {
    id: 'chat',
    items: [
      { section: 'chat',        label: 'Chat',         path: '/chat',           icon: 'chat' },
    ],
  },
  {
    id: 'admin',
    items: [
      { section: 'agents',      label: 'Agents',       path: '/agents',         icon: 'agents' },
      { section: 'memory',      label: 'Memory',       path: '/memory',         icon: 'memory' },
      { section: 'models',      label: 'Models',       path: '/models',         icon: 'models' },
      { section: 'automation',  label: 'Automation',   path: '/workflows',      icon: 'automation' },
      { section: 'connections', label: 'Connections',  path: '/connections',    icon: 'connections' },
      { section: 'skills',      label: 'Skills',       path: '/skills/browse',  icon: 'skills' },
      { section: 'settings',    label: 'Settings',     path: '/settings',       icon: 'settings' },
      { section: 'crm',         label: 'Возможности', path: '/valentina/opportunities', icon: 'crm' },
      { section: 'crm',         label: 'Заявки',      path: '/valentina/leads',      icon: 'crm' },
      { section: 'crm',         label: 'Компании',    path: '/valentina/companies',  icon: 'crm' },
      { section: 'crm',         label: 'Источники',   path: '/valentina/sources',    icon: 'crm' },
      { section: 'crm',         label: 'Исследования',path: '/valentina/research',   icon: 'crm' },
    ],
  },
  {
    id: 'logs',
    items: [
      { section: 'stats',       label: 'Stats',        path: '/stats',          icon: 'stats' },
      { section: 'logs',        label: 'Logs',         path: '/logs',           icon: 'logs' },
      { section: 'inbox',       label: 'Activity Log', path: '/inbox',          icon: 'inbox' },
    ],
  },
]

export const NAV_ITEMS = NAV_GROUPS.flatMap((g) => [...g.items])