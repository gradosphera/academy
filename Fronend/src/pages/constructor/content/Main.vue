<script setup>
import {onMounted} from "vue"
import TabPageLayout from "../../../components/TabPageLayout.vue"

import Courses from "../../../components/Courses.vue"
import {getMiniApp} from "../../../api/api.js";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {useProductsStore} from "../../../store/productsStore.js";

const miniAppStore = useMiniAppStore();
const productStore = useProductsStore();
onMounted(async() => {
  const miniAppResp = await getMiniApp();
  if (miniAppResp.data) {
    miniAppStore.setMiniAppData(miniAppResp.data.mini_app);
    miniAppStore.setTutorialData(miniAppResp.data.mini_app?.slides)
    productStore.setAllProducts(miniAppResp.data.mini_app.products);
    productStore.setProductAccess(miniAppResp.data.accesses);
  }
  const tg = window.Telegram.WebApp

  if (tg.platform === "android") {
    tg.disableVerticalSwipes()
  }
})
</script>
<template>
  <TabPageLayout>
    <Courses/>
  </TabPageLayout>
</template>
<style scoped lang="scss">
</style>
