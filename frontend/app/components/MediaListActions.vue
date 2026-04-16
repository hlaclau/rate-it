<script setup lang="ts">
import type { ListEntry } from '~/composables/useList'

const props = defineProps<{
  externalId: string
  mediaType: 'movie' | 'series'
}>()

const { isAuthenticated } = useAuth()
const { addOrUpdate, remove, fetchStatus } = useList()

// Entry state
const listEntry = ref<ListEntry | null>(null)

const loadStatus = async () => {
  listEntry.value = await fetchStatus(props.externalId)
}

watch(isAuthenticated, (v) => { if (v) loadStatus() }, { immediate: true })

// Add / Edit modal
const editOpen = ref(false)
const formStatus = ref<'watched' | 'plan_to_watch'>('plan_to_watch')
const formRating = ref<number | null>(null)
const formReview = ref('')
const saving = ref(false)

const openEdit = () => {
  if (listEntry.value) {
    formStatus.value = listEntry.value.status
    formRating.value = listEntry.value.rating
    formReview.value = listEntry.value.review ?? ''
  } else {
    formStatus.value = 'plan_to_watch'
    formRating.value = null
    formReview.value = ''
  }
  editOpen.value = true
}

watch(formStatus, (s) => {
  if (s === 'plan_to_watch') formRating.value = null
})

const submitForm = async () => {
  saving.value = true
  try {
    await addOrUpdate({
      external_id: props.externalId,
      source: 'tmdb',
      type: props.mediaType,
      status: formStatus.value,
      rating: formStatus.value === 'watched' ? formRating.value : null,
      review: formReview.value.trim() || null,
    })
    await loadStatus()
    editOpen.value = false
  } finally {
    saving.value = false
  }
}

// Remove confirm modal
const removeOpen = ref(false)
const removing = ref(false)

const confirmRemove = async () => {
  if (!listEntry.value) return
  removing.value = true
  try {
    await remove(listEntry.value.media_id)
    listEntry.value = null
    removeOpen.value = false
  } finally {
    removing.value = false
  }
}
</script>

<template>
  <div v-if="isAuthenticated">

    <!-- Trigger area -->
    <div v-if="listEntry" class="flex flex-wrap items-center gap-3">
      <UBadge :color="listEntry.status === 'watched' ? 'success' : 'info'" size="lg">
        {{ listEntry.status === 'watched' ? 'Watched' : 'Plan to watch' }}
      </UBadge>
      <span v-if="listEntry.rating" class="flex items-center gap-1.5 font-semibold text-sm">
        <UIcon name="i-lucide-star" class="size-4 text-yellow-400 fill-yellow-400" />
        {{ listEntry.rating }}/10
      </span>
      <p v-if="listEntry.review" class="text-muted text-sm italic truncate max-w-xs">
        "{{ listEntry.review }}"
      </p>
      <div class="flex gap-2 ml-auto">
        <UButton size="sm" variant="soft" color="neutral" leading-icon="i-lucide-pencil" @click="openEdit">
          Edit
        </UButton>
        <UButton size="sm" variant="ghost" color="error" leading-icon="i-lucide-trash-2" @click="removeOpen = true">
          Remove
        </UButton>
      </div>
    </div>

    <UButton v-else leading-icon="i-lucide-bookmark-plus" @click="openEdit">
      Add to list
    </UButton>

    <!-- Add / Edit modal -->
    <UModal v-model:open="editOpen" :title="listEntry ? 'Edit list entry' : 'Add to your list'" :description="listEntry ? 'Update your status, rating or review.' : 'Track this title in your personal list.'">
      <template #body>
        <div class="flex flex-col gap-5">
          <UFormGroup label="Status">
            <USelect
              v-model="formStatus"
              :items="[
                { label: 'Plan to watch', value: 'plan_to_watch' },
                { label: 'Watched', value: 'watched' },
              ]"
              value-key="value"
              label-key="label"
            />
          </UFormGroup>

          <UFormGroup v-if="formStatus === 'watched'" label="Rating">
            <div class="flex items-center gap-3">
              <UInput
                v-model.number="formRating"
                type="number"
                :min="1"
                :max="10"
                placeholder="1 – 10"
                class="w-28"
              />
              <span class="text-sm text-muted">out of 10</span>
            </div>
          </UFormGroup>

          <UFormGroup label="Review" hint="Optional">
            <UTextarea
              v-model="formReview"
              placeholder="Write a short review…"
              :rows="4"
              autoresize
            />
          </UFormGroup>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="editOpen = false">Cancel</UButton>
          <UButton :loading="saving" leading-icon="i-lucide-check" @click="submitForm">
            {{ listEntry ? 'Save changes' : 'Add to list' }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- Remove confirmation modal -->
    <UModal v-model:open="removeOpen" title="Remove from list" description="This will remove the title from your list. You can always add it back later.">
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="removeOpen = false">Cancel</UButton>
          <UButton color="error" :loading="removing" leading-icon="i-lucide-trash-2" @click="confirmRemove">
            Remove
          </UButton>
        </div>
      </template>
    </UModal>

  </div>
</template>
