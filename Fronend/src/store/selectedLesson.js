import { defineStore } from "pinia"
import {computed, ref} from "vue"

import { useSelectedCourse } from "./selectedCourse.js"

export const useSelectedLesson = defineStore("lesson", () => {
  const courseStore = useSelectedCourse()

  const lessonActiveTab = ref(1)
  const selectedLesson = ref(null)
  const allLessons = computed(() => {
    return courseStore.selectedCourse?.chapters?.flatMap(chapter => chapter.lessons);
  })

  const updateLessonActiveTab = value => {
    lessonActiveTab.value = value
  }

  const toNextLesson = () => {
    const currentIndex = allLessons.value.findIndex(lesson => lesson.lessonId === selectedLesson.value.lessonId)

    if (currentIndex >= 0 && currentIndex < allLessons.value.length - 1) {
      selectLesson(allLessons.value[currentIndex + 1])
    }
  }

  const toPreviousLesson = () => {
    const currentIndex = allLessons.value.findIndex(lesson => lesson.lessonId === selectedLesson.value.lessonId)

    if (currentIndex >= 0) {
      selectLesson(allLessons.value[currentIndex - 1])
    }
  }

  const selectLesson = lesson => {
    selectedLesson.value = lesson
  }

  return {
    selectedLesson,
    selectLesson,
    allLessons,
    lessonActiveTab,
    updateLessonActiveTab,
    toNextLesson,
    toPreviousLesson,
  }
})
