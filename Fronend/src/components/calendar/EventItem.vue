<script setup>
import {useSelectedCourse} from "../../store/selectedCourse.js"
import {useSelectedLesson} from "../../store/selectedLesson.js"
import {useNavTabsStore} from "../../store/tabsStore.js"
import {useI18n} from "vue-i18n";

defineProps(["events"])

const {t, locale} = useI18n();
const courseStore = useSelectedCourse()
const lessonStore = useSelectedLesson()
const navTabsStore = useNavTabsStore()

const openLesson = lesson => {
  courseStore.selectCourse(coursesData.at(0)) // change when there will be few courses
  lessonStore.selectLesson(lesson)
  navTabsStore.updateActiveTab('selected_lesson')
}

function eventInProgress(start, end) {
  const startStream = start.getTime();
  const endStream = end.getTime();
  const now = Date.now();

  return startStream < now && endStream > now;
}

const openLink = (link) => {
  window.Telegram.WebApp.openLink(link);
}
</script>

<template>
  <div v-for="event in events" class="upcoming-events_item event">
    <p class="event_date">{{ event.language[locale]?.formatedDate }}</p>
    <p class="event_title">{{ event.language[locale]?.title }}</p>
    <div class="event_btns">
      <button v-if="eventInProgress(event.start, event.end)" type="button" class="event_btn-join" @click="openLink(event.link)">
        <span>{{t('general.buttons.joinStream')}}</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none">
          <path fill-rule="evenodd" clip-rule="evenodd" d="M17.5 9.99994C17.5 10.221 17.4122 10.4329 17.2559 10.5892L12.2559 15.5892C11.9305 15.9147 11.4028 15.9147 11.0774 15.5892C10.752 15.2638 10.752 14.7361 11.0774 14.4107L14.6548 10.8333H3.33333C2.8731 10.8333 2.5 10.4602 2.5 9.99994C2.5 9.53971 2.8731 9.16661 3.33333 9.16661H14.6548L11.0774 5.58922C10.752 5.26378 10.752 4.73614 11.0774 4.4107C11.4028 4.08527 11.9305 4.08527 12.2559 4.4107L17.2559 9.41069C17.4122 9.56697 17.5 9.77893 17.5 9.99994Z" fill="white"/>
        </svg>
      </button>
      <button v-else type="button" class="event_btn-learn-more" @click="openLesson(event)">
        <span>{{ t('general.buttons.openStreamLesson') }}</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none">
          <path fill-rule="evenodd" clip-rule="evenodd" d="M17.5 9.99994C17.5 10.221 17.4122 10.4329 17.2559 10.5892L12.2559 15.5892C11.9305 15.9147 11.4028 15.9147 11.0774 15.5892C10.752 15.2638 10.752 14.7361 11.0774 14.4107L14.6548 10.8333H3.33333C2.8731 10.8333 2.5 10.4602 2.5 9.99994C2.5 9.53971 2.8731 9.16661 3.33333 9.16661H14.6548L11.0774 5.58922C10.752 5.26378 10.752 4.73614 11.0774 4.4107C11.4028 4.08527 11.9305 4.08527 12.2559 4.4107L17.2559 9.41069C17.4122 9.56697 17.5 9.77893 17.5 9.99994Z" fill="white"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import "../../assets/styles/main";

.event {
  padding: 16px;
  border-radius: 16px;
  background: var(--7-white-transparent-bg, rgba(255, 255, 255, 0.07));

  &_date {
    margin-bottom: 8px;
    color: #919191;
    font-size: 13px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  &_title {
    margin-bottom: 24px;
    color: var(--White, var(--system-white, #fff));
    font-size: 17px;
    font-style: normal;
    font-weight: 500;
    line-height: 125%;
  }

  &_btns {
    display: flex;
    gap: 10px;

    & button {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 4px;
      padding: 0 12px;
      height: 44px;
      width: 100%;
      border-radius: 12px;
      background: #7581EC;
      color: #FFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%; /* 20.8px */
    }

    & .event_btn-learn-more {
      background: rgba(255, 255, 255, 0.07);
      color: $white;
    }
  }
}
</style>
