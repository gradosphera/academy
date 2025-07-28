<template>
  <div class="quiz">
    <div class="quiz_container">
      <div class="quiz_progress-bar">
        <div :style="{width: progressStatus + '%'}" class="quiz_progress"></div>
      </div>
      <div>
        <QuizForm
            :answers="chosenAnswers"
            @set-answer="setAnswer"
            :is-answer-checked="isAnswerChecked"
            :question="questionDetails"
        />
      </div>
      <div class="quiz_btns">
        <button v-if="!isAnswerChecked" :disabled="chosenAnswers.length === 0"
                @click="checkAnswer" class="quiz_btn"
        >
          {{t('general.buttons.checkQuizReply')}}
        </button>
        <button v-else @click="handleNextQuestion" :class="`quiz_btn ${isAnswerCorrect ? 'correct' : 'failed'}`">
          <SVGQuizCorrectAnswer v-if="isAnswerCorrect"/>
          <SVGWrongQuizAnswer v-else/>
          {{ `${buttonStatus} ${isLastQuestion ? t('general.buttons.finishQuiz') : t('general.buttons.nextQuizQuestion')}` }}
        </button>
      </div>
    </div>
    <Teleport to="body">
      <Transition name="slide-up">
        <div v-if="isResultOpen" ref="modal" class="quiz_modal">
          <div class="quiz_modal_container">
            <h3 class="quiz_modal_title">{{t('general.quiz.yourResult')}}</h3>
            <div class="quiz_modal_progress-wrapper">
              <div class="quiz_modal_progress">
                <RoundedProgressBar
                    :stroke-width="4"
                    :show-text="false"
                    class="quiz_modal_progress-circle"
                    :size="164"
                    :progress-color="'#106410'"
                    :background-color="circleColor"
                    :value="successAnswersInPercent"
                />
                <div class="quiz_modal_progress-content">
                  <span class="percents">{{ successAnswersInPercent }}%</span>
                </div>
              </div>
              <div class="quiz_modal_results">
                <div class="wrapper correct">
                  <div class="text"><SVGQuizCorrectAnswer/>{{t('general.quiz.correctAnswers')}}</div>
                  <div class="result"><span>{{ `${answersStatistic.correct}` }}</span> / {{quiz.length}}</div>
                </div>
                <div class="spacer"></div>
                <div class="wrapper wrong">
                  <div class="text"><SVGWrongQuizAnswer/>{{t('general.quiz.wrongAnswers')}}</div>
                  <div class="result"><span>{{ `${answersStatistic.failed}` }}</span> / {{quiz.length}}</div>
                </div>
              </div>
            </div>
            <div class="quiz_modal_btns">
              <BtnNextLessonOther @cta="emits('onClose')" component="result" />
              <button type="button" @click="emits('onClose')" class="quiz_modal_close">
                {{t('general.buttons.close')}}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
<script setup>
import {computed, onMounted, onUnmounted, ref} from "vue";
import QuizForm from "./QuizForm.vue";
import {useSelectedLesson} from "../../store/selectedLesson.js";
import RoundedProgressBar from "../RoundedProgressBar.vue";
import BtnNextLessonOther from "../SelectedLesson/BtnNextLessonOther.vue";
import {useSelectedCourse} from "../../store/selectedCourse.js";
import SVGQuizCorrectAnswer from "../svg/SVGQuizCorrectAnswer.vue";
import SVGWrongQuizAnswer from "../svg/SVGWrongQuizAnswer.vue";
import {useNavTabsStore} from "../../store/tabsStore.js";
import {useI18n} from "vue-i18n";

const props = defineProps({
  quiz: Array,
  id: Number,
})

const emits = defineEmits(['onClose']);
const lessonStore = useSelectedLesson();
const courseStore = useSelectedCourse();
const navBarStore = useNavTabsStore();
const {t, locale} = useI18n();

