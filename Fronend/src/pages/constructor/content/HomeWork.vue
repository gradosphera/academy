<script setup>
import { onMounted } from "vue"
import TabPageLayout from "../../../components/TabPageLayout.vue"
import HomeworkList from "../../../components/Homework/HomeworkList.vue"
import HomeworkEmpty from "../../../components/Homework/Empty.vue"

const { repliesMap } = useHomeworksStore();

onMounted(() => {
  const tg = window.Telegram.WebApp

  if (tg.platform === "android") {
    tg.disableVerticalSwipes()
  }
})
</script>

<template>
  <TabPageLayout>
    <div class="homework_content">
      <p class="homework_title">Домашка</p>
      <HomeworkList v-if="repliesMap.length > 0" />
      <HomeworkEmpty v-else />
    </div>
  </TabPageLayout>
</template>

<style scoped lang="scss">
@import "../../../assets/styles/main";

.homework {
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
