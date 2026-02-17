<template>
  <div 
    ref="containerRef" 
    class="sphere-tag-cloud-container"
    @mousedown="onMouseDown"
    @wheel="onWheel"
    @mouseenter="isHovering = true"
    @mouseleave="isHovering = false; isDragging = false"
    @touchstart="onTouchStart"
    @touchmove="onTouchMove"
    @touchend="onTouchEnd"
  >
    <!-- 背景效果 -->
    <div class="background-effect">
      <div class="background-glow"></div>
      <div class="background-grid"></div>
    </div>
    
    <canvas ref="canvasRef" class="tag-canvas"></canvas>
    
    <Transition name="fade">
      <div v-if="activeTag && !isDragging" class="tag-tooltip" :style="tooltipStyle">
        <div class="tooltip-content">
          <div class="tooltip-tag">{{ activeTag.element.tag }}</div>
          <div class="tooltip-count">{{ activeTag.element.count }}</div>
        </div>
      </div>
    </Transition>

    <div class="controls-hint">
      <span class="hint-text">{{ isMobile ? '触摸旋转 / 双指缩放' : '拖拽旋转 / 滚轮缩放' }}</span>
      <div class="hint-icons">
        <span class="hint-icon">{{ isMobile ? '👆' : '🖱️' }}</span>
        <span class="hint-icon">{{ isMobile ? '🔍' : '🔍' }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'

interface Tag {
  tag: string
  count: number
}

const props = defineProps<{
  tags: Tag[]
}>()

const containerRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const isHovering = ref(false)
const isDragging = ref(false)
const activeTag = ref<TagItem | null>(null)
const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 768
}

// 变换状态
const rotation = ref({ x: 0, y: 0 })
const velocity = ref({ x: 0.003, y: 0.003 }) 
const zoom = ref(1)
const mousePos = ref({ x: 0, y: 0 })

interface TagItem {
  x: number; y: number; z: number;
  tx: number; ty: number; tz: number;
  cx: number; cy: number;
  scale: number;
  alpha: number;
  blur: number;
  fontSize: number;
  element: Tag;
  baseHue: number;
}

let tags3D: TagItem[] = []
let animationId: number | null = null
const FOCAL_LENGTH = 500
const BASE_RADIUS = 160

function initTags() {
  const n = props.tags.length
  if (n === 0) return

  const maxCount = Math.max(...props.tags.map(t => t.count))
  const minCount = Math.min(...props.tags.map(t => t.count))

  tags3D = props.tags.map((tag, i) => {
    // 斐波那契螺旋
    const phi = Math.acos(-1 + (2 * i) / n)
    const theta = Math.sqrt(n * Math.PI) * phi

    const x = Math.cos(theta) * Math.sin(phi)
    const y = Math.sin(theta) * Math.sin(phi)
    const z = Math.cos(phi)

    const fontSize = 14 + (tag.count - minCount) / (maxCount - minCount || 1) * 18
    
    return {
      x: x * BASE_RADIUS,
      y: y * BASE_RADIUS,
      z: z * BASE_RADIUS,
      tx: 0, ty: 0, tz: 0, cx: 0, cy: 0,
      scale: 1, alpha: 1, blur: 0,
      fontSize,
      element: tag,
      baseHue: Math.random() * 360
    }
  })
}