const currentQuestionId = ref(0);
const chosenAnswers = ref([]);
const isAnswerChecked = ref(false);
const isAnswerCorrect = ref(false);
const isResultOpen = ref(false);
const answersStatistic = ref({correct: 0, failed: 0});
const answersStatisticForLocaleStorage = ref([]);
const successAnswersInPercent = ref(0);

const questionDetails = computed(() => {
  return props.quiz.find(item => item.id === currentQuestionId.value)
})

const buttonStatus = computed(() => {
  if (isAnswerCorrect.value) {
    return t('general.buttons.correctQuizReply');
  }

  return t('general.buttons.wrongQuizReply');
})

const isLastQuestion = computed(() => {
  return currentQuestionId.value === props.quiz.length - 1;
});

const progressStatus = computed(() => {
  return ((currentQuestionId.value + 1) / props.quiz.length) * 100;
})

const checkAnswer = () => {
  const itemForLocaleStorage = {
    question_id: currentQuestionId.value,
    chosen_answers: chosenAnswers.value,
  };

  let isFailed = false;

  questionDetails.value.correctAnswer.forEach((item) => {
    if (!chosenAnswers.value.includes(item)) {
      isFailed = true;
    }
  })

  if (chosenAnswers.value.length > questionDetails.value.correctAnswer.length) {
    isFailed = true;
  }

  if (!isFailed) {
    answersStatistic.value.correct += 1;
    isAnswerCorrect.value = true;
  } else {
    answersStatistic.value.failed += 1;
  }

  answersStatisticForLocaleStorage.value.push(itemForLocaleStorage);
  isAnswerChecked.value = true;
};

const circleColor = computed(() => {
  if (successAnswersInPercent.value === 0) {
    return '#E53935';
  }

  return 'rgba(162, 236, 127, 0.12)';
})

const setAnswer = (id) => {
  if (questionDetails.value?.correctAnswer.length > 1) {
    if (chosenAnswers.value.includes(id)) {
      chosenAnswers.value = chosenAnswers.value.filter(item => item !== id);
    } else {
      chosenAnswers.value.push(id);
    }
  } else {
    if (chosenAnswers.value.length > 0) {
      chosenAnswers.value[0] = id;
    } else {
      chosenAnswers.value.push(id);
    }
  }
}

const calcTotalCorrectAnswers = () => {
  successAnswersInPercent.value = Math.round((answersStatistic.value.correct / props.quiz.length) * 100);
}

// const handleNextQuestion = () => {
//   if (isLastQuestion.value) {
//     calcTotalCorrectAnswers();
//     quizStore.updateQuiz({
//       lesson_id: props.id,
//       quiz: answersStatisticForLocaleStorage.value,
//       total_result: successAnswersInPercent.value
//     });
//     homeworkStore.saveReply({
//       lesson_id: props.id,
//       type: 'quiz',
//       reply: {
//         title: {en: lessonStore.selectedLesson.language.en.title, uk: lessonStore.selectedLesson.language.uk.title,},
//         text: `${successAnswersInPercent.value}%`
//       }
//     })
//     courseStore.checkAsOpened(lessonStore.selectedLesson.lessonId, {is_homework_done: true});
//     isResultOpen.value = true;
//
//     return;
//   }
//
//   currentQuestionId.value += 1;
//   isAnswerChecked.value = false;
//   chosenAnswers.value = [];
//   isAnswerCorrect.value = false;
// };

onMounted(() => {
  navBarStore.toggleNavBarVisibility();
})

onUnmounted(() => {
  navBarStore.toggleNavBarVisibility();
})
</script>

<style scoped lang="scss">
@import "../../assets/styles/main";

