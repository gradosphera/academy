import { defineStore } from "pinia"
import { ref } from "vue"

export const useSelectedCourse = defineStore("course", () => {
  const selectedCourse = ref(null)
  const courseProgress = ref(JSON.parse(localStorage.getItem("courseProgress")) || {})

  const checkAsOpened = (lessonId, data=null) => {
    if (!selectedCourse.value) return // Ensure a course is selected first

    const courseId = selectedCourse.value.id

    if (!courseProgress.value[courseId]) {
      courseProgress.value[courseId] = [] // Initialize as an array if not already present
    }

    if (!courseProgress.value[courseId].find(lesson => lesson.lesson_id === lessonId)) {
      courseProgress.value[courseId].push({lesson_id: lessonId, ...data}) // Add the lesson if it's not already added
    } else {
      const lessonIndex = courseProgress.value[courseId].findIndex(lesson => lesson.lesson_id === lessonId);

      courseProgress.value[courseId][lessonIndex] = {lesson_id: lessonId, ...data};
    }
    saveToLocalStorage()
  }

  const saveToLocalStorage = () => {
    localStorage.setItem("courseProgress", JSON.stringify(courseProgress.value));
  }

  const getProgress = totalLessons => {
    if (!selectedCourse.value) return 0

    const courseId = selectedCourse.value.id
    const openedLessons = courseProgress.value?.[courseId] || []
    return (openedLessons.length / totalLessons) * 100
  }

  const selectCourse = course => {
    selectedCourse.value = course
  }

  return {
    selectedCourse,
    selectCourse,
    checkAsOpened,
    getProgress,
    courseProgress,
  }
})