function render() {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d', { alpha: true })
  if (!canvas || !ctx || !containerRef.value) return

  const rect = containerRef.value.getBoundingClientRect()
  if (canvas.width !== rect.width) {
    canvas.width = rect.width
    canvas.height = rect.height
  }

  const { width, height } = canvas
  const centerX = width / 2
  const centerY = height / 2

  ctx.clearRect(0, 0, width, height)

  // 1. 动力学处理
  if (!isDragging.value) {
    const targetVel = isHovering.value ? 0.001 : 0.004
    velocity.value.x += (targetVel - velocity.value.x) * 0.05
    velocity.value.y += (targetVel - velocity.value.y) * 0.05
    rotation.value.x += velocity.value.x
    rotation.value.y += velocity.value.y
  }

  // 2. 3D 变换与景深计算
  tags3D.forEach(item => {
    // 旋转变换
    let y1 = item.y * Math.cos(rotation.value.x) - item.z * Math.sin(rotation.value.x)
    let z1 = item.y * Math.sin(rotation.value.x) + item.z * Math.cos(rotation.value.x)
    let x2 = item.x * Math.cos(rotation.value.y) - z1 * Math.sin(rotation.value.y)
    let z2 = item.x * Math.sin(rotation.value.y) + z1 * Math.cos(rotation.value.y)

    item.tx = x2; item.ty = y1; item.tz = z2

    const perspective = FOCAL_LENGTH / (FOCAL_LENGTH + z2 * zoom.value)
    item.scale = perspective * zoom.value
    item.cx = centerX + x2 * item.scale
    item.cy = centerY + y1 * item.scale
    
    // 景深算法：
    // z2 越小（越负）代表越近。z2 越大（越正）代表越远。
    // 计算 0 (最前) 到 1 (最后) 的归一化距离
    const normalizedDepth = (z2 + BASE_RADIUS) / (2 * BASE_RADIUS)
    
    // 越远越透明 (0.2 ~ 1.0)
    item.alpha = 1.0 - (normalizedDepth * 0.7)
    // 越远越模糊 (0px ~ 1px)
    item.blur = normalizedDepth * 1
  })

  // 3. 排序 (后进先出绘制)
  const sortedTags = [...tags3D].sort((a, b) => a.tz - b.tz)

  // 4. 绘制
  sortedTags.forEach(item => {
    ctx.save()
    ctx.translate(item.cx, item.cy)
    
    // 应用模糊滤镜
    if (item.blur > 0.5) {
      ctx.filter = `blur(${item.blur.toFixed(1)}px)`
    } else {
      ctx.filter = 'none'
    }
    
    const isMousedOver = activeTag.value === item
    const fontSize = Math.max(3, item.fontSize * item.scale)
    
    ctx.font = `600 ${fontSize}px "Segoe UI", Roboto, Helvetica, Arial, sans-serif`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    
    // 文字颜色：使用较深且饱和的颜色以适配白底
    // 越远颜色越浅
    const lightness = 30 + (1 - item.alpha) * 40 
    ctx.fillStyle = `hsla(${item.baseHue}, 60%, ${isMousedOver ? 10 : lightness}%, ${item.alpha})`
    
    // 如果被选中，轻微放大
    const finalScale = isMousedOver ? 1.2 : 1
    ctx.scale(finalScale, finalScale)

    ctx.fillText(item.element.tag, 0, 0)
    ctx.restore()
  })

  // 5. 交互检测
  if (!isDragging.value) {
    let hit = null
    // 倒序检测（先检测前面的）
    const hitTestList = [...tags3D].sort((a, b) => a.tz - b.tz)
    for (const item of hitTestList) {
      const dx = item.cx - mousePos.value.x
      const dy = item.cy - mousePos.value.y
      const dist = Math.sqrt(dx * dx + dy * dy)
      // 只允许点击比较清晰的标签 (前面一半的球体)
      if (dist < (item.fontSize * item.scale) / 2 + 5 && item.tz < 50) {
        hit = item
        break
      }
    }
    activeTag.value = hit
  }

  animationId = requestAnimationFrame(render)
}

// --- 事件监听 ---
const onMouseDown = (e: MouseEvent) => {
  isDragging.value = true
  let lastX = e.clientX
  let lastY = e.clientY

  const onMouseMove = (me: MouseEvent) => {
    const dx = me.clientX - lastX
    const dy = me.clientY - lastY
    rotation.value.y += dx * 0.008
    rotation.value.x -= dy * 0.008
    velocity.value = { x: -dy * 0.004, y: dx * 0.004 }
    lastX = me.clientX
    lastY = me.clientY
    updateMouseCoord(me)
  }

  const onMouseUp = () => {
    isDragging.value = false
    window.removeEventListener('mousemove', onMouseMove)
    window.removeEventListener('mouseup', onMouseUp)
  }
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
}

const onWheel = (e: WheelEvent) => {
  e.preventDefault()
  const newZoom = zoom.value * (e.deltaY > 0 ? 0.95 : 1.05)
  if (newZoom > 0.6 && newZoom < 2.5) zoom.value = newZoom
}

// 触摸事件处理
let lastTouchX = 0
let lastTouchY = 0
let lastTouchDist = 0

const onTouchStart = (e: TouchEvent) => {
  e.preventDefault()
  isDragging.value = true
  
  if (e.touches && e.touches.length === 1 && e.touches[0]) {
    lastTouchX = e.touches[0].clientX
    lastTouchY = e.touches[0].clientY
  } else if (e.touches && e.touches.length === 2 && e.touches[0] && e.touches[1]) {
    const dx = e.touches[0].clientX - e.touches[1].clientX
    const dy = e.touches[0].clientY - e.touches[1].clientY
    lastTouchDist = Math.sqrt(dx * dx + dy * dy)
  }
}

