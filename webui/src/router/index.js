// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'

const Login = () => import('@/views/Login.vue')
const MapView = () => import('@/views/MapView.vue')

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{ path: '/', redirect: '/map' },
		{ path: '/login', name: 'login', component: Login, meta: { guestOnly: true } },
		{ path: '/map', name: 'map', component: MapView, meta: { requiresAuth: true } },
		{ path: '/:pathMatch(.*)*', redirect: '/map' }
	]
})

router.beforeEach((to, _from, next) => {
	const authed = !!localStorage.getItem('bubble_user')
	if (to.meta.requiresAuth && !authed) return next({ name: 'login' })
	if (to.meta.guestOnly && authed) return next({ name: 'map' })
	next()
})

export default router


