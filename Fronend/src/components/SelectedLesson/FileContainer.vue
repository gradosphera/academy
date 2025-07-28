<template>
  <div class="file-block" @click="openLink">
    <div class="file-block_logo">
      <SVGLinkIcon v-if="item?.url?.length"/>
      <SVGPDFFileIcon v-else/>
    </div>
    <div class="file-block_texts">
      <p class="file-block_title">{{ item?.title }}</p>
      <p class="file-block_url">{{ item?.url?.length ? item?.url : formatFileSize(item?.size) }}</p>
    </div>
    <div v-if="item.type === 'link'" class="file-block_icon">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="21" viewBox="0 0 20 21" fill="none">
        <path
            d="M5.64704 13.7919C5.35414 14.0848 5.35414 14.5597 5.64704 14.8526C5.93993 15.1455 6.41481 15.1455 6.7077 14.8526L5.64704 13.7919ZM14.5729 6.67678C14.5729 6.26257 14.2371 5.92678 13.8229 5.92678L7.07285 5.92678C6.65864 5.92678 6.32285 6.26257 6.32285 6.67678C6.32285 7.09099 6.65864 7.42678 7.07285 7.42678H13.0729V13.4268C13.0729 13.841 13.4086 14.1768 13.8229 14.1768C14.2371 14.1768 14.5729 13.841 14.5729 13.4268L14.5729 6.67678ZM6.7077 14.8526L14.3532 7.20711L13.2925 6.14645L5.64704 13.7919L6.7077 14.8526Z"
            fill="white"
        />
      </svg>
    </div>
  </div>
</template>
<script setup>
import SVGLinkIcon from "../svg/SVGLinkIcon.vue";
import SVGPDFFileIcon from "../svg/SVGPDFFileIcon.vue";
import {formatFileSize, getFile, submitAnalyticsData} from "../../helpers/index.js";
import {computed} from "vue";
import {useProductsStore} from "../../store/productsStore.js";

const props = defineProps({
  item: Object,
})
const productStore = useProductsStore();

const emits = defineEmits(['openModal', 'setLink'])

const materialFormat = computed(() => {
  if(props.item?.filename?.length) {
    const materialLink = props.item?.filename?.split(".");

    if (materialLink) {
      return materialLink[materialLink.length - 1];
    }
  }
  return '';
})

const previewFormats = ['pdf', 'txt']

const openLink = () => {
  if (props.item?.url?.length) {

    submitAnalyticsData('material_open', {
      lesson_title: productStore.selectedLesson?.title,
      material_type: 'link',
      material_title: props.item?.title,
    })

    window.open(props.item.url, '_blank');
    return;
  }

  if(props.item?.filename?.length) {
    submitAnalyticsData('material_open', {
      lesson_title: productStore.selectedLesson?.title,
      material_type: 'file',
      material_title: props.item?.title,
    })

    if (previewFormats.includes(materialFormat.value)) {
      emits('setLink', props.item);
      emits('openModal');
    } else {
      window.open(getFile(props.item.filename), '_blank');
    }
  }
}
</script>
<style scoped lang="scss">
.file-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  gap: 16px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.04);
  overflow: hidden;
  cursor: pointer;

  &_logo {
    width: 40px;
    height: 40px;
  }

  &_texts {
    display: flex;
    gap: 4px;
    flex-direction: column;
    justify-content: space-between;
    flex-grow: 1;
    overflow: hidden;
  }

  &_title {
    color: var(--White, #fff);
    font-size: 15px;
    font-style: normal;
    font-weight: 600;
    line-height: 120%;
  }

  &_url {
    color: #919191;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    font-size: 13px;
    font-style: normal;
    font-weight: 400;
    line-height: normal;
  }

  &_icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;

    svg {
      min-width: 20px;
      min-height: 20px;
    }
  }
}
</style>