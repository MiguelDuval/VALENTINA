import { createRouter, createWebHashHistory } from 'vue-router'
import OverviewView from '../views/OverviewView.vue'
import ChatView from '../views/ChatView.vue'
import AgentsView from '../views/AgentsView.vue'
import ModelsView from '../views/ModelsView.vue'
import ConnectionsView from '../views/ConnectionsView.vue'
import SettingsView from '../views/SettingsView.vue'
import LogsView from '../views/LogsView.vue'
import StatsView from '../views/StatsView.vue'
import InboxView from '../views/InboxView.vue'
import WorkflowsView from '../views/WorkflowsView.vue'
import CloudView from '../views/CloudView.vue'
import SkillsView from '../views/SkillsView.vue'
import MemoryView from '../views/MemoryView.vue'
import OpportunitiesView from '../views/OpportunitiesView.vue'
import LeadsView from '../views/LeadsView.vue'
import CompaniesView from '../views/CompaniesView.vue'
import SourcesView from '../views/SourcesView.vue'
import ResearchView from '../views/ResearchView.vue'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/overview' },
    { path: '/overview', component: OverviewView },
    { path: '/chat/:sessionId?', component: ChatView, props: true },
    { path: '/space/:spaceId', component: ChatView, props: true },
    { path: '/agents/:agentName?', component: AgentsView, props: true },
    { path: '/models/:provider?', component: ModelsView, props: true },
    { path: '/connections', component: ConnectionsView },
    { path: '/settings', component: SettingsView },
    { path: '/logs', component: LogsView },
    { path: '/stats', component: StatsView },
    { path: '/inbox', component: InboxView },
    { path: '/memory', component: MemoryView },
    { path: '/routines', redirect: '/workflows' },
    { path: '/workflows/:id?', component: WorkflowsView, props: true },
    { path: '/workflows/:id/runs/:runId', component: WorkflowsView, props: true },
    { path: '/cloud', component: CloudView },
    { path: '/skills/:tab?', component: SkillsView, props: true },
    { path: '/valentina/opportunities', component: OpportunitiesView },
    { path: '/valentina/leads', component: LeadsView },
    { path: '/valentina/companies', component: CompaniesView },
    { path: '/valentina/sources', component: SourcesView },
    { path: '/valentina/research', component: ResearchView },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})
