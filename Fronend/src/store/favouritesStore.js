import { defineStore } from "pinia"
import { ref } from "vue"
import {submitAnalyticsData} from "../helpers/index.js";

export const useFavoriteLessons = defineStore("favorites", () => {
  const favoriteLessonIds = ref(JSON.parse(localStorage.getItem("favoriteLessons")) || [])

  const toggleFavorite = lesson => {
    if (!lesson) return;

    if (!isFavorite(lesson.id)) {
      favoriteLessonIds.value.push(lesson.id)

      submitAnalyticsData('favorite_toggle', {
        lesson_title: lesson.title,
        action: 'add',
      })
    }
    else {
      favoriteLessonIds.value = favoriteLessonIds.value.filter(id => id !== lesson.id)

      submitAnalyticsData('favorite_toggle', {
        lesson_title: lesson.title,
        action: 'remove',
      })
    }

    saveToLocalStorage()
  }

  const isFavorite = lessonId => {
    return favoriteLessonIds.value.includes(lessonId)
  }

  const saveToLocalStorage = () => {
    localStorage.setItem("favoriteLessons", JSON.stringify(favoriteLessonIds.value))
  }

  return {
    favoriteLessonIds,

    toggleFavorite,
    isFavorite,
  }
})
