import { createRouter, createWebHistory } from 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    standalone?: boolean
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboards',
      component: () => import('@/pages/DashboardListPage.vue'),
    },
    {
      path: '/templates',
      name: 'templates',
      component: () => import('@/pages/TemplatesPage.vue'),
    },
    {
      path: '/templates/new',
      name: 'template-new',
      component: () => import('@/pages/TemplateEditorPage.vue'),
    },
    {
      path: '/templates/:id',
      name: 'template-edit',
      component: () => import('@/pages/TemplateEditorPage.vue'),
      props: true,
    },
    {
      path: '/create',
      name: 'create-dashboard',
      component: () => import('@/pages/CreateDashboardPage.vue'),
    },
    {
      path: '/dashboards/:id',
      name: 'dashboard-view',
      component: () => import('@/pages/DashboardViewPage.vue'),
      props: true,
      meta: { standalone: true },
    },
  ],
})

export default router