.quiz {
  width: 100%;
  height: 100vh;
  background-color: #0A0A0A;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 10000;

  &_container {
    height: 100%;
    position: relative;
    padding: 34px 24px 40px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  &_title {
    display: block;
    color: rgba(255, 255, 255, 0.5);

    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;

    margin-bottom: 16px;
  }

  &_progress {
    height: 100%;
    width: 50%;
    background: #7581EC;

    &-bar {
      width: 100%;
      position: absolute;
      top: 0;
      left: 0;
      height: 10px;
      background: linear-gradient(0deg, rgba(0, 0, 0, 0.20) 0%, rgba(0, 0, 0, 0.20) 100%), rgba(117, 129, 236, 0.50);
    }
  }

  &_btns {
    margin-top: 10px;
  }

  &_btn {
    height: 48px;
    border-radius: 12px;
    width: 100%;

    color: $white;
    font-size: 16px;
    font-style: normal;
    font-weight: 500;
    line-height: 130%;

    display: flex;
    align-items: center;
    justify-content: center;
    gap: 5px;

    background-color: rgba(255, 255, 255, 0.12);

    &:disabled {
      cursor: not-allowed;
      color: rgba(255, 255, 255, 0.30);
    }

    &.correct {
      color: #107910;
      background: rgba(3, 99, 3, 0.12);;
    }

    &.failed {
      color: #E53935;
      background: rgba(229, 57, 53, 0.12);
    }
  }

  &_modal {
    position: fixed;
    z-index: 10000;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 100vh;
    background-color: #0A0A0A;
    padding: 24px 0 0;

    &_container {
      border-radius: 32px 32px 0 0;
      height: 100%;
      padding: 24px 24px 40px;
      background-color: rgba(255, 255, 255, 0.04);
      display: flex;
      flex-direction: column;
      gap: 20px;
      justify-content: space-between;
    }

    &_progress {
      position: relative;

      &-content {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        display: flex;
        flex-direction: column;
        gap: 10px;
        align-items: center;

        & .percents {
          font-size: 40px;
          font-style: normal;
          font-weight: 400;
          line-height: 120%;
        }

        & span {
          color: $white;
          font-size: 16px;
          font-weight: 500;
          line-height: 130%;
        }
      }

      &-wrapper {
        margin: 0 auto;
        display: flex;
        flex-direction: column;
        gap: 32px;
        align-items: center;
      }

      &-bar {
        min-height: 152px;
        min-width: 152px;
        max-width: 152px;
        max-height: 152px;
        border-radius: 50%;
        display: flex;
        flex-direction: column;
        gap: 10px;
        justify-content: center;
        align-items: center;
        border: 4px solid rgba(162, 236, 127, 0.12);
      }
    }

    &_results {
      display: flex;
      flex-direction: column;
      background: rgba(255, 255, 255, 0.04);
      border-radius: 8px;
      padding: 16px;
      min-width: 232px;

      & .wrapper {
        display: flex;
        align-items: center;
        justify-content: space-between;
      }

      & .text {
        display: flex;
        align-items: center;
        gap: 8px;

        font-size: 13px;
        font-style: normal;
        font-weight: 400;
        line-height: normal;
      }

      & .correct {
        color: #106410;

        & .result {
          & span {
            color: #106410;
          }
        }
      }

      & .wrong {
        color: #E53935;

        & .result {
           & span {
             color: #E53935;
           }
        }
      }

      & .spacer {
        margin: 16px 0;
        width: 100%;
        height: 1px;
        background: rgba(255, 255, 255, 0.5);
      }

      & .result {
        color: rgba(255, 255, 255, 0.50);
        font-size: 13px;
        font-style: normal;
        font-weight: 400;
        line-height: normal;
        & span {
          color: rgba(255, 255, 255, 0.50);
          font-size: 13px;
          font-style: normal;
          font-weight: 400;
          line-height: normal;
        }
      }
    }

    &_title {
      color: $white;

      font-size: 22px;
      font-style: normal;
      font-weight: 700;
      line-height: 125%;
    }

    &_close {
      min-height: 46px;
      width: 100%;
      border-radius: 12px;

      color: $white;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
      background-color: rgba(255, 255, 255, 0.12);
    }

    &_btns {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
  }
}

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
</style>