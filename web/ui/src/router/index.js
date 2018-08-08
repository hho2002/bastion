import Vue from 'vue'
import Router from 'vue-router'
import store from '@/store'
import Index from '@/components/Index'
import Login from '@/components/Login'
import Dashboard from '@/components/Dashboard'
import Servers from '@/components/Servers'
import Users from '@/components/Users'
import Sessions from '@/components/Sessions'
import Profile from '@/components/Profile'

Vue.use(Router)

let router = new Router({
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index
    },
    {
      path: '/login',
      name: 'Login',
      component: Login
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: Dashboard,
      meta: {
        requiresLoggedIn: true
      }
    },
    {
      path: '/servers',
      name: 'Servers',
      component: Servers,
      meta: {
        requiresLoggedInAsAdmin: true
      }
    },
    {
      path: '/users',
      name: 'Users',
      component: Users,
      meta: {
        requiresLoggedInAsAdmin: true
      }
    },
    {
      path: '/sessions',
      name: 'Sessions',
      component: Sessions,
      meta: {
        requiresLoggedInAsAdmin: true
      }
    },
    {
      path: '/profile',
      name: 'Profile',
      component: Profile,
      meta: {
        requiresLoggedInAsAdmin: true
      }
    }
  ]
})

router.beforeEach((to, from, next) => {
  if ((to.meta.requiresLoggedIn || to.meta.requiresLoggedInAsAdmin) && !store.getters.isLoggedIn) {
    next('/login')
  } else if (to.meta.requiresLoggedInAsAdmin && !store.getters.isLoggedInAsAdmin) {
    next('/dashboard')
  } else {
    next(true)
  }
})

export default router
