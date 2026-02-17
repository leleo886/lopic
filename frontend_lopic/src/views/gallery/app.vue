<template>
  <div class="art-gallery-container" :style="backgroundStyle">
    <div class="top-fixed-elements">
      <div class="top-left">
        <div class="logo-area">
          <span class="logo-text">{{ galleryConfig.Title || projectName }}</span>
        </div>
      </div>
      <div class="top-center">
        <div class="control-console">
          <div class="console-content">
            <!-- 分类下拉菜单 -->
            <div class="category-dropdown">
              <button class="dropdown-toggle" @click="toggleDropdown">
                {{ currentAlbumTitle }}
              </button>
              <div class="dropdown-menu" v-if="isDropdownOpen">
                <button
                  v-for="album in albums"
                  :key="album.id"
                  :class="['dropdown-item', { active: currentAlbumId === album.id }]"
                  @click="handleAlbumChange(album.id); isDropdownOpen = false"
                >
                  {{ album.name }}
                </button>
              </div>
            </div>
            
            <!-- 搜索部分 -->
            <div class="search-section">
              <SearchOutlined class="search-icon" />
              <input 
                class="console-search" 
                placeholder="搜索标题或标签" 
                v-model="searchQuery"
                @focus="activateSearch"
                @blur="deactivateSearch"
                @keyup.enter="handleSearch"
              />
              <button v-if="searchQuery" class="clear-search" @click="clearSearch">×</button>
            </div>
          </div>
        </div>
      </div>
      <div class="top-right">
        <button @click="handleUploadClick" class="user-profile-btn">
          <UploadOutlined class="" />
        </button>
      </div>
    </div>

    <!-- 左侧分页栏 -->
    <aside class="pagination-sidebar">
      <div class="pagination-content">
        <div class="pagination-list">
          <!-- 上一页按钮 -->
          <button
            class="pagination-item prev-btn"
            @click="handlePageChange(page - 1)"
            :disabled="page <= 1"
          >
            ←
          </button>
          
          <!-- 页码按钮 -->
          <button
            v-for="(btn, index) in paginationButtons"
            :key="index"
            :class="['pagination-item', { active: page === btn, disabled: typeof btn === 'string' }]"
            :disabled="typeof btn === 'string'"
            @click="typeof btn === 'number' && handlePageChange(btn)"
          >
            {{ btn }}
          </button>
          
          <!-- 页码输入 -->
          <div class="page-input-container">
            <input
              v-model="inputPage"
              class="page-input"
              placeholder="页码"
              min="1"
              :max="totalPages"
              @keyup.enter="handlePageInput"
            />
          </div>
          
          <!-- 下一页按钮 -->
          <button
            class="pagination-item next-btn"
            @click="handlePageChange(page + 1)"
            :disabled="page >= totalPages"
          >
            →
          </button>
        </div>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="gallery-content">
      <!-- 加载中提示 -->
      <div v-if="loading && list.length === 0" class="loading-container">
        <div class="loading-content">
          <div class="loading-spinner-large"></div>
          <p class="loading-text">Loading...</p>
        </div>
      </div>
      
      <!-- 无数据提示 -->
      <div v-else-if="list.length === 0 && !loading" class="no-data-container">
        <div class="no-data-content">
          <div class="no-data-icon">🖼️</div>
          <h3>NO DATA</h3>
        </div>
      </div>
      
      <!-- 瀑布流 -->
      <Waterfall
        v-else
        class="artwork-waterfall"
        :list="list"
        :breakpoints="breakpoints"
        :gutter="gutter"
        imgSelector="file_url"
      >
        <template #default="{ item }">
          <div
            class="artwork-card"
            :data-id="item.id"
            @mousemove="onCardMove($event, item)"
            @mouseleave="onCardLeave"
            @click="openDetailModal(item)"
          >
            <div class="card-content">
              <div class="artwork-image-container">
                <img
                  :src="getFileUrl(item.thumbnail_url)"
                  :style="{ aspectRatio: item.width/item.height > 1/0.7 ?  1/0.7 :(item.width + '/' + item.height) }"
                  @load="onImgLoaded"
                  :alt="item.original_name"
                  class="artwork-image"
                />
                
                <!-- 悬停光效 -->
                <div class="hover-glow" :style="{ '--cursor-x': '50%', '--cursor-y': '50%' } as any"></div>
                
                <!-- 悬停遮罩 -->
                <div class="artwork-overlay">
                  <div class="overlay-content">
                    <div class="artwork-meta">
                      <h3 class="artwork-title">{{ item.original_name.split('.')[0] }}</h3>
                      <p class="artwork-dimensions">{{ item.width }} × {{ item.height }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </Waterfall>

      <!-- 加载更多提示 -->
      <div v-if="list.length > 0" ref="loadMoreRef" class="infinite-loader">
        <div v-if="loading" class="loading-animation">
          <div class="loading-dot"></div>
          <div class="loading-dot"></div>
          <div class="loading-dot"></div>
        </div>
        <p v-else-if="finished" class="end-of-gallery">End of Collection</p>
      </div>

      <!-- 页脚 -->
      <footer class="gallery-footer">
        <div class="footer-content">
          <div class="footer-divider"></div>
          <div class="footer-text">
            <span class="footer-brand">{{ galleryConfig.Title || projectName }}</span>
            <span class="footer-separator">·</span>
            <span class="footer-copyright">© {{ new Date().getFullYear() }}</span>
          </div>
          <div class="footer-decoration">
            <span class="deco-line"></span>
            <span class="deco-dot"></span>
            <span class="deco-line"></span>
          </div>
        </div>
      </footer>
    </main>

    <!-- 全屏图片查看模态框 -->
    <div v-if="selectedArtwork" class="fullscreen-modal" @click="closeDetailModal">
      <div class="modal-content" @click.stop>
        <button class="modal-close" @click="closeDetailModal">×</button>
        <div 
          class="modal-image-container"
          @wheel="handleModalScroll"
          @mousedown="startDrag"
          @touchstart="handleTouchStart"
          @touchmove="handleTouchMove"
          @touchend="handleTouchEnd"
        >
          <!-- 加载图标 -->
          <div v-if="isImageLoading" class="image-loading">
            <div class="loading-spinner"></div>
          </div>
          
          <img 
            :src="getFileUrl(isOriginalImage ? selectedArtwork.file_url : selectedArtwork.thumbnail_url)" 
            :alt="selectedArtwork.original_name" 
            class="modal-image"
            :style="{
              transform: `scale(${modalScale}) translate(${modalTranslate.x}px, ${modalTranslate.y}px)`,
              transition: isDragging ? 'none' : 'transform 0.2s ease',
              opacity: isImageLoading ? 0.5 : 1
            }"
            @load="isImageLoading = false"
          />
        </div>
        <div class="modal-info">
          <h2>{{ selectedArtwork.original_name }}</h2>
          <p class="modal-dimensions">分辨率: {{ selectedArtwork.width }} × {{ selectedArtwork.height }}</p>
          <p class="modal-file-size">大小: {{ formatFileSize(selectedArtwork.file_size) }}</p>
          <p class="modal-mime-type">类型: {{ selectedArtwork.mime_type }}</p>
          
          <div class="modal-tags">
            <h3>标签</h3>
            <div class="tags-container">
              <span 
                v-for="(tag, index) in selectedArtwork.tags" 
                :key="index"
                class="tag"
              >
                {{ tag }}
              </span>
            </div>
          </div>
          
          <button v-if="!isOriginalImage" class="download-button" @click="viewOriginalImage(selectedArtwork)">
            查看原图
          </button>
          <button v-else class="download-button" @click="viewThumbnailImage(selectedArtwork)">
            查看缩略图
          </button>
        </div>
      </div>
    </div>
    
    <!-- 自定义内容 -->
    <div v-if="galleryConfig.CustomContent" v-html="galleryConfig.CustomContent" class="custom-content"></div>
  </div>
