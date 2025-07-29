<script setup>
import {useSelectedLesson} from "../../store/selectedLesson.js";
import SVGDoubleArrowRight from "../svg/SVGDoubleArrowRight.vue";
import {useI18n} from "vue-i18n";

const {t} = useI18n();
const props = defineProps({
  component: {type: String, default: "", required: false},
})
const emits = defineEmits(['cta'])

const selectedLessonStore = useSelectedLesson()
const lessonStore = useSelectedLesson()

const toggleNewLessonTab = tab => {
  lessonStore.updateLessonActiveTab(tab)
}

const handleClick = () => {
  emits('cta');

  selectedLessonStore.toNextLesson();
  toggleNewLessonTab(1);
}
</script>

<template>
  <button
      @click="handleClick"
      :class="['btn_next-other', component]"
  >
    {{t('general.buttons.nextLesson')}}
    <SVGDoubleArrowRight :class="[{'white': component !== 'result'}]"/>
  </button>
</template>

<style scoped lang="scss">
@use "../../assets/styles/main.scss" as *;

.btn_next-other {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 48px;
  border-radius: 12px;
  width: 100%;
  background-color: rgba(255, 255, 255, 0.12);

  color: $white;
  font-size: 16px;
  font-style: normal;
  font-weight: 500;
  line-height: 130%;

  ::v-deep(.white) {
    & path {
      fill: $white;
      stroke: $white;
    }
  }
}

.result {
  background: #7581EC;
  color: #FFF;

  ::v-deep(svg) {
    & path {
      fill: #FFF;
    }
  }
}
</style>