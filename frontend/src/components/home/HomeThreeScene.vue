<template>
  <canvas ref="canvasRef" class="home-three-scene" aria-hidden="true"></canvas>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const canvasRef = ref<HTMLCanvasElement | null>(null)

let cleanupScene: (() => void) | null = null

onMounted(async () => {
  const canvas = canvasRef.value
  if (!canvas) return

  const THREE = await import('three')
  const scene = new THREE.Scene()
  const camera = new THREE.PerspectiveCamera(38, 1, 0.1, 100)
  camera.position.set(0, 0, 8.4)

  const renderer = new THREE.WebGLRenderer({
    alpha: true,
    antialias: true,
    canvas,
    powerPreference: 'high-performance'
  })
  renderer.setClearColor(0x000000, 0)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.8))
  renderer.outputColorSpace = THREE.SRGBColorSpace
  renderer.toneMapping = THREE.ACESFilmicToneMapping
  renderer.toneMappingExposure = 1.15

  type Disposable = { dispose: () => void }
  const geometries: Disposable[] = []
  const materials: Disposable[] = []
  const trackGeometry = <T extends Disposable>(value: T) => {
    geometries.push(value)
    return value
  }
  const trackMaterial = <T extends Disposable>(value: T) => {
    materials.push(value)
    return value
  }

  const root = new THREE.Group()
  const network = new THREE.Group()
  root.add(network)
  scene.add(root)

  const coreMaterial = trackMaterial(new THREE.MeshStandardMaterial({
    color: 0x89ffe5,
    emissive: 0x1bc9a2,
    emissiveIntensity: 2.8,
    metalness: 0.2,
    roughness: 0.18
  }))
  const core = new THREE.Mesh(trackGeometry(new THREE.IcosahedronGeometry(0.62, 2)), coreMaterial)
  network.add(core)

  const coreShellMaterial = trackMaterial(new THREE.MeshBasicMaterial({
    color: 0xd9fff6,
    transparent: true,
    opacity: 0.28,
    wireframe: true,
    blending: THREE.AdditiveBlending,
    depthWrite: false
  }))
  const coreShell = new THREE.Mesh(
    trackGeometry(new THREE.IcosahedronGeometry(0.98, 2)),
    coreShellMaterial
  )
  network.add(coreShell)

  const haloMaterial = trackMaterial(new THREE.MeshBasicMaterial({
    color: 0x65f4d2,
    transparent: true,
    opacity: 0.44,
    blending: THREE.AdditiveBlending,
    depthWrite: false
  }))
  const haloA = new THREE.Mesh(trackGeometry(new THREE.TorusGeometry(1.18, 0.012, 8, 160)), haloMaterial)
  haloA.rotation.x = Math.PI / 2.25
  network.add(haloA)
  const haloB = new THREE.Mesh(
    trackGeometry(new THREE.TorusGeometry(1.48, 0.008, 8, 180)),
    trackMaterial(haloMaterial.clone())
  )
  haloB.rotation.set(Math.PI / 2.7, Math.PI / 5, 0)
  network.add(haloB)

  const nodePositions = [
    new THREE.Vector3(2.55, 1.42, 0.18),
    new THREE.Vector3(2.82, -0.92, -0.08),
    new THREE.Vector3(-1.08, -2.02, 0.12),
    new THREE.Vector3(-2.58, 0.74, -0.18)
  ]
  const nodeColors = [0xd9ff66, 0xff786b, 0x8ea5ff, 0xffffff]
  const nodeMeshes = nodePositions.map((position, index) => {
    const group = new THREE.Group()
    group.position.copy(position)

    const material = trackMaterial(new THREE.MeshStandardMaterial({
      color: nodeColors[index],
      emissive: nodeColors[index],
      emissiveIntensity: 1.45,
      metalness: 0.08,
      roughness: 0.28
    }))
    const node = new THREE.Mesh(trackGeometry(new THREE.OctahedronGeometry(0.22, 1)), material)
    group.add(node)

    const ring = new THREE.Mesh(
      trackGeometry(new THREE.TorusGeometry(0.42, 0.009, 6, 80)),
      trackMaterial(new THREE.MeshBasicMaterial({
        color: nodeColors[index],
        transparent: true,
        opacity: 0.42,
        blending: THREE.AdditiveBlending,
        depthWrite: false
      }))
    )
    ring.rotation.x = Math.PI / 2
    group.add(ring)
    network.add(group)
    return { group, node, ring }
  })

  const connectionMaterial = trackMaterial(new THREE.MeshBasicMaterial({
    color: 0x7ff6da,
    transparent: true,
    opacity: 0.18,
    blending: THREE.AdditiveBlending,
    depthWrite: false
  }))

  const curves = nodePositions.map((position, index) => {
    const direction = index % 2 === 0 ? 1 : -1
    const curve = new THREE.CubicBezierCurve3(
      new THREE.Vector3(0, 0, 0),
      new THREE.Vector3(position.x * 0.28, position.y * 0.18 + direction * 0.78, 0.72 * direction),
      new THREE.Vector3(position.x * 0.72, position.y * 0.72 - direction * 0.32, -0.38 * direction),
      position
    )
    const path = new THREE.Mesh(
      trackGeometry(new THREE.TubeGeometry(curve, 64, 0.008, 6, false)),
      index === 0 ? connectionMaterial : trackMaterial(connectionMaterial.clone())
    )
    network.add(path)
    return curve
  })

  const packets = curves.flatMap((curve, curveIndex) => {
    return Array.from({ length: 3 }, (_, packetIndex) => {
      const material = trackMaterial(new THREE.MeshBasicMaterial({
        color: nodeColors[curveIndex],
        transparent: true,
        opacity: 0.95,
        blending: THREE.AdditiveBlending,
        depthWrite: false
      }))
      const mesh = new THREE.Mesh(trackGeometry(new THREE.SphereGeometry(0.045, 10, 10)), material)
      network.add(mesh)
      return {
        curve,
        mesh,
        offset: packetIndex / 3 + curveIndex * 0.11,
        speed: 0.11 + curveIndex * 0.012
      }
    })
  })

  let seed = 73421
  const random = () => {
    seed = (seed * 16807) % 2147483647
    return (seed - 1) / 2147483646
  }
  const pointCount = 240
  const pointPositions = new Float32Array(pointCount * 3)
  for (let index = 0; index < pointCount; index += 1) {
    pointPositions[index * 3] = (random() - 0.5) * 12
    pointPositions[index * 3 + 1] = (random() - 0.5) * 7
    pointPositions[index * 3 + 2] = -1.2 - random() * 3.8
  }
  const pointGeometry = trackGeometry(new THREE.BufferGeometry())
  pointGeometry.setAttribute('position', new THREE.BufferAttribute(pointPositions, 3))
  const points = new THREE.Points(
    pointGeometry,
    trackMaterial(new THREE.PointsMaterial({
      color: 0xb8d7ce,
      size: 0.026,
      transparent: true,
      opacity: 0.45,
      depthWrite: false
    }))
  )
  scene.add(points)

  scene.add(new THREE.AmbientLight(0xb9fff0, 0.45))
  const coreLight = new THREE.PointLight(0x63f6d5, 12, 8)
  coreLight.position.set(0, 0.2, 2.2)
  scene.add(coreLight)
  const warmLight = new THREE.PointLight(0xff6d61, 7, 7)
  warmLight.position.set(3.2, -1.8, 2.4)
  scene.add(warmLight)
  const limeLight = new THREE.PointLight(0xd9ff66, 5, 6)
  limeLight.position.set(2.4, 2.1, 1.8)
  scene.add(limeLight)

  const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  const pointer = new THREE.Vector2(0, 0)
  let frameId = 0

  const handlePointerMove = (event: PointerEvent) => {
    pointer.x = (event.clientX / window.innerWidth - 0.5) * 2
    pointer.y = (event.clientY / window.innerHeight - 0.5) * 2
  }

  const resize = () => {
    const rect = canvas.getBoundingClientRect()
    const width = Math.max(1, Math.floor(rect.width))
    const height = Math.max(1, Math.floor(rect.height))
    renderer.setSize(width, height, false)
    camera.aspect = width / height
    camera.position.z = camera.aspect < 0.9 ? 9.2 : 8.4
    root.position.set(camera.aspect > 1.18 ? 1.25 : 0, camera.aspect < 0.9 ? -0.42 : 0, 0)
    root.scale.setScalar(camera.aspect < 0.9 ? 0.76 : 1)
    camera.updateProjectionMatrix()
  }

  const resizeObserver = new ResizeObserver(resize)
  resizeObserver.observe(canvas)
  window.addEventListener('pointermove', handlePointerMove, { passive: true })
  resize()

  const animate = (timeMs = 0) => {
    const time = timeMs * 0.001
    root.rotation.x += (-pointer.y * 0.045 - root.rotation.x) * 0.035
    root.rotation.y += (pointer.x * 0.075 - root.rotation.y) * 0.035
    network.rotation.y = time * 0.045
    core.rotation.x = time * 0.46
    core.rotation.y = time * 0.62
    coreShell.rotation.x = -time * 0.16
    coreShell.rotation.y = time * 0.2
    coreShell.scale.setScalar(1 + Math.sin(time * 1.8) * 0.035)
    haloA.rotation.z = time * 0.34
    haloB.rotation.z = -time * 0.22
    points.rotation.y = time * 0.006

    nodeMeshes.forEach(({ group, node, ring }, index) => {
      const phase = time * (0.72 + index * 0.08) + index
      node.rotation.x = phase
      node.rotation.y = phase * 1.15
      ring.rotation.z = phase * (index % 2 === 0 ? 0.36 : -0.36)
      group.scale.setScalar(1 + Math.sin(phase * 1.9) * 0.055)
    })

    packets.forEach(({ curve, mesh, offset, speed }) => {
      const progress = (time * speed + offset) % 1
      mesh.position.copy(curve.getPoint(progress))
      mesh.scale.setScalar(0.7 + Math.sin(progress * Math.PI) * 0.7)
    })

    renderer.render(scene, camera)
    if (!prefersReducedMotion) {
      frameId = window.requestAnimationFrame(animate)
    }
  }

  animate()

  cleanupScene = () => {
    window.cancelAnimationFrame(frameId)
    window.removeEventListener('pointermove', handlePointerMove)
    resizeObserver.disconnect()
    geometries.forEach((geometry) => geometry.dispose())
    materials.forEach((material) => material.dispose())
    renderer.dispose()
  }
})

onBeforeUnmount(() => {
  cleanupScene?.()
  cleanupScene = null
})
</script>

<style scoped>
.home-three-scene {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
