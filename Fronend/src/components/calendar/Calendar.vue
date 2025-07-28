<template>
  <div class="calendar">
    <vue-cal
        :locale="locale"
        hide-view-selector
        :time="false"
        small
        active-view="month"
        :disable-views="['years', 'year', 'week', 'day']"
        @cell-click="handleDayClick"
        :eventsOnMonthView="true"
        :events="events"
    >
      <template #arrow-prev>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path
              d="M7.83566 11.6448C7.83566 11.3642 7.9383 11.1316 8.15727 10.9126L13.4262 5.76002C13.6041 5.58211 13.8162 5.5 14.0694 5.5C14.5826 5.5 15 5.90372 15 6.41693C15 6.67011 14.8905 6.90276 14.7058 7.08751L10.0253 11.6379L14.7058 16.202C14.8905 16.3868 15 16.6126 15 16.8726C15 17.3858 14.5826 17.7964 14.0694 17.7964C13.8162 17.7964 13.6041 17.7074 13.4262 17.5295L8.15727 12.3769C7.93146 12.158 7.83566 11.9253 7.83566 11.6448Z"
              fill="white"
          />
        </svg>
      </template>
      <template #arrow-next>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path
              d="M16.1643 11.6448C16.1643 11.3642 16.0617 11.1316 15.8427 10.9126L10.5738 5.76002C10.3959 5.58211 10.1838 5.5 9.93061 5.5C9.41741 5.5 9 5.90372 9 6.41693C9 6.67011 9.10948 6.90276 9.29424 7.08751L13.9747 11.6379L9.29424 16.202C9.10948 16.3868 9 16.6126 9 16.8726C9 17.3858 9.41741 17.7964 9.93061 17.7964C10.1838 17.7964 10.3959 17.7074 10.5738 17.5295L15.8427 12.3769C16.0685 12.158 16.1643 11.9253 16.1643 11.6448Z"
              fill="white"
          />
        </svg>
      </template>
    </vue-cal>
    <Teleport to="body">
      <Transition name="slide-up">
        <div v-if="dayEvents.length" ref="modal" class="modal-overlay">
          <div class="modal-events">
            <EventItem :events="dayEvents"/>
          </div>
          <div class="modal-close-btn-container">
            <button type="button" @click="dayEvents = []" class="modal-close-btn">
              <span>{{t('general.buttons.close')}}</span>
            </button>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
<script setup>
import {ref} from "vue"
import VueCal from "vue-cal"
import "vue-cal/dist/vuecal.css"

import {isSameDay} from "../../helpers/dates"
import EventItem from "./EventItem.vue"
import {onClickOutside} from "@vueuse/core"
import {useI18n} from "vue-i18n";

const {t, locale} = useI18n();
function getStreamLessons(courses) {
  return courses.flatMap(course => course.chapters.flatMap(chapter => chapter.lessons.filter(lesson => lesson.lessonType === "stream")))
}

const events = getStreamLessons(coursesData)

const dayEvents = ref([])
const handleDayClick = date => (dayEvents.value = events.filter(event => isSameDay(event.start, date))) // maybe change later, if there will be range eventd

const modal = ref(null)

onClickOutside(modal, () => {
  dayEvents.value = []
})
</script>
<style lang="scss">
.calendar {
  width: 100%;
}

.modal-overlay {
  position: fixed;
  z-index: 10000;
  bottom: 0;
  left: 0;
  width: 100%;
  max-height: 88%;
  display: flex;
  flex-direction: column;
  border-radius: 32px 32px 0 0;
  background: #202022;
  box-shadow: 0 3px 30px 0 rgba(0, 0, 0, 0.16);

  .modal-events {
    height: fit-content;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    overflow-y: auto;
  }

  .modal-close-btn-container {
    border-top: 1px solid #242426;
    padding: 18px 24px;

    .modal-close-btn {
      width: 100%;
      display: flex;
      height: 48px;
      justify-content: center;
      align-items: center;
      gap: 8px;
      border-radius: 12px;
      background: var(--12-white, rgba(255, 255, 255, 0.12));

      span {
        color: var(--White, var(--system-white, #fff));
        font-size: 16px;
        font-style: normal;
        font-weight: 500;
        line-height: 130%; /* 20.8px */
      }
    }
  }
}

// Popup Animation
//////////////////
//////////////////

.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.6s ease;
}

.slide-up-enter-from {
  transform: translateY(100%);
}

.slide-up-enter-to {
  transform: translateY(0);
}

.slide-up-leave-from {
  transform: translateY(0);
}

.slide-up-leave-to {
  transform: translateY(100%);
}

// Calendar Customization
/////////////////////////
/////////////////////////

.vuecal {
  color: var(--White, var(--system-white, #fff));

  &__weekdays-headings .weekday-label .full {
    display: none;
  }

  &__weekdays-headings .weekday-label .xsmall {
    display: none;
  }

  &__weekdays-headings .weekday-label .small {
    display: inline;
  }

  &__cell:before {
    border: none;
  }

  &__flex {
    row-gap: 24px;
  }

  &__arrow--prev {
    margin-left: 0;
    display: flex;
    min-width: 34px;
    width: 34px;
    min-height: 34px;
    height: 34px;
    justify-content: center;
    align-items: center;
    border-radius: 12px;
    background: var(--12-white, rgba(255, 255, 255, 0.12));
  }

  &__arrow--next {
    margin-right: 0;
    display: flex;
    min-width: 34px;
    width: 34px;
    min-height: 34px;
    height: 34px;
    justify-content: center;
    align-items: center;
    border-radius: 12px;
    background: var(--12-white, rgba(255, 255, 255, 0.12));
  }

  &__weekdays-headings {
    border: none;
  }

  &__cell {
    height: 30px;
  }

  &__cell-content {
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    max-width: 30px;
    width: 30px;
    height: 30px;
    border-radius: 8px;
  }

  &__cell--today,
  &__cell--selected {
    background-color: inherit;

    .vuecal__cell-content {
      background: rgba(255, 255, 255, 0.5);
    }
  }

  &__cell--has-events {
    .vuecal__cell-content {
      background: rgba(255, 255, 255, 0.12);
    }
  }

  &__title-bar {
    margin-bottom: 20px;
  }

  &__title {
    font-size: 22px;
    font-style: normal;
    font-weight: 500;
    line-height: 120%; /* 26.4px */
  }

  &__heading {
    height: 100%;
  }

  &__cell-date {
    font-size: 15px;
    font-style: normal;
    font-weight: 400;
    line-height: normal;
  }

  &__cell-events {
    display: none;
  }

  &__cell--out-of-scope {
    visibility: hidden;
  }

  .weekday-label {
    color: var(--White, var(--system-white, #fff));
    opacity: 0.4;
    font-size: 13px;
    font-style: normal;
    font-weight: 400;
    line-height: normal;
  }
}

.vuecal__cell--today {
  & .vuecal__cell-content {
    background: rgba(255, 255, 255, 0.12);
  }
}

.vuecal__cell--has-events {
  background-color: inherit;

  .vuecal__cell-content {
    background: rgba(117, 129, 236, 0.30);
  }
}

.vuecal__cell--selected.vuecal__cell--has-events {
  background-color: inherit;

  .vuecal__cell-date {
    color: #FFF;
  }

  .vuecal__cell-content {
    background: linear-gradient(180deg, #464D88 0%, #7581EC 100%);
  }
}
</style>
