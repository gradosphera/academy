<template>
  <div class="homework">
    <div class="homework_description description">
      <p v-if="false" class="description_subtitle">Опис</p>
      <p class="description_text" v-html="replaceLinksInText(selectedLesson.language[locale]?.homeworkTask, selectedLesson.telegramLinks)"></p>
    </div>

    <div v-if="selectedLesson.homeTasksMaterials" class="homework_files">
      <div v-for="(elem, i) in selectedLesson.homeTasksMaterials" :key="i" class="">
        <FileContainer :item="elem" />
      </div>
    </div>

    <div v-if="savedReply && !selectedLesson.quiz" class="homework_reply reply">
      <div class="reply_head">
        <p class="reply_subtitle">Ваша відповідь</p>
        <div class="reply_options">
          <SVGHorizontalDots :style="{cursor: 'pointer'}" @click="isOptionsOpen = !isOptionsOpen"/>
          <Transition class="fade-in">
            <div ref="optionsRef" v-if="isOptionsOpen" :class="['reply_options-list']">
              <button @click="removeItem">Видалити <SVGTrash/></button>
              <button v-if="false">Відредагувати <SVGEdit/></button>
            </div>
          </Transition>
        </div>
      </div>
      <p class="reply_text">{{ savedReply.reply.text }}</p>
    </div>

    <div  class="homework_buttons">
      <button
          v-if="selectedLesson.language[locale].quiz"
          @click="startQuiz"
          :class="['homework_quiz-elem', {'homework_quiz-again': getQuizResults}]"
          :disabled="isQuizButtonDisabled"
      >
        <span v-if="getQuizResults"><span :class="`${isResultSuccess ? 'success' : 'failed'}`">({{getQuizResults.total_result}}%)</span> {{t('general.buttons.repeatQuiz')}}</span>
        <span v-else>{{t('general.buttons.start')}}</span>
      </button>
      <BtnNextLesson class="homework_next" v-if="getQuizResults || savedReply"/>
    </div>
    <Quiz :id="selectedLesson.lessonId" @on-close="isQuizOpen = false" :quiz="quiz" v-if="isQuizOpen"/>
  </div>
</template>
<script setup>
import {computed, ref} from "vue"
import {useSelectedLesson} from "../../store/selectedLesson.js"
import {onClickOutside} from "@vueuse/core"
import Quiz from "../Quiz/Quiz.vue";
import FileContainer from "./FileContainer.vue";
import SVGTrash from "../svg/SVGTrash.vue";
import SVGEdit from "../svg/SVGEdit.vue";
import SVGHorizontalDots from "../svg/SVGHorizontalDots.vue";
import BtnNextLesson from "./BtnNextLesson.vue";
import {useI18n} from "vue-i18n";
import {storeToRefs} from "pinia";
import {useSelectedCourse} from "../../store/selectedCourse.js";

const selectedLessonStore = useSelectedLesson();
const {selectedLesson} = storeToRefs(selectedLessonStore);
const courseStore = useSelectedCourse();
const isOptionsOpen = ref(false);
const optionsRef = ref(null);

const {t, locale} = useI18n();
const replyContent = ref("")
const isQuizOpen = ref(false);
const isFilesButtonVisible = ref(false);
const attachmentsRef = ref(null);

const quiz = computed(() => {
   return selectedLesson.value?.language[locale.value]?.quiz || [];
})

const handleOutsideClick = () => {
  if (input.value) {
    input.value.blur()
  }
}

const input = ref(null)
onClickOutside(input, () => handleOutsideClick())
onClickOutside(optionsRef, () => {isOptionsOpen.value = false})
onClickOutside(attachmentsRef, () => isFilesButtonVisible.value = false);

const startQuiz = () => {
  isQuizOpen.value = true;
}

const isQuizButtonDisabled = computed(() => {
  const {lessonType, lessonId} = selectedLesson.value;

  if (lessonType === 'stream') {
    const progress = courseStore.courseProgress["1"] || [];

    return !progress.find(val => val.lesson_id === lessonId);
  }

  return false;
})

