<script setup>
import { storeToRefs } from "pinia"
import PodcastType from "./PodcastType.vue"
import VideoType from "./VideoType.vue"
import TextType from "./TextType.vue"
import StreamType from "./StreamType.vue"
import {useProductsStore} from "../../../store/productsStore.js";

const productStore = useProductsStore();
const { selectedLesson } = storeToRefs(productStore);
</script>

<template>
  <div v-if="selectedLesson" class="content" :key="selectedLesson?.id">
    <VideoType v-if="selectedLesson.materials?.[0]?.content_type === 'circle_video' || selectedLesson.materials?.[0]?.content_type === 'video'"/>
    <PodcastType v-if="selectedLesson.materials?.[0]?.content_type === 'audio'" />
    <TextType v-if="selectedLesson.materials?.[0]?.content_type === 'text'" />
    <StreamType v-if="selectedLesson.materials?.[0]?.content_type === 'stream'"  />
  </div>
</template>

<style scoped lang="scss">
.content {
  height: 100%;
}
</style>
