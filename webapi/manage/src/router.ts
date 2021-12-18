import { createRouter, createWebHashHistory } from 'vue-router'
import Home from './pages/home/Home.vue'
import Router from './pages/Router/Router.vue'
import Server from './pages/Server/Server.vue'
import Transfer from './pages/Transfer/Transfer.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/transfer',
      name: 'transfer',
      component: Transfer
    },
    {
      path: '/router',
      name: 'router',
      component: Router
    },
    {
      path: '/server',
      name: 'server',
      component: Server
    },
  ]
})

export default router