import { useEffect, useRef } from 'react'
import './StarfieldBackground.css'

// 可调整参数
const CONFIG = {
  particleDensity: 0.0007,
  particleSpeedMin: 1.0,
  particleSpeedMax: 3.5,
  particleSizeMin: 0.5,
  particleSizeMax: 1.5,
  particleColor: 'rgba(255, 255, 255',
  mouseRadius: 150,
  mouseForce: 0.02,
  connectionDistanceBase: 7,
  connectionDistanceRatio: 0.021,
  connectionOpacity: 0.7,
  connectionColor: 'rgba(102, 126, 234',
  connectionLineWidth: 0.5,
  trailLength: 5,
  trailOpacity: 0.5,
  trailWidthMultiplier: 0.5,
  fadeSpeed: 0.1,
  depthRange: 1500,
  depthThreshold: 300,
  trailDepthThreshold: 500,
  projectionDistance: 1000,
  opacityDepthDivisor: 1000,
  glowSizeMultiplier: 2,
  glowOpacity: 0.2,
  glowMinSize: 1
}

function StarfieldBackground() {
  const canvasRef = useRef(null)
  const particlesRef = useRef([])
  const mouseRef = useRef({ x: null, y: null, radius: CONFIG.mouseRadius })
  const animationFrameRef = useRef(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    let width, height
    let dynamicConnectionDistance = 150

    // 粒子类
    class Particle {
      constructor() {
        this.reset()
        this.y = Math.random() * height
      }

      reset() {
        this.x = Math.random() * width
        this.y = Math.random() * height
        this.z = Math.random() * CONFIG.depthRange
        this.size = Math.random() * (CONFIG.particleSizeMax - CONFIG.particleSizeMin) + CONFIG.particleSizeMin
        this.speed = Math.random() * (CONFIG.particleSpeedMax - CONFIG.particleSpeedMin) + CONFIG.particleSpeedMin
        this.color = CONFIG.particleColor
      }

      update() {
        this.z -= this.speed
        
        if (this.z <= 0) {
          this.reset()
          this.z = CONFIG.depthRange
        }

        const scale = CONFIG.projectionDistance / (CONFIG.projectionDistance + this.z)
        this.projectedX = (this.x - width / 2) * scale + width / 2
        this.projectedY = (this.y - height / 2) * scale + height / 2
        this.projectedSize = this.size * scale

        const mouse = mouseRef.current
        if (mouse.x !== null && mouse.y !== null) {
          const dx = this.projectedX - mouse.x
          const dy = this.projectedY - mouse.y
          const distance = Math.sqrt(dx * dx + dy * dy)
          
          if (distance < mouse.radius) {
            const force = (mouse.radius - distance) / mouse.radius
            this.x += dx * force * CONFIG.mouseForce
            this.y += dy * force * CONFIG.mouseForce
          }
        }
      }

      draw() {
        const opacity = Math.min(1, (CONFIG.depthRange - this.z) / CONFIG.opacityDepthDivisor)
        
        ctx.beginPath()
        ctx.arc(this.projectedX, this.projectedY, this.projectedSize, 0, Math.PI * 2)
        ctx.fillStyle = `${this.color}, ${opacity})`
        ctx.fill()

        if (this.projectedSize > CONFIG.glowMinSize) {
          ctx.beginPath()
          ctx.arc(this.projectedX, this.projectedY, this.projectedSize * CONFIG.glowSizeMultiplier, 0, Math.PI * 2)
          ctx.fillStyle = `${this.color}, ${opacity * CONFIG.glowOpacity})`
          ctx.fill()
        }

        if (this.z < CONFIG.trailDepthThreshold) {
          const prevScale = CONFIG.projectionDistance / (CONFIG.projectionDistance + this.z + this.speed * CONFIG.trailLength)
          const prevX = (this.x - width / 2) * prevScale + width / 2
          const prevY = (this.y - height / 2) * prevScale + height / 2

          ctx.beginPath()
          ctx.moveTo(prevX, prevY)
          ctx.lineTo(this.projectedX, this.projectedY)
          ctx.strokeStyle = `${this.color}, ${opacity * CONFIG.trailOpacity})`
          ctx.lineWidth = this.projectedSize * CONFIG.trailWidthMultiplier
          ctx.stroke()
        }
      }
    }

    function connectParticles() {
      const particles = particlesRef.current
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].projectedX - particles[j].projectedX
          const dy = particles[i].projectedY - particles[j].projectedY
          const distance = Math.sqrt(dx * dx + dy * dy)

          if (distance < dynamicConnectionDistance &&
              particles[i].z < CONFIG.depthThreshold &&
              particles[j].z < CONFIG.depthThreshold) {
            const opacity = (dynamicConnectionDistance - distance) / dynamicConnectionDistance * CONFIG.connectionOpacity
            ctx.beginPath()
            ctx.moveTo(particles[i].projectedX, particles[i].projectedY)
            ctx.lineTo(particles[j].projectedX, particles[j].projectedY)
            ctx.strokeStyle = `${CONFIG.connectionColor}, ${opacity})`
            ctx.lineWidth = CONFIG.connectionLineWidth
            ctx.stroke()
          }
        }
      }
    }

    function init() {
      width = canvas.width = window.innerWidth
      height = canvas.height = window.innerHeight
      
      const screenArea = width * height
      const particleCount = Math.floor(screenArea * CONFIG.particleDensity)
      
      dynamicConnectionDistance = Math.max(80, width * CONFIG.connectionDistanceRatio)
      
      particlesRef.current = []
      for (let i = 0; i < particleCount; i++) {
        particlesRef.current.push(new Particle())
      }
    }

    function animate() {
      ctx.fillStyle = `rgba(9, 10, 15, ${CONFIG.fadeSpeed})`
      ctx.fillRect(0, 0, width, height)

      particlesRef.current.forEach(particle => {
        particle.update()
        particle.draw()
      })

      connectParticles()

      animationFrameRef.current = requestAnimationFrame(animate)
    }

    const handleResize = () => init()
    
    const handleMouseMove = (e) => {
      mouseRef.current.x = e.clientX
      mouseRef.current.y = e.clientY
    }

    const handleMouseLeave = () => {
      mouseRef.current.x = null
      mouseRef.current.y = null
    }

    const handleTouchMove = (e) => {
      e.preventDefault()
      mouseRef.current.x = e.touches[0].clientX
      mouseRef.current.y = e.touches[0].clientY
    }

    const handleTouchEnd = () => {
      mouseRef.current.x = null
      mouseRef.current.y = null
    }

    window.addEventListener('resize', handleResize)
    canvas.addEventListener('mousemove', handleMouseMove)
    canvas.addEventListener('mouseleave', handleMouseLeave)
    canvas.addEventListener('touchmove', handleTouchMove)
    canvas.addEventListener('touchend', handleTouchEnd)

    init()
    animate()

    return () => {
      window.removeEventListener('resize', handleResize)
      canvas.removeEventListener('mousemove', handleMouseMove)
      canvas.removeEventListener('mouseleave', handleMouseLeave)
      canvas.removeEventListener('touchmove', handleTouchMove)
      canvas.removeEventListener('touchend', handleTouchEnd)
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current)
      }
    }
  }, [])

  return <canvas ref={canvasRef} className="starfield-canvas" />
}

export default StarfieldBackground
