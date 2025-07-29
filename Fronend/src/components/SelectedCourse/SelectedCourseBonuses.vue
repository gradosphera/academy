<template>
  <div class="course-bonuses">
    <div v-for="(bonus, i) in data" :key="i" class="course-bonuses__block">
      <span class="course-bonuses__block_title">{{bonus.title}}</span>
      <div class="course-bonuses__block_list">
        <div v-for="(item, index) in bonus.bonuses" :key="`bonus-${i}-${index}`" class="course-bonuses__block_item">
          <div class="inner">
            <SVGFileIcon v-if="item.filename?.length"/>
            <SVGLinkInField v-if="item.url?.length" />
            <div @click="openSource(item)" class="title">
              <span class="name">{{ item.title }}</span>
              <span>{{ getBonusText(item) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import {formatFileSize, getFile} from "../../helpers/index.js";
import SVGFileIcon from "../svg/SVGFileIcon.vue";
import SVGLinkInField from "../svg/SVGLinkInField.vue";

const props = defineProps({
  data: Array,
})

const getBonusText = (bonus) => {
  if (bonus?.url?.length) {
    return bonus?.url;
  } else if (bonus?.filename?.length) {
    return formatFileSize(bonus.size);
  }
}

const openSource = (item) => {
  if (item) {
    if (item.url?.length) {
      window.open(item.url, '_blank');

      return;
    }

    window.open(getFile(item.filename), '_blank');
  }
}
</script>
<style scoped lang="scss">
.course-bonuses {
  padding: 0 24px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  &__block {
    display: flex;
    flex-direction: column;
    gap: 16px;

    &_title {
      color: #FFF;
      font-size: 15px;
      font-style: normal;
      font-weight: 600;
      line-height: 120%;
    }

    &_list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    &_item {
      position: relative;
      height: 72px;
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.04);
      padding: 16px;

      & .inner {
        display: flex;
        gap: 16px;
        align-items: center;
        margin-right: 52px;

        & svg {
          min-width: 40px;
          min-height: 40px;
        }
      }

      & .name {
        color: #FFF;
        font-size: 15px;
        font-style: normal;
        font-weight: 600;
        line-height: 120%;
      }

      & .title {
        cursor: pointer;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;

        display: flex;
        flex-direction: column;

        & span:first-child {
          color: #FFF;
          font-size: 15px;
          font-style: normal;
          font-weight: 600;
          line-height: 120%;
        }

        & span {
          overflow: hidden;
          text-overflow: ellipsis;
          color: #919191;
          font-size: 13px;
          font-style: normal;
          font-weight: 400;
          line-height: normal;
        }
      }
    }
  }
}
</style>