</template>


<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Waterfall } from 'vue-waterfall-plugin-next'
import 'vue-waterfall-plugin-next/dist/style.css'
import './app.css'
import './modal.css'
import gsap from 'gsap'
import { galleryApi } from '../../api/services'
import type { GalleryConfig } from '../../types/api'
import {
  UploadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue';
import { projectName } from '../../api/axios';
import { getFileUrl } from '../../utils/index';

const route = useRoute();
const username = computed(() => route.params.username as string || '$admin$');

const router = useRouter();

/* ---------- 状态管理 ---------- */
const albums = ref<any[]>([])
const list = ref<any[]>([])
const currentAlbumId = ref<number | null>(null)
const selectedArtwork = ref<any>(null)
const searchQuery = ref('')
const isSearching = ref(false)
const isSearchActive = ref(false)
const isDropdownOpen = ref(false) // 控制分类下拉菜单的显示/隐藏
const galleryConfig = ref<GalleryConfig>({ Title: '', BackgroundImage: '', CustomContent: '' }) // 画廊配置

/* ---------- 计算属性 ---------- */
const backgroundStyle = computed(() => {
  if (galleryConfig.value.BackgroundImage) {
    return {
      background: `
        linear-gradient(135deg, rgba(51, 51, 51, 0.9) 0%, rgba(82, 82, 82, 0.9) 50%, rgba(100, 100, 100, 0.9) 100%),
        url('${galleryConfig.value.BackgroundImage}')
      `,
      backgroundSize: 'cover',
      backgroundPosition: 'center',
      backgroundRepeat: 'no-repeat',
      backgroundAttachment: 'fixed'
    }
  }
  return {}
})

// 模态框状态管理
const modalScale = ref(1)
const modalTranslate = ref({ x: 0, y: 0 })
const isDragging = ref(false)
const startX = ref(0)
const startY = ref(0)
const startTranslateX = ref(0)
const startTranslateY = ref(0)
const isOriginalImage = ref(false) // 标记是否显示原图
const isImageLoading = ref(false) // 标记图片是否正在加载

const page = ref(1) // 当前显示的页码
const pageSize = 20
const loading = ref(false)
const finished = ref(false)
const loadMoreRef = ref<HTMLElement | null>(null)
const totalPages = ref(1) // 总页数
const observer = ref<IntersectionObserver | null>(null) // 无限滚动观察器
const loadedPages = ref<Set<number>>(new Set()) // 已加载的页码集合
const isInitialLoading = ref(true) // 标记是否处于初始加载阶段

/* ---------- 计算属性 ---------- */
const currentAlbumTitle = computed(() => {
  const album = albums.value.find(a => a.id === currentAlbumId.value)
  return album ? album.name : 'All Artworks'
})

// 计算分页按钮列表（支持中间省略）
const paginationButtons = computed(() => {
  const buttons = []
  const total = totalPages.value
  const current = page.value
  
  // 添加第一页
  if (total > 0) {
    buttons.push(1)
    
    // 如果总页数大于 9
    if (total > 9) {
      // 计算显示范围，确保当前页码前后3个按钮都能显示
      let start = Math.max(2, current - 3)
      let end = Math.min(total - 1, current + 3)
      
      // 确保至少显示7个页码按钮（当前页前后各3个）
      if (end - start < 6) {
        if (start === 2) {
          // 靠近开头，扩展结束位置
          end = Math.min(total - 1, start + 6)
        } else if (end === total - 1) {
          // 靠近结尾，扩展开始位置
          start = Math.max(2, end - 6)
        }
      }
      
      // 如果开始位置大于2，显示省略号
      if (start > 2) {
        buttons.push('...')
      }
      
      // 显示页码按钮
      for (let i = start; i <= end; i++) {
        buttons.push(i)
      }
      
      // 如果结束位置小于total-1，显示省略号
      if (end < total - 1) {
        buttons.push('...')
      }
    }
    // 如果总页数小于等于 9
    else {
      // 显示所有页码
      for (let i = 2; i <= total - 1; i++) {
        buttons.push(i)
      }
    }
    
    // 添加最后一页（如果不是第一页）
    if (total > 1) {
      buttons.push(total)
    }
  }
  
  return buttons
})

// 输入的页码
const inputPage = ref('')

/* ---------- 瀑布流配置 ---------- */
const breakpoints = {
  1800: { rowPerView: 5 },
  1500: { rowPerView: 4 },
  1200: { rowPerView: 3 },
  600: { rowPerView: 2 },
}

// 响应式gutter值
const gutter = computed(() => {
  const width = window.innerWidth
  if (width >= 1800) {
    return 30
  } else if (width >= 1500) {
    return 25
  } else if (width >= 1200) {
    return 20
  } else if (width >= 600) {
    return 15
  } else {
    return 10
  }
})

// 监听窗口大小变化，更新gutter
onMounted(() => {
  window.addEventListener('resize', () => {
    // 触发重新渲染
  })
})

/* ---------- API 请求 ---------- */
// 加载指定页的数据
const loadPage = async (albumId: number, targetPage: number, isAppend = false) => {
  if (loading.value) return
  if (loadedPages.value.has(targetPage)) return // 避免重复加载同一页
  
  loading.value = true

  try {
    const res = await galleryApi.getGalleryAlbumImages(username.value, albumId, {
      page: targetPage,
      page_size: pageSize
    })

    const images = res.data.images || []
    
    if (isAppend) {
      // 无限滚动追加数据
      list.value.push(...images)
    } else {
      // 分页切换替换数据
      list.value = images
    }
    
    // 标记该页已加载
    loadedPages.value.add(targetPage)
    
    // 更新总页数
    const totalCount = res.data.total || 0
    totalPages.value = Math.ceil(totalCount / pageSize)
    
    // 更新当前页码
    page.value = targetPage

    // 如果返回的数据量小于pageSize，说明已经是最后一页
    if (images.length < pageSize) {
      finished.value = true
    }
  } catch (error) {
    console.error('Failed to fetch images:', error)
  } finally {
    loading.value = false
  }
}

// 初始化加载或切换相册
const fetchImages = async (albumId: number, reset = false) => {
  if (reset) {
    isInitialLoading.value = true
    list.value = []
    page.value = 1
    loadedPages.value.clear()
    finished.value = false
    window.scrollTo({ top: 0, behavior: 'smooth' })
    await loadPage(albumId, 1, false)
    // 延迟关闭初始加载标志，避免观察器立即触发
    setTimeout(() => {
      isInitialLoading.value = false
    }, 500)
  }
}

// 无限滚动加载下一页
const loadNextPage = async (albumId: number) => {
  const nextPageNum = page.value + 1
  if (nextPageNum > totalPages.value) {
    finished.value = true
    return
  }
  await loadPage(albumId, nextPageNum, true)
}

const handleAlbumChange = async (id: number) => {
  currentAlbumId.value = id
  await fetchImages(id, true)
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

// 切换分类下拉菜单
const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value
}

/* ---------- 卡片交互动画 ---------- */
const onCardMove = (e: MouseEvent, _: any) => {
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top

  // 更新光标位置变量
  el.style.setProperty('--cursor-x', `${x}px`)
  el.style.setProperty('--cursor-y', `${y}px`)

  // 使用 GSAP 创建高级 3D 变换效果
  gsap.to(el, {
    scale: 1.03,
    x: (x / rect.width - 0.5) * 8,
    y: (y / rect.height - 0.5) * 8,
    rotationY: (x / rect.width - 0.5) * 5,
    rotationX: -(y / rect.height - 0.5) * 5,
    duration: 0.3,
    ease: 'power2.out',
    '--card-tilt-x': `${(x / rect.width - 0.5) * 20}deg`,
    '--card-tilt-y': `${-(y / rect.height - 0.5) * 20}deg`,
  })
  
  // 增强悬停效果：通过增加图像的初始缩放比例来实现轻微放大效果
  const imgElement = el.querySelector('.artwork-image') as HTMLImageElement
  if (imgElement) {
    gsap.to(imgElement, {
      scale: 1.04, // Slightly more than the default 1.1 to create a subtle zoom effect
      duration: 0.3,
      ease: 'power2.out'
    })
  }
}

const onCardLeave = (e: MouseEvent) => {
  const el = e.currentTarget as HTMLElement
  gsap.to(el, {
    scale: 1,
    x: 0,
    y: 0,
    rotationX: 0,
    rotationY: 0,
    duration: 0.6,
    ease: 'elastic.out(1, 0.6)',
    '--card-tilt-x': '0deg',
    '--card-tilt-y': '0deg',
  })
  
  // 恢复图像缩放
  const imgElement = el.querySelector('.artwork-image') as HTMLImageElement
  if (imgElement) {
    gsap.to(imgElement, {
      scale: 1.04, // Return to the original loaded scale
      duration: 0.6,
      ease: 'elastic.out(1, 0.6)'
    })
  }
}

/* ---------- 图片加载动画 ---------- */
const onImgLoaded = (e: Event) => {
  const img = e.target as HTMLImageElement
  img.classList.add('loaded')
  
  // 添加淡入动画
  gsap.fromTo(img, 
    { opacity: 0, scale: 1.1 },
    { opacity: 1, scale: 1, duration: 0.8, ease: 'power2.out' }
  )
}


/* ---------- 模态框功能 ---------- */
const openDetailModal = (item: any) => {
  selectedArtwork.value = item
  // 重置模态框状态
  modalScale.value = 1
  modalTranslate.value = { x: 0, y: 0 }
  document.body.style.overflow = 'hidden' // 防止背景滚动
  document.body.style.touchAction = 'none' // 禁止浏览器默认触摸行为
}

const closeDetailModal = () => {
  selectedArtwork.value = null
  isOriginalImage.value = false // 重置为缩略图模式
  document.body.style.overflow = '' // 恢复背景滚动
  document.body.style.touchAction = '' // 恢复浏览器默认触摸行为
}

const handleModalScroll = (e: WheelEvent) => {
  e.preventDefault()
  const delta = e.deltaY > 0 ? 0.9 : 1.1
  const newScale = Math.max(0.1, Math.min(5, modalScale.value * delta))
  modalScale.value = newScale
}

const startDrag = (e: MouseEvent) => {
  // 防止事件冒泡
  e.preventDefault()
  e.stopPropagation()
  
  isDragging.value = true
  startX.value = e.clientX
  startY.value = e.clientY
  startTranslateX.value = modalTranslate.value.x
  startTranslateY.value = modalTranslate.value.y
  
  // 添加全局事件监听器
  document.addEventListener('mousemove', drag)
  document.addEventListener('mouseup', stopDrag)
  document.addEventListener('mouseleave', stopDrag)
}

const drag = (e: MouseEvent) => {
  // 防止事件冒泡
  e.preventDefault()
  e.stopPropagation()
  
  if (!isDragging.value) return
  const dx = e.clientX - startX.value
  const dy = e.clientY - startY.value
  modalTranslate.value = {
    x: startTranslateX.value + dx,
    y: startTranslateY.value + dy
  }
}

const stopDrag = () => {
  isDragging.value = false
  // 移除全局事件监听器
  document.removeEventListener('mousemove', drag)
  document.removeEventListener('mouseup', stopDrag)
  document.removeEventListener('mouseleave', stopDrag)
}

// 触摸事件状态
const touchStartDistance = ref(0)
const touchStartScale = ref(1)
const initialPinchCenter = ref({ x: 0, y: 0 })

// 触摸开始
const handleTouchStart = (e: TouchEvent) => {
  e.preventDefault()
  
  if (e.touches.length === 1 && e.touches[0]) {
    // 单指拖拽
    isDragging.value = true
    startX.value = e.touches[0].clientX
    startY.value = e.touches[0].clientY
    startTranslateX.value = modalTranslate.value.x
    startTranslateY.value = modalTranslate.value.y
  } else if (e.touches.length === 2 && e.touches[0] && e.touches[1]) {
    // 双指缩放
    const dx = e.touches[0].clientX - e.touches[1].clientX
    const dy = e.touches[0].clientY - e.touches[1].clientY
    touchStartDistance.value = Math.sqrt(dx * dx + dy * dy)
    touchStartScale.value = modalScale.value
    
    // 计算双指中心点
    initialPinchCenter.value = {
      x: (e.touches[0].clientX + e.touches[1].clientX) / 2,
      y: (e.touches[0].clientY + e.touches[1].clientY) / 2
    }
  }
}

// 触摸移动
const handleTouchMove = (e: TouchEvent) => {
  e.preventDefault()
  
  if (e.touches.length === 1 && isDragging.value && e.touches[0]) {
    // 单指拖拽
    const dx = e.touches[0].clientX - startX.value
    const dy = e.touches[0].clientY - startY.value
    modalTranslate.value = {
      x: startTranslateX.value + dx,
      y: startTranslateY.value + dy
    }
  } else if (e.touches.length === 2 && e.touches[0] && e.touches[1]) {
    // 双指缩放
    const dx = e.touches[0].clientX - e.touches[1].clientX
    const dy = e.touches[0].clientY - e.touches[1].clientY
    const currentDistance = Math.sqrt(dx * dx + dy * dy)
    
    if (touchStartDistance.value > 0) {
      const scale = currentDistance / touchStartDistance.value
      const newScale = Math.max(0.1, Math.min(5, touchStartScale.value * scale))
      modalScale.value = newScale
    }
  }
}

// 触摸结束
const handleTouchEnd = (e: TouchEvent) => {
  e.preventDefault()
  
  if (e.touches.length < 2) {
    touchStartDistance.value = 0
  }
  if (e.touches.length === 0) {
    isDragging.value = false
  }
}

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 查看原图
const viewOriginalImage = (_: any) => {
  // 只有当当前不是原图模式时，才切换并显示加载状态
  if (!isOriginalImage.value) {
    isImageLoading.value = true // 开始加载
    isOriginalImage.value = true // 切换到原图模式
  }
}

// 查看缩略图
const viewThumbnailImage = (_: any) => {
  // 只有当当前是原图模式时，才切换并显示加载状态
  if (isOriginalImage.value) {
    isImageLoading.value = true // 开始加载
    isOriginalImage.value = false // 切换到缩略图模式
  }
}

/* ---------- 搜索功能 ---------- */
const activateSearch = () => {
  isSearchActive.value = true
}

const deactivateSearch = () => {
  isSearchActive.value = false
}

// 搜索加载指定页
const loadSearchPage = async (targetPage: number, isAppend = false) => {
  if (loading.value) return
  if (loadedPages.value.has(targetPage)) return
  
  loading.value = true
  
  try {
    const res = await galleryApi.getGallerySearch(username.value, {
      query: searchQuery.value.trim(),
      page: targetPage,
      page_size: pageSize
    })
    
    const images = res.data.images || []
    
    if (isAppend) {
      list.value.push(...images)
    } else {
      list.value = images
    }
    
    loadedPages.value.add(targetPage)
    
    // 更新总页数
    const totalCount = res.data.total || 0
    totalPages.value = Math.ceil(totalCount / pageSize)
    
    // 更新当前显示的页码为当前请求的页码
    page.value = targetPage
    
    if (images.length < pageSize) {
      finished.value = true
    }
  } catch (error) {
    console.error('Failed to search images:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = async () => {
  if (!searchQuery.value.trim()) return
  
  isInitialLoading.value = true
  isSearching.value = true
  list.value = []
  page.value = 1
  loadedPages.value.clear()
  finished.value = false
  
  await loadSearchPage(1, false)
  
  // 延迟关闭初始加载标志
  setTimeout(() => {
    isInitialLoading.value = false
  }, 500)
}

// 处理分页点击
const handlePageChange = async (pageNum: number) => {
  if (pageNum === page.value) return
  
  // 清空已加载页码记录
  loadedPages.value.clear()
  finished.value = false
  isInitialLoading.value = true
  
  // 重置页面滚动距离到顶部
  window.scrollTo({ top: 0, behavior: 'smooth' })
  
  if (isSearching.value) {
    await loadSearchPage(pageNum, false)
  } else {
    await loadPage(currentAlbumId.value || 0, pageNum, false)
  }
  
  // 延迟关闭初始加载标志
  setTimeout(() => {
    isInitialLoading.value = false
  }, 500)
}

const clearSearch = () => {
  searchQuery.value = ''
  isSearching.value = false
  loadedPages.value.clear()
  isInitialLoading.value = true
  if (currentAlbumId.value) {
    handleAlbumChange(currentAlbumId.value)
  }
}

// 处理页码输入
const handlePageInput = () => {
  const pageNum = parseInt(inputPage.value)
  if (!isNaN(pageNum) && pageNum >= 1 && pageNum <= totalPages.value) {
    handlePageChange(pageNum)
    inputPage.value = ''
  } else {
    inputPage.value = ''
  }
}

// 无限滚动加载搜索的下一页
const fetchSearchResults = async () => {
  if (!searchQuery.value.trim() || loading.value || finished.value) return
  
  const nextPageNum = page.value + 1
  if (nextPageNum > totalPages.value) {
    finished.value = true
    return
  }
  
  await loadSearchPage(nextPageNum, true)
}

const handleUploadClick = () => {
  router.push({ name: 'Upload' })
}

/* ---------- 生命周期与初始化 ---------- */
// 获取画廊配置
const fetchGalleryConfig = async () => {
  try {
    const res = await galleryApi.getGalleryConfig()
    galleryConfig.value = res.data
  } catch (error) {
    console.error('Failed to fetch gallery config:', error)
  }
}

onMounted(async () => {
  // 获取画廊配置
  await fetchGalleryConfig()
  
  // 设置无限滚动观察器
  observer.value = new IntersectionObserver(
    (entries) => {
      if (entries[0] && entries[0].isIntersecting && !loading.value && !finished.value && !isInitialLoading.value) {
        if (isSearching.value) {
          fetchSearchResults()
        } else if (currentAlbumId.value) {
          loadNextPage(currentAlbumId.value)
        }
      }
    },
    { rootMargin: '300px' }
  )

  // 监听 loadMoreRef 的变化，当元素渲染后开始观察
  watch(loadMoreRef, (el) => {
    if (el && observer.value) {
      observer.value.observe(el)
    }
  })
  
  // 获取数据（放在观察器设置之后，避免初始加载时触发无限滚动）
  try {
    const res = await galleryApi.getGalleryAlbums(username.value)
    albums.value = res.data.albums
    
    if (albums.value.length) {
      await handleAlbumChange(albums.value[0].id)
    }
  } catch (error:any) {
    console.error('Failed to fetch albums:', error)
    const errorCode = error.response?.data?.code;
    if (errorCode == "USER_NOT_FOUND"){
        router.push({ name: 'NotFound' })
    }
  }
})


</script>