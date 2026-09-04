import { createRouter, createWebHistory } from 'vue-router'
import ClinicListView from '../views/ClinicListView.vue'
import ClinicDetailView from '../views/ClinicDetailView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/clinics' },
    { path: '/clinics', component: ClinicListView },
    { path: '/clinics/:id', component: ClinicDetailView, props: true },
  ],
})
