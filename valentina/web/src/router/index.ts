import { createRouter, createWebHashHistory } from 'vue-router'
import OpportunitiesView from '../views/OpportunitiesView.vue'
import LeadsView from '../views/LeadsView.vue'
import CompaniesView from '../views/CompaniesView.vue'
import SourcesView from '../views/SourcesView.vue'
import ResearchView from '../views/ResearchView.vue'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/valentina/opportunities' },
    { path: '/valentina/opportunities', component: OpportunitiesView },
    { path: '/valentina/leads', component: LeadsView },
    { path: '/valentina/companies', component: CompaniesView },
    { path: '/valentina/sources', component: SourcesView },
    { path: '/valentina/research', component: ResearchView },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})