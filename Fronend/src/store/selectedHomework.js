import { defineStore } from "pinia"
import { ref } from "vue"

export const useSelectedHomework = defineStore("homework", () => {
  const selectedHomework = ref(null)

  const selectHomework = homework => {
    selectedHomework.value = homework
  }

  return {
    selectedHomework,
    selectHomework,
  }
})
