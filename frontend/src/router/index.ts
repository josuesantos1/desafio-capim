import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import SearchView from '../views/SearchView.vue'
import ClinicPublicView from '../views/ClinicPublicView.vue'
import DentistPublicView from '../views/DentistPublicView.vue'
import ClinicListView from '../views/ClinicListView.vue'
import ClinicDetailView from '../views/ClinicDetailView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/search', component: SearchView },
    { path: '/c/:id', component: ClinicPublicView, props: true },
    { path: '/d/:clinicId/:dentistId', component: DentistPublicView, props: true },
    { path: '/clinics', component: ClinicListView },
    { path: '/clinics/:id', component: ClinicDetailView, props: true },
  ],
})
