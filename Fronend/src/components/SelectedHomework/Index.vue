<template>
  <div class="selected-homework">
    <p class="selected-homework_lesson-title">{{ selectedHomework.reply.title[locale] }}</p>
    <div class="selected-homework_block">
      <div class="selected-homework_head">
        <p class="selected-homework_subtitle">{{t('general.homeworks.yourReply')}}</p>
        <div class="selected-homework_status-container">
          <div class="selected-homework_options">
            <SVGVerticalDots @click="isOptionsOpen = !isOptionsOpen"/>
            <Transition class="fade-in">
              <div ref="optionsRef" v-if="isOptionsOpen" :class="['selected-homework_options-list']">
                <button @click="removeItem">{{t('general.buttons.delete')}} <SVGTrash/></button>
                <button v-if="false">Відредагувати <SVGEdit/></button>
              </div>
            </Transition>
          </div>
        </div>
      </div>
      <p class="selected-homework_text">{{ getReplyLessonText(selectedHomework) }}</p>
    </div>
    <div v-if="getQuizFromLocale" class="selected-homework_quiz-list">
      <QuizForm
          v-for="(item, i) in getQuizFromLesson"
          :key="i"
          :question="item"
          :answers="getQuizFromLocale.quiz[i].chosen_answers"
          :is-answer-checked="true"
          :static="true"
      />
    </div>
  </div>
</template>
<script setup>
import {computed, onMounted, onUnmounted, ref} from "vue"
import { useNavTabsStore } from "../../store/tabsStore"
import { useSelectedHomework } from "../../store/selectedHomework"
import { storeToRefs } from "pinia"
import QuizForm from "../Quiz/QuizForm.vue";
import SVGVerticalDots from "../svg/SVGVerticalDots.vue";
import SVGTrash from "../svg/SVGTrash.vue";
import SVGEdit from "../svg/SVGEdit.vue";
import {onClickOutside} from "@vueuse/core"
import {useI18n} from "vue-i18n";

const tg = window.Telegram.WebApp
const navTabsStore = useNavTabsStore()
const quizStore = useQuizStore();
const isOptionsOpen = ref(false);
const selectedHomeworkStore = useSelectedHomework()
const { selectedHomework } = storeToRefs(selectedHomeworkStore)
const optionsRef = ref(null);
const {t, locale} = useI18n();

function backToHomeworkPage() {
  navTabsStore.updateActiveTab('homework')
}

onClickOutside(optionsRef, () => {isOptionsOpen.value = false})

const getQuizFromLesson = computed(() => {
  let quiz = null;
  if (getQuizFromLocale.value) {
    const chapter = telegramWalletCourse.chapters.find(chapter => {
      const findLesson = chapter.lessons.find(lesson => lesson.lessonId === +selectedHomework.value.lesson_id);
      if (findLesson) {
        quiz = findLesson.language[locale.value]?.quiz;
        return chapter;
      }
    })
  }

  return quiz;
})

const getReplyLessonText = (reply) => {
  if (reply.type === 'quiz') {
    if (locale.value === 'ru') {
      return `Результаты теста: ${reply.reply.text} правильных ответов!`;
    }

    return `Результат тесту: ${reply.reply.text} правильних відповідей!`;
  }

  return reply.reply.text;
}

const removeItem = () => {
  if (getQuizFromLocale.value) {
    quizStore.removeQuiz(getQuizFromLocale.value.lesson_id);
  }

  backToHomeworkPage();
}

onMounted(() => {
  tg.BackButton.isVisible = true
  tg.BackButton.onClick(() => backToHomeworkPage())

  if (tg.platform === "android") {
    tg.disableVerticalSwipes()
  }
})

onUnmounted(() => {
  tg.BackButton.isVisible = false
})
</script>
<style scoped lang="scss">
@use "../../assets/styles/main.scss" as *;

.selected-homework {
  height: calc(100vh - 1px);
  overflow-y: auto;
  overflow-x: hidden;
  position: relative;
  background: var(--Neutral-01, #0a0a0a);
  padding: 16px 24px 100px;
  z-index: 6;

  &_lesson-title {
    margin-bottom: 16px;
    color: var(--White, var(--system-white, #fff));
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
    opacity: 0.5;
  }

  &_block {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 10px 16px 16px;
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.04);
  }

  &_head {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &_subtitle {
    color: var(--White, var(--system-white, #fff));
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
    opacity: 0.3;
  }

  &_status-container {
    display: flex;
    align-items: center;
    gap: 5px;
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
      font-weight: 300;
      line-height: 120%; /* 15.6px */
    }
  }

  &_text {
    color: var(--70-white, rgba(255, 255, 255, 0.7));
    font-size: 16px;
    font-style: normal;
    font-weight: 400;
    line-height: 130%; /* 20.8px */
    letter-spacing: 0.16px;
  }

  &_quiz-list {
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 32px;
  }

  &_options {
    position: relative;
    display: flex;
    & svg {
      cursor: pointer;
    }

    &-list {
      position: absolute;
      right: 0;
      top: calc(100% + 12px);
      display: flex;
      flex-direction: column;
      gap: 8px;

      & button {
        height: 48px;
        width: 177px;
        border-radius: 16px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 12px;
        background-color: #3B3B3C;

        color: $white;

        font-size: 16px;
        font-weight: 400;
        letter-spacing: 0.16px;
      }
    }
  }
}

.fade-in-enter-active,
.fade-in-leave-active {
  transition: opacity 0.6s ease;
}

.fade-in-enter-from {
  opacity: 0;
}

.fade-in-enter-to {
  opacity: 1;
}

.fade-in-leave-from {
  opacity: 1;
}

.fade-in-leave-to {
  opacity: 0;
}
</style>
