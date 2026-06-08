import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('../views/Login.vue') },
    {
      path: '/',
      component: () => import('../views/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/dashboard' },
        { path: 'dashboard', component: () => import('../views/Dashboard.vue') },
        { path: 'tenants', component: () => import('../views/Tenants.vue') },
        { path: 'categories', component: () => import('../views/Categories.vue') },
        { path: 'banners', component: () => import('../views/Banners.vue') },
        { path: 'site-settings', component: () => import('../views/SiteSettings.vue') },
        { path: 'home-featured', component: () => import('../views/HomeFeatured.vue') },
        { path: 'menus', component: () => import('../views/Menus.vue') },
        { path: 'roles', component: () => import('../views/Roles.vue') },
        { path: 'admins', component: () => import('../views/Admins.vue') },
        { path: 'products', component: () => import('../views/Products.vue') },
        { path: 'products/new', component: () => import('../views/ProductForm.vue') },
        { path: 'products/:id/edit', component: () => import('../views/ProductForm.vue') },
        { path: 'coupons', component: () => import('../views/Coupons.vue') },
        { path: 'users', component: () => import('../views/Users.vue') },
        { path: 'orders', component: () => import('../views/Orders.vue') },
        { path: 'forbidden', component: () => import('../views/Forbidden.vue') },
      ],
    },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const token = localStorage.getItem('admin_token')
  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }
  if (to.path === '/login' && token) {
    next('/')
    return
  }
  if (to.meta.requiresAuth && token) {
    const auth = useAuthStore()
    if (!auth.admin) {
      try {
        await auth.restore()
      } catch {
        next('/login')
        return
      }
    }
    if (to.path !== '/forbidden' && to.path !== '/dashboard' && auth.menuPaths.length > 0) {
      const base = to.path.split('/').slice(0, 2).join('/') || to.path
      const allowed = auth.menuPaths.some((p) => to.path === p || to.path.startsWith(p + '/'))
      if (!allowed && !auth.hasPath(base)) {
        next('/forbidden')
        return
      }
    }
  }
  next()
})

export default router
