<script setup lang="ts">
const { user, authInitialized } = useAuth()

watchEffect(() => {
  if (!authInitialized.value) return
  if (user.value) {
    navigateTo(`/list/${user.value.username}`, { replace: true })
  } else {
    navigateTo('/', { replace: true })
  }
})
</script>

<template>
  <div />
</template>
