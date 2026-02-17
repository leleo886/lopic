import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '../stores/auth';

// 路由配置
const routes: Array<RouteRecordRaw> = [
   {
      path: '/',
      name: '主页',
      component: () => import('../views/gallery/app.vue'),
      meta: {
        title: '主页',
      },
    },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: {
      requiresAuth: false,
      title: '登录',
    },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/RegisterView.vue'),
    meta: {
      requiresAuth: false,
      title: '注册',
    },
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('../views/RequestPasswordResetView.vue'),
    meta: {
      requiresAuth: false,
      title: '重置密码',
    },
  },
  {
    path: '/gallery/:username?',
    name: 'Gallery',
    component: () => import('../views/gallery/app.vue'),
    meta: {
      title: '画廊',
    },
  },
  {
    path: '/manage',
    component: () => import('../views/Layout.vue'),
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/DashboardView.vue'),
        meta: {
          title: '仪表盘',
        },
      },
      {
        path: 'albums',
        name: 'Albums',
        component: () => import('../views/AlbumsView.vue'),
        meta: {
          title: '我的相册',
        },
      },
      {
        path: 'album/:id',
        name: 'AlbumDetail',
        component: () => import('../views/AlbumDetailView.vue'),
        meta: {
          title: '相册详情',
        },
      },
      {
        path: 'images',
        name: 'Images',
        component: () => import('../views/ImagesView.vue'),
        meta: {
          title: '我的图片',
        },
      },
      {
        path: 'upload',
        name: 'Upload',
        component: () => import('../views/UploadView.vue'),
        meta: {
          title: '上传图片',
        },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/ProfileView.vue'),
        meta: {
          title: '个人设置',
        },
      },
    ],
  },
  // 管理员路由
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('../views/admin/AdminLayout.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: '管理员中心',
    },
    children: [
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('../views/admin/UsersView.vue'),
        meta: {
          title: '用户管理',
        },
      },
      {
        path: 'roles',
        name: 'AdminRoles',
        component: () => import('../views/admin/RolesView.vue'),
        meta: {
          title: '角色管理',
        },
      },
      {
        path: 'albums',
        name: 'AdminAlbums',
        component: () => import('../views/admin/AlbumsView.vue'),
        meta: {
          title: '相册管理',
        },
      },
      {
        path: 'images',
        name: 'AdminImages',
        component: () => import('../views/admin/ImagesView.vue'),
        meta: {
          title: '图片管理',
        },
      },
      {
        path: 'backup',
        name: 'AdminBackup',
        component: () => import('../views/admin/BackupView.vue'),
        meta: {
          title: '备份与恢复',
        },
      },
      {
        path: 'storage',
        name: 'AdminStorage',
        component: () => import('../views/admin/StorageView.vue'),
        meta: {
          title: '存储管理',
        },
      },
      {
        path: 'system',
        name: 'AdminSystem',
        component: () => import('../views/admin/SystemView.vue'),
        meta: {
          title: '系统设置',
        },
      },
    ],
  },
  // 404页面
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../views/NotFoundView.vue'),
    meta: {
      title: '页面不存在',
    },
  },
];

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore();
  
  // 恢复用户状态
  if (!authStore.user && authStore.isAuthenticated) {
    authStore.restoreState();
  }
  
  // 检查是否需要认证
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' });
  } 
  // 检查是否需要管理员权限
  else if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next({ name: 'Dashboard' });
  } 
  // 已登录用户不能访问登录页
  else if (to.name === 'Login' && authStore.isAuthenticated) {
    next({ name: 'Dashboard' });
  } 
  else {
    next();
  }
});

export default router;
