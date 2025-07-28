import { defineStore } from "pinia"
import { ref } from "vue"

export const useAuthorStore = defineStore("author", () => {
  const isShowAuthor = ref(false);

  const toggleAuthorVisibility = () => {
    isShowAuthor.value = !isShowAuthor.value
  }

  return {
    isShowAuthor,
    toggleAuthorVisibility,
  }
})