const removeItem = () => {
  isOptionsOpen.value = false;
}

const replaceLinksInText = (text, links = []) => {
  let isTextIncludesLink = false;
  let link = '';

  if (links.length === 0) return text;

  for (let i = 0; i < links.length; i++) {
    if (text.includes(links[i])) {
      isTextIncludesLink = true;
      link = links[i];

      break;
    }
  }

  if (isTextIncludesLink) {
    return text.replace(link, `<a href="https://t.me/phanto_mr" target="_blank">${link}</a>`)
  }

  return text;
}

const isResultSuccess = computed(() => {
  return getQuizResults.value.total_result > 50;
})

</script>
<style scoped lang="scss">
@import "../../assets/styles/main";

.homework {
  height: auto;
  display: flex;
  flex-direction: column;
  padding-bottom: 35px;
  color: var(--White, var(--system-white, #fff));

  .description {
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    padding: 16px;

    &_subtitle {
      margin-bottom: 12px;
      opacity: 0.3;
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_text {
      white-space: pre-wrap;
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.16px;
      opacity: 0.8;

      ::v-deep(a) {
        color: #7581EC;
        cursor: pointer;
      }
    }
  }

  &_files {
    display: flex;
    flex-direction: column;
    margin: 16px 0 40px;
  }

  &_reply {
    margin-bottom: 39px;
  }

  &_buttons {
    width: 100%;
    position: fixed;
    bottom: 104px;
    left: 50%;
    transform: translateX(-50%);

    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &_quiz {

    &-elem {
      margin: 0 24px;
      height: 44px;
      border-radius: 12px;
      color: #FFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
      background-color: #7581EC;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;

      &:disabled {
        cursor: not-allowed;
        opacity: 0.6;
      }
    }

    &-again {
      color: $white;
      background: rgba(255, 255, 255, 0.12);

      & .success {
        color: #68E355;
      }

      & .failed {
        color: #EB4F57;
      }
    }
  }

  .reply {
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    padding: 16px;

    &_head {
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    &_subtitle {
      color: var(--White, var(--system-white, #fff));
      opacity: 0.3;
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_text {
      opacity: 0.8;
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      /* 20.8px */
      letter-spacing: 0.16px;
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
          line-height: 130%;
          letter-spacing: 0.16px;
        }
      }
    }
  }
}

.input {
  position: fixed;
  left: 0;
  bottom: 88px;
  width: 100%;
  height: 54px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  background: #0a0a0a;
  background: rgba(11, 11, 11, 0.35);
  backdrop-filter: blur(10px);

  &__popUp {
    position: absolute;
    z-index: 2;
    left: 20px;
    bottom: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  & .attachment {
    width: 195px;
    height: 50px;
    padding: 0 12px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #3B3B3C;
    cursor: pointer;

    color: #FFF;

    font-size: 16px;
    font-style: normal;
    font-weight: 400;
    line-height: 130%;
    letter-spacing: 0.16px;
  }

  &__add {
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 50%;
    min-width: 24px;
    min-height: 24px;
    max-width: 24px;
    max-height: 24px;
    background: rgba(255, 255, 255, 0.08);
    cursor: pointer;
  }

  input {
    width: 100%;
    border-radius: 9999px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    padding: 8px 12px;
    color: var(--White, var(--system-white, #fff));
    background: #0a0a0a;
    font-size: 15px;
    font-style: normal;
    font-weight: 400;
    line-height: normal;
    background: rgba(11, 11, 11, 0.35);
    backdrop-filter: blur(10px);
  }

  button {
    width: 24px;
    height: 24px;
    background: transparent;
  }
}

.slide-left-enter-active, .slide-left-leave-active {
  transition: transform .5s ease;
}

.slide-left-enter-from, .slide-left-leave-to {
  transform: translateX(-120%);
}

.slide-left-enter-to, .slide-left-leave-from {
  transform: translateX(0);
}
</style>
