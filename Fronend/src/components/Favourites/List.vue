<template>
  <div ref="favouritesListRef" class="favourites-list">
    <SkeletonFavorites v-if="isMediaLoading" v-for="favoriteLesson in list" :key="favoriteLesson.id" />
    <FavoriteItem v-show="!isMediaLoading" v-for="favoriteLesson in list" :key="favoriteLesson.id" :lesson="favoriteLesson" />
  </div>
</template>
<script setup>
import FavoriteItem from "./Item.vue"
import SkeletonFavorites from "../Skeletons/SkeletonFavorites.vue";
import {onMounted, ref} from "vue";
import {useMediaLoader} from "../../composable/useMediaLoader.js";

defineProps({
  list: {type: Array},
})
const favouritesListRef = ref(null);
const {isMediaLoading, waitForMediaLoad} = useMediaLoader();

onMounted(async() => {
  await waitForMediaLoad(favouritesListRef.value)
})
</script>
<style scoped lang="scss">
.favourites-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
</style>
