<template>
  <div class="favorite-item" @click="moveToLesson">
    <div class="favorite-item_image">
      <img :src="getLessonCover()" alt="" />
      <button type="button" @click.stop="favouritesStore.toggleFavorite(lesson?.id)" class="favorite-item_heart-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 18 18" fill="none">
          <path
            d="M16.875 7.17188C16.875 12.0938 9.57726 16.0777 9.26648 16.2422C9.18457 16.2863 9.09301 16.3093 9 16.3093C8.90699 16.3093 8.81543 16.2863 8.73352 16.2422C8.42273 16.0777 1.125 12.0938 1.125 7.17188C1.1263 6.0161 1.58601 4.90803 2.40327 4.09077C3.22053 3.27351 4.3286 2.8138 5.48438 2.8125C6.93633 2.8125 8.20758 3.43687 9 4.49227C9.79242 3.43687 11.0637 2.8125 12.5156 2.8125C13.6714 2.8138 14.7795 3.27351 15.5967 4.09077C16.414 4.90803 16.8737 6.0161 16.875 7.17188Z"
            fill="white"
          />
        </svg>
      </button>
    </div>
    <div class="favorite-item_bottom">
      <p class="favorite-item_title">{{ lesson?.title }}</p>
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
        <path
          d="M15.1643 11.6448C15.1643 11.3642 15.0617 11.1316 14.8427 10.9126L9.57383 5.76002C9.39592 5.58211 9.18379 5.5 8.93061 5.5C8.41741 5.5 8 5.90372 8 6.41693C8 6.67011 8.10948 6.90276 8.29424 7.08751L12.9747 11.6379L8.29424 16.202C8.10948 16.3868 8 16.6126 8 16.8726C8 17.3858 8.41741 17.7964 8.93061 17.7964C9.18379 17.7964 9.39592 17.7074 9.57383 17.5295L14.8427 12.3769C15.0685 12.158 15.1643 11.9253 15.1643 11.6448Z"
          fill="#EBEBF5"
          fill-opacity="0.3"
        />
      </svg>
    </div>
  </div>
</template>
<script setup>
import { useNavTabsStore } from "../../store/tabsStore.js"
import { useFavoriteLessons } from "../../store/favouritesStore.js"
import {getFile} from "../../helpers/index.js";
import {getLesson} from "../../api/api.js";
import {useProductsStore} from "../../store/productsStore.js";

const navTabsStore = useNavTabsStore();
const favouritesStore = useFavoriteLessons();
const props = defineProps(["lesson"]);
const productStore = useProductsStore();

const getLessonCover = () => {
  if (props.lesson?.materials) {
    const coverItem = props.lesson?.materials.find(item => item.category === 'lesson_cover');

    return getFile(coverItem?.filename);
  }
}

async function moveToLesson() {
  const resp = await getLesson(props.lesson?.id);

  if (resp.data) {
    productStore.setSelectedLesson(resp.data.lesson);
  }
  navTabsStore.updateActiveTab('selected_lesson');
}
</script>
<style scoped lang="scss">
.favorite-item {
  cursor: pointer;
  color: white;
  width: 100%;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  padding: 16px;
  flex-direction: column;
  gap: 12px;

  &_image {
    position: relative;

    img {
      border-radius: 8px;
      max-width: 100%;
    }
  }

  &_heart-icon {
    display: flex;
    width: 32px;
    height: 32px;
    padding: 7px;
    justify-content: center;
    align-items: center;
    border-radius: 525418.438px;
    background: rgba(182, 182, 182, 0.15);
    backdrop-filter: blur(10.508474349975586px);
    position: absolute;
    top: 10px;
    right: 10px;
    z-index: 10;
  }

  &_bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}
</style>
