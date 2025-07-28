<template>
  <TabPageLayout>
    <div class="favourites_content">
      <p class="favourites_title">{{t('general.favourites.title')}}</p>
      <template v-if="!isDataLoading">
        <FavouritesList v-if="favoriteLessons.length" :list="favoriteLessons"/>
        <FavouritesEmpty v-else />
      </template>
    </div>
  </TabPageLayout>
</template>
<script setup>
import {onMounted, ref, watch} from "vue"
import { storeToRefs } from "pinia"

import TabPageLayout from "../../../components/TabPageLayout.vue"
import FavouritesList from "../../../components/Favourites/List.vue"
import FavouritesEmpty from "../../../components/Favourites/Empty.vue"

import { useFavoriteLessons } from "../../../store/favouritesStore.js"
import {useI18n} from "vue-i18n";
import {getLesson} from "../../../api/api.js";

const {t} = useI18n();
const favouritesStore = useFavoriteLessons()
const { favoriteLessonIds } = storeToRefs(favouritesStore);
const favoriteLessons = ref([]);
const isDataLoading = ref(true);

const getFavouriteLessons = async() => {
  try {
    for (const id of favoriteLessonIds.value) {
      const resp = await getLesson(id);

      if (resp.data) {
        favoriteLessons.value.push(resp.data.lesson);
      }
    }
  } finally {
    isDataLoading.value = false;
  }
}

watch(() => favoriteLessonIds.value, (newVal) => {
  favoriteLessons.value = favoriteLessons.value.filter(item => newVal.includes(item.id));
});

onMounted(async() => {
  const tg = window.Telegram.WebApp;

  if (tg.platform === "android") {
    tg.disableVerticalSwipes()
  }

  await getFavouriteLessons();
})
</script>
<style scoped lang="scss">
.favourites {
  &_title {
    margin-bottom: 18px;
    color: var(--White, var(--system-white, #fff));
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
    opacity: 0.5;
  }

  &_content {
    height: 100%;
    position: relative;
    z-index: 2;
  }
}
</style>
