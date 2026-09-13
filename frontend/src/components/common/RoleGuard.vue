<template>
  <slot v-if="allowed"></slot>
  <el-empty v-else description="无权限查看该内容" :image-size="80" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuth } from '@/hooks/useAuth'
import type { UserRoleValue } from '@/constants/user'

const props = defineProps<{ roles: UserRoleValue[] }>()
const { hasRole } = useAuth()
const allowed = computed(() => hasRole(props.roles))
</script>