const onTouchMove = (e: TouchEvent) => {
  e.preventDefault()
  
  if (e.touches && e.touches.length === 1 && e.touches[0]) {
    const touch = e.touches[0]
    const dx = touch.clientX - lastTouchX
    const dy = touch.clientY - lastTouchY
    rotation.value.y += dx * 0.008
    rotation.value.x -= dy * 0.008
    velocity.value = { x: -dy * 0.004, y: dx * 0.004 }
    lastTouchX = touch.clientX
    lastTouchY = touch.clientY
    
    if (containerRef.value) {
      const rect = containerRef.value.getBoundingClientRect()
      mousePos.value = { x: touch.clientX - rect.left, y: touch.clientY - rect.top }
    }
  } else if (e.touches && e.touches.length === 2 && e.touches[0] && e.touches[1]) {
    const dx = e.touches[0].clientX - e.touches[1].clientX
    const dy = e.touches[0].clientY - e.touches[1].clientY
    const dist = Math.sqrt(dx * dx + dy * dy)
    
    if (lastTouchDist > 0) {
      const scale = dist / lastTouchDist
      const newZoom = zoom.value * scale
      if (newZoom > 0.6 && newZoom < 2.5) zoom.value = newZoom
    }
    lastTouchDist = dist
  }
}

const onTouchEnd = (e: TouchEvent) => {
  e.preventDefault()
  isDragging.value = false
  lastTouchDist = 0
}

const updateMouseCoord = (e: MouseEvent) => {
  if (containerRef.value) {
    const rect = containerRef.value.getBoundingClientRect()
    mousePos.value = { x: e.clientX - rect.left, y: e.clientY - rect.top }
  }
}

const tooltipStyle = computed(() => ({
  left: `${mousePos.value.x + 15}px`,
  top: `${mousePos.value.y + 15}px`
}))

onMounted(() => {
  initTags()
  render()
  checkMobile()
  containerRef.value?.addEventListener('mousemove', updateMouseCoord)
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  window.removeEventListener('resize', checkMobile)
})

watch(() => props.tags, initTags, { deep: true })
</script>

<style scoped>
.sphere-tag-cloud-container {
  position: relative;
  width: 100%;
  height: 550px;
  background: 
    radial-gradient(circle at 20% 30%, rgba(147, 197, 253, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 70%, rgba(167, 139, 250, 0.1) 0%, transparent 50%),
    linear-gradient(135deg, #fefefe 0%, #f0f9ff 100%);
  border-radius: 28px;
  overflow: hidden;
  cursor: grab;
  user-select: none;
  touch-action: none;
  border: 1px solid rgba(255, 255, 255, 0.4);
  box-shadow: 
    0 12px 35px rgba(0, 0, 0, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    inset 0 -1px 0 rgba(0, 0, 0, 0.05);
  transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.sphere-tag-cloud-container:hover {
  box-shadow: 
    0 15px 35px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 rgba(0, 0, 0, 0.05);
}

.sphere-tag-cloud-container:active {
  cursor: grabbing;
  transform: scale(0.98);
}

/* 背景效果 */
.background-effect {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
  pointer-events: none;
}

.background-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 60%;
  height: 60%;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.1) 0%, transparent 70%);
  border-radius: 50%;
  animation: pulse 4s ease-in-out infinite;
}

.background-grid {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: 
    linear-gradient(rgba(0, 0, 0, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.03) 1px, transparent 1px);
  background-size: 30px 30px;
  opacity: 0.3;
}

@keyframes pulse {
  0%, 100% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.5;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.1);
    opacity: 0.8;
  }
}

.tag-canvas {
  position: relative;
  width: 100%;
  height: 100%;
  z-index: 2;
}

/* 增强的 Tooltip 样式 */
.tag-tooltip {
  position: absolute;
  pointer-events: none;
  z-index: 100;
  white-space: nowrap;
}

.tooltip-content {
  background: var(--primary-dark);
  color: white;
  padding: 8px 12px;
  border-radius: 12px;
  box-shadow: 
    0 10px 25px rgba(0, 0, 0, 0.2),
    0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
}

.tooltip-tag {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 4px;
  color: #f8fafc;
}

.tooltip-count {
  font-size: 12px;
  font-weight: 400;
  color: #f8fafc;
}

/* 增强的控制提示 */
.controls-hint {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3;
  pointer-events: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.hint-text {
  color: #64748b;
  font-size: 12px;
  letter-spacing: 1px;
  text-transform: uppercase;
  font-weight: 500;
}

.hint-icons {
  display: flex;
  gap: 12px;
}

.hint-icon {
  font-size: 16px;
  opacity: 0.7;
  transition: opacity 0.3s ease;
}

.sphere-tag-cloud-container:hover .hint-icon {
  opacity: 1;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}
</style>