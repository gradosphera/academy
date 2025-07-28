<script setup>
import EventItem from "./EventItem.vue"
import {computed} from "vue";

function getStreamLessons(courses) {
  return courses.flatMap(course => course.chapters.flatMap(chapter => chapter.lessons.filter(lesson => lesson.lessonType === "stream")))
}

const events = getStreamLessons(coursesData)

const filterEvents = computed(() => {
  const now = Date.now();

  return events.filter(lesson => {
    const startStream = lesson.start.getTime();
    const endStream = lesson.end.getTime();

    return (startStream < now && endStream > now) || startStream > now;
  })

})
</script>

<template>
  <div class="upcoming-events">
    <EventItem :events="filterEvents" />
  </div>
</template>

<style scoped lang="scss">
.upcoming-events {
  display: flex;
  flex-direction: column;
  gap: 15px;
}
</style>
