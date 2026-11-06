<script setup lang="ts">
import { cn } from '@/lib/utils'
import { badgeVariants, type BadgeVariants } from './index'

interface BadgeProps {
  variant?: BadgeVariants['variant']
  class?: string
}

const props = defineProps<BadgeProps>()
</script>

<template>
  <div :class="cn(badgeVariants({ variant: props.variant }), props.class)">
    <slot />
  </div>
</template>
