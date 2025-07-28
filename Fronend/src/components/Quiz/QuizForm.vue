<script setup>
import SVGMultiAnswers from "../svg/SVGMultiAnswers.vue";
import {useI18n} from "vue-i18n";

const props = defineProps({
  question: Object,
  static: {type: Boolean, default: false, reflect: false},
  isAnswerChecked: {type: Boolean, default: false, reflect: false},
  answers: {type: Array, default: []},
})
const {t} = useI18n();
const emits = defineEmits(['setAnswer'])
const handleAnswerSelect = (answerId) => {
  if (props.static || props.isAnswerChecked) {
    return;
  }

  emits('setAnswer', answerId);
};

const getTitle = (answers) => {
  if (answers.length > 1) return t('general.quiz.chooseSeveralAnswer');

  return t('general.quiz.chooseOneAnswer');
}
</script>

<template>
  <div class="quiz_form">
    <span class="quiz_form_title">{{getTitle(question.correctAnswer)}}</span>
    <h3 class="quiz_form_question" v-html="`${question.id + 1}. ${question.question}`"></h3>
    <img class="quiz_form_image" v-if="question.image" :src="question.image" alt="Image">
    <ul class="quiz_form_list">
      <li
          v-for="item in question.answers"
          :key="item.id"
          :class="[
              'quiz_form_answer', {
            'selected': answers.includes(item.id),
            'correct': isAnswerChecked && answers.includes(item.id) && question.correctAnswer.includes(item.id),
            'failed': isAnswerChecked && answers.includes(item.id) && !question.correctAnswer.includes(item.id),
            'greenText': isAnswerChecked && question.correctAnswer.includes(item.id)
          }]"
          @click="handleAnswerSelect(item.id)"
      >
        <div v-if="question.correctAnswer.length < 2" class="quiz_form_circle"></div>
        <div v-else class="quiz_form_square">
          <SVGMultiAnswers class="icon" />
        </div>
        {{ item.text }}
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
@import "../../assets/styles/main";

.quiz {
  &_form {
    display: flex;
    flex-direction: column;
    gap: 24px;

    &_title {
      color: rgba(255, 255, 255, .5);
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_question {
      color: $white;
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.16px;
    }

    &_image {
      height: 190px;
      width: 340px;
      border-radius: 16px;
      overflow: hidden;
      align-self: center;

      & img {
        height: 100%;
        width: 100%;
        object-fit: cover;
      }
    }

    &_list {
      display: flex;
      flex-direction: column;
      gap: 24px;
    }

    &_answer {
      display: flex;
      gap: 16px;
      align-items: center;
      color: rgba(255, 255, 255, 0.70);

      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.16px;

      cursor: pointer;

      &.selected {
        color: $white;

        & .quiz_form {
          &_circle {
            border: 7px solid #7581EC;
          }

          &_square {
            background-color: #7581EC;
            border-color: #7581EC;
          }
        }
      }

      &.correct {
        color: #106410;

        & .quiz_form {
          &_circle {
            border: 7px solid #106410;
          }

          &_square {
            background-color: #106410;
            border-color: #106410;
          }
        }
      }

      &.greenText {
        color: #106410;
      }

      &.failed {
        color: #E53935;

        & .quiz_form {
          &_circle {
            border: 7px solid #E53935;
          }

          &_square {
            background-color: #E53935;
            border-color: #E53935;
          }
        }
      }
    }

    &_circle {
      min-width: 24px;
      min-height: 24px;
      border-radius: 50%;
      border: 2px solid rgba(255, 255, 255, 0.70);
    }

    &_square {
      min-width: 24px;
      min-height: 24px;
      border-radius: 6px;
      border: 1px solid rgba(255, 255, 255, 0.70);
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>