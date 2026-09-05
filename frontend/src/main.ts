import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import { router } from './router'
import { vReveal } from './directives/reveal'

createApp(App).use(createPinia()).use(router).directive('reveal', vReveal).mount('#app')
