<script setup lang="ts">
import type { ListEntry } from '~/composables/useList'

const props = defineProps<{
  externalId: string
  mediaType: 'movie' | 'series'
}>()

const { isAuthenticated } = useAuth()
const { addOrUpdate, remove, fetchStatus } = useList()

const listEntry = ref<ListEntry | null>(null)

const loadStatus = async () => {
  listEntry.value = await fetchStatus(props.externalId)
}

watch(
  isAuthenticated,
  (v) => {
    if (v) loadStatus()
  },
  { immediate: true }
)

const editOpen = ref(false)
const formStatus = ref<'watched' | 'plan_to_watch'>('plan_to_watch')
const formRating = ref<number | null>(null)
const formReview = ref('')
const saving = ref(false)
const hoverRating = ref(0)

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
    <div v-if="listEntry" class="flex flex-wrap items-center gap-3">
      <UBadge
        :color="listEntry.status === 'watched' ? 'success' : 'info'"
        size="lg"
      >
        {{ listEntry.status === 'watched' ? 'Watched' : 'Plan to watch' }}
      </UBadge>
      <span
        v-if="listEntry.rating"
        class="flex items-center gap-1.5 font-semibold text-sm"
      >
        <UIcon
          name="i-lucide-star"
          class="size-4 text-yellow-400 fill-yellow-400"
        />
        {{ listEntry.rating }}/10
      </span>
      <p
        v-if="listEntry.review"
        class="text-muted text-sm italic truncate max-w-xs"
      >
        "{{ listEntry.review }}"
      </p>
      <div class="flex gap-2 ml-auto">
        <UButton
          size="sm"
          variant="soft"
          color="neutral"
          leading-icon="i-lucide-pencil"
          @click="openEdit"
        >
          Edit
        </UButton>
        <UButton
          size="sm"
          variant="ghost"
          color="error"
          leading-icon="i-lucide-trash-2"
          @click="removeOpen = true"
        >
          Remove
        </UButton>
      </div>
    </div>

    <UButton v-else leading-icon="i-lucide-bookmark-plus" @click="openEdit">
      Add to list
    </UButton>

    <UModal
      v-model:open="editOpen"
      :title="listEntry ? 'Edit list entry' : 'Add to your list'"
      :description="
        listEntry
          ? 'Update your status, rating or review.'
          : 'Track this title in your personal list.'
      "
    >
      <template #body>
        <div class="flex flex-col gap-8 items-center text-center py-4">
          <div class="w-full max-w-sm space-y-3">
            <p class="text-sm font-medium text-default">Status</p>
            <div
              class="flex justify-center p-1 bg-muted rounded-xl border border-default"
            >
              <button
                v-for="s in [
                  {
                    label: 'Plan to Watch',
                    value: 'plan_to_watch',
                    icon: 'i-lucide-clock',
                  },
                  {
                    label: 'Watched',
                    value: 'watched',
                    icon: 'i-lucide-check-circle',
                  },
                ]"
                :key="s.value"
                class="flex-1 flex items-center justify-center gap-2 py-2 px-4 rounded-lg text-sm font-medium transition-all"
                :class="
                  formStatus === s.value
                    ? 'bg-primary text-white shadow-sm ring-1 ring-primary-600'
                    : 'text-muted hover:text-default hover:bg-default'
                "
                @click="formStatus = s.value as any"
              >
                <UIcon :name="s.icon" class="size-4" />
                {{ s.label }}
              </button>
            </div>
          </div>

          <div
            v-if="formStatus === 'watched'"
            class="w-full max-w-sm space-y-4 animate-in fade-in slide-in-from-top-2 duration-300"
          >
            <p class="text-sm font-medium text-default">Your Rating</p>
            <div class="flex flex-col items-center gap-4">
              <div class="flex items-center gap-2">
                <UInput
                  v-model.number="formRating"
                  type="number"
                  :min="1"
                  :max="10"
                  size="xl"
                  class="w-24 text-center [&_input]:text-center [&_input]:font-bold [&_input]:text-xl"
                  variant="outline"
                  color="primary"
                  placeholder="—"
                />
                <span class="text-xl font-bold text-muted">/ 10</span>
              </div>

              <div class="flex gap-1">
                <button
                  v-for="i in 10"
                  :key="i"
                  type="button"
                  class="group p-0.5 focus:outline-none"
                  @click="formRating = i"
                  @mouseenter="hoverRating = i"
                  @mouseleave="hoverRating = 0"
                >
                  <UIcon
                    :name="
                      i <= (hoverRating || formRating || 0)
                        ? 'i-lucide-star'
                        : 'i-lucide-star'
                    "
                    class="size-6 transition-all duration-200"
                    :class="[
                      i <= (hoverRating || formRating || 0)
                        ? 'text-yellow-400 fill-yellow-400 scale-110'
                        : 'text-zinc-300 dark:text-zinc-700 hover:scale-105',
                      i <= hoverRating && i > (formRating || 0)
                        ? 'opacity-50'
                        : '',
                    ]"
                  />
                </button>
              </div>
            </div>
          </div>

          <div class="w-full max-w-sm space-y-3">
            <div
              class="flex items-center justify-center gap-1.5 text-sm font-medium text-default"
            >
              <span>Review</span>
              <span class="text-xs text-muted font-normal">(optional)</span>
            </div>
            <UTextarea
              v-model="formReview"
              placeholder="What did you think? Share your thoughts..."
              :rows="4"
              autoresize
              class="w-full [&_textarea]:text-center [&_textarea]:resize-none"
              variant="outline"
            />
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="editOpen = false"
            >Cancel</UButton
          >
          <UButton
            :loading="saving"
            leading-icon="i-lucide-check"
            @click="submitForm"
          >
            {{ listEntry ? 'Save changes' : 'Add to list' }}
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="removeOpen"
      title="Remove from list"
      description="This will remove the title from your list. You can always add it back later."
    >
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="removeOpen = false"
            >Cancel</UButton
          >
          <UButton
            color="error"
            :loading="removing"
            leading-icon="i-lucide-trash-2"
            @click="confirmRemove"
          >
            Remove
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
