'use client'

import { useEffect, useRef, useState } from 'react'
import { animate, motion, useMotionValue, useTransform } from 'framer-motion'
import { cn } from '@/lib/utils'

interface AnimatedValueProps {
  value: number
  precision?: number
  prefix?: string
  suffix?: string
  className?: string
  duration?: number
  formatFunc?: (value: number) => string
}

export function AnimatedValue({
  value,
  precision = 2,
  prefix = '',
  suffix = '',
  className,
  duration = 0.5,
  formatFunc,
}: AnimatedValueProps) {
  const previousValue = useRef(value)
  const motionValue = useMotionValue(previousValue.current)
  const [isMounted, setIsMounted] = useState(false)

  // Format the displayed value
  const formattedValue = useTransform(motionValue, (latest) => {
    if (formatFunc) {
      return formatFunc(latest)
    }
    return `${prefix}${latest.toFixed(precision)}${suffix}`
  })

  // Trigger animation on value change
  useEffect(() => {
    if (!isMounted) {
      setIsMounted(true)
      return
    }

    const controls = animate(motionValue, value, {
      duration,
      ease: 'easeInOut',
    })

    previousValue.current = value

    return () => controls.stop()
  }, [value, motionValue, duration, isMounted])

  return (
    <motion.span className={cn(className)}>
      {formattedValue}
    </motion.span>
  )
}

export function AnimatedPercentage({
  value,
  className,
  showSign = true,
  duration = 0.5,
}: {
  value: number
  className?: string
  showSign?: boolean
  duration?: number
}) {
  const formatPercentage = (val: number) => {
    const sign = showSign && val > 0 ? '+' : ''
    return `${sign}${val.toFixed(2)}%`
  }

  return (
    <AnimatedValue
      value={value}
      formatFunc={formatPercentage}
      className={className}
      duration={duration}
    />
  )
}

export function AnimatedCurrency({
  value,
  currency = 'SOL',
  className,
  duration = 0.5,
}: {
  value: number
  currency?: string
  className?: string
  duration?: number
}) {
  const formatCurrency = (val: number) => {
    return `${val.toFixed(2)} ${currency}`
  }

  return (
    <AnimatedValue
      value={value}
      formatFunc={formatCurrency}
      className={className}
      duration={duration}
    />
  )
}