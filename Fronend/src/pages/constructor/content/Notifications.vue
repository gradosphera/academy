<script setup>
import { onMounted, onUnmounted } from "vue"
import { notifications } from "../../../constants/notifications.js"
import { useNavTabsStore } from "../../../store/tabsStore.js"
import BgBlurSquare from "../../../components/BgBlurSquare.vue"

const tg = window.Telegram.WebApp
const navTabsStore = useNavTabsStore()

function backToMainPage() {
  navTabsStore.updateActiveTab('main')
}

onMounted(() => {
  tg.BackButton.isVisible = true
  tg.BackButton.onClick(() => backToMainPage())

  if (tg.platform === "android") {
    tg.disableVerticalSwipes()
  }
})

onUnmounted(() => {
  tg.BackButton.isVisible = false
})
</script>

<template>
  <div class="notifications">
    <div class="notifications_container">
      <div class="notifications_head head">
        <p class="head_title">Повідомлення</p>
        <div class="head_icon">
          <img src="../../../assets/svg/settings-icon.svg" alt="" />
        </div>
      </div>
      <div class="notifications_list">
        <div v-for="notification in notifications" :key="notification.id" class="notifications_item notification">
          <div class="notification_top">
            <div v-if="!notification.isSeen" class="notification_seen-icon">
              <img src="../../../assets/svg/notifications.svg" alt="" />
            </div>
            <p class="notification_date">{{ notification.date }}</p>
          </div>
          <p class="notification_text">{{ notification.text }}</p>
        </div>
      </div>
    </div>
    <BgBlurSquare />
  </div>
</template>

<style scoped lang="scss">
.notifications {
  color: var(--White, var(--system-white, #fff));
  background: var(--Neutral-01, #0a0a0a);
  height: calc(100vh - 1px);
  overflow-y: auto;
  padding: 16px 24px 100px;

  &_container {
    position: relative;
    z-index: 2;
  }

  .head {
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;

    &_title {
      opacity: 0.5;
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_icon {
      display: flex;
      width: 32px;
      height: 32px;
      justify-content: center;
      align-items: center;
      border-radius: 9.6px;
      background: rgba(255, 255, 255, 0.07);
    }
  }

  &_list {
    display: flex;
    flex-direction: column;
    gap: 15px;

    .notification {
      display: flex;
      flex-direction: column;
      gap: 8px;
      padding: 16px;
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.04);
      cursor: pointer;

      &_top {
        display: flex;
        align-items: center;
        gap: 6px;
      }

      &_seen-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 8px;
        height: 8px;
      }

      &_date {
        color: #919191;
        font-size: 13px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }

      &_text {
        font-size: 17px;
        font-style: normal;
        font-weight: 500;
        line-height: 125%; /* 21.25px */
      }
    }
  }
}
</style>
