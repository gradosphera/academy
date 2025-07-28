<template>
  <div class="materials">
    <div class="materials_content">
      <div v-for="material in lessonMaterials" :key="material" class="materials_item">
        <FileContainer @open-modal="openModal" @set-link="setSourceLink" :item="material" />
      </div>
    </div>

    <Modal :is-open="isModalOpen" :position="'bottom'" >
      <template #bottom>
        <div class="materials__modal">
          <div class="materials__modal_top">
            <span>{{modalTitle}}</span>
            <button class="materials__modal_close" @click="closeModal">
              <SVGClose/>
            </button>
          </div>
          <iframe
              v-if="materialFormat === 'txt'"
              :allowfullscreen="true"
              :src="sourceLink"
          >
          </iframe>
          <UIPDFViewer v-if="materialFormat === 'pdf'" :source="sourceLink"/>
        </div>
      </template>
    </Modal>
  </div>
</template>
<script setup>
import FileContainer from "./FileContainer.vue";
import {useProductsStore} from "../../store/productsStore.js";
import Modal from "../UI/Modal.vue";
import {computed, ref} from "vue";
import UIPDFViewer from "../UI/UIPDFViewer.vue";
import {getFile} from "../../helpers/index.js";
import SVGClose from "../svg/SVGClose.vue";

const { lessonMaterials } = useProductsStore();

const isModalOpen = ref(false);
const sourceLink = ref('');
const selectedFile = ref(null);

const openModal = () => isModalOpen.value = true;
const closeModal = () => isModalOpen.value = false;
const setSourceLink = (file) => {
  if (file) {
    selectedFile.value = file;
    sourceLink.value = getFile(file.filename);
  }
}
const materialFormat = computed(() => {
  const materialLink = selectedFile.value?.filename?.split(".");

  if (materialLink) {
    return materialLink[materialLink.length - 1];
  }

  return '';
})
const modalTitle = computed(() => {
  if (!selectedFile.value) return "";

  return selectedFile.value?.title + '.' + materialFormat.value
})
</script>
<style scoped lang="scss">
.materials {
  height: 100%;
  color: var(--White, var(--system-white, #fff));
  display: flex;
  flex-direction: column;
  gap: 16px;
  justify-content: space-between;
  margin-bottom: 100px;

  &_content {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &_controls {
    width: 100%;
    position: fixed;
    bottom: 104px;
    left: 50%;
    transform: translateX(-50%);
  }

  &__modal {
    height: 100vh;
    display: flex;
    width: 100%;
    background: #1C1C1C;
    flex-direction: column;
    gap: 24px;
    padding: 24px 20px 40px;
    overflow: scroll;

    &_top {
      width: 100%;
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;

      color: #CECECE;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
    }

    &_close {
      position: absolute;
      top: 0;
      right: 0;
      background: transparent;
    }

    & iframe {
      border: none;
      border-radius: 12px;
      height: 100%;
      background: white;
    }
  }
}
</style>
