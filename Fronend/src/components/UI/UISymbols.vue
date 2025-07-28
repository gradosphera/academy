<template>
  <div class="symbols">{{`${currentLength}/${maxLength} ${t('general.main.symbols')}`}}</div>
</template>
<script setup>
import {useI18n} from "vue-i18n";
import {computed} from "vue";

const props = defineProps({
  currentLength: Number,
  maxLength: Number,
  isBorder: {type: Boolean, default: true},
})

const {t} = useI18n();

const borderStyle = computed(() => {
  if (props.isBorder) {
    return '1px solid #242426'
  }

  return 'none'
})
</script>
<style scoped lang="scss">
.symbols {
  width: 100%;
  color: rgba(255, 255, 255, 0.50);
  font-size: 12px;
  font-style: normal;
  font-weight: 400;
  line-height: normal;
  padding-top: 2px;
  margin-top: 2px;
  border-top: v-bind(borderStyle);
}
</style>