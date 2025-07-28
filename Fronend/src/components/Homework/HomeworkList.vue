<template>
  <div class="homework-replies">
    <div v-for="(reply, i) in []" :key="i" @click="openHomework(reply)" class="homework-replies_item reply">
      <div class="reply_head">
        <p class="reply_lesson-title">{{ reply.reply.title[locale] }}</p>
        <div class="reply_arrow-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
            <path
              d="M16 11.9967C16 11.7229 15.8997 11.4958 15.6858 11.2821L10.5377 6.25376C10.3639 6.08013 10.1566 6 9.90926 6C9.40783 6 9 6.39399 9 6.89482C9 7.1419 9.10697 7.36895 9.28749 7.54925L13.8606 11.99L9.28749 16.4441C9.10697 16.6244 9 16.8447 9 17.0985C9 17.5993 9.40783 18 9.90926 18C10.1566 18 10.3639 17.9132 10.5377 17.7396L15.6858 12.7112C15.9064 12.4975 16 12.2705 16 11.9967Z"
              fill="#EBEBF5"
              fill-opacity="0.3"
            />
          </svg>
        </div>
      </div>
      <p class="reply_text">{{ getReplyLessonText(reply) }}</p>
    </div>
  </div>
</template>
<script setup>
import { useSelectedHomework } from "../../store/selectedHomework.js"
import { useNavTabsStore } from "../../store/tabsStore.js"
import {useI18n} from "vue-i18n";

const {locale} = useI18n();
const navTabsStore = useNavTabsStore();
const selectedHomeworkStore = useSelectedHomework();

const getReplyLessonText = (reply) => {
  if (reply.type === 'quiz') {
    if (locale.value === 'ru') {
      return `Результаты теста: ${reply.reply.text} правильных ответов!`;
    }

    return `Результат тесту: ${reply.reply.text} правильних відповідей!`;
  }

  return reply.reply.text;
}

const openHomework = homework => {
  selectedHomeworkStore.selectHomework(homework)
  navTabsStore.updateActiveTab('selected_homework')
}
</script>
<style scoped lang="scss">
.homework-replies {
  display: flex;
  flex-direction: column;
  gap: 15px;

  .reply {
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.04);

    &_head {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    &_lesson-title {
      color: var(--White, var(--system-white, #fff));
      font-size: 17px;
      font-style: normal;
      font-weight: 500;
      line-height: 120%; /* 20.4px */
    }

    &_arrow-icon {
      display: flex;
      justify-content: center;
      align-items: center;
      width: 24px;
      height: 24px;
    }

    &_status-container {
      display: inline-block;
    }

    &_status {
      display: inline-flex;
      align-items: center;
      padding: 0 6px;
      gap: 4px;
      height: 24px;
      border-radius: 10px;
      background: rgba(145, 145, 145, 0.07);

      span {
        opacity: 0.7;
        color: var(--White, var(--system-white, #fff));
        font-size: 13px;
        font-style: normal;
        font-weight: 400;
        line-height: normal;
      }
    }

    &_text {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: wrap;
      display: -webkit-box;
      -webkit-line-clamp: 4; // Limit to 4 lines
      -webkit-box-orient: vertical;
      color: var(--70-white, rgba(255, 255, 255, 0.7));
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%; /* 20.8px */
      letter-spacing: 0.16px;
    }
  }
}
</style>
