<script setup>
import {computed, onMounted, ref} from "vue"
import TabPageLayout from "../../../components/TabPageLayout.vue"
import UpcomingEvents from "../../../components/calendar/UpcomingEvents.vue"
import Calendar from "../../../components/calendar/Calendar.vue"
import {useI18n} from "vue-i18n";

const {t} = useI18n();
const navTabs = computed(() => {
  return [
    { value: 1, title: t('general.calendar.nearestEvents') },
    { value: 2, title: t('general.calendar.calendar') },
  ]
});

const calendarTab = ref(1)
const toggleCalendarTab = tab => {
  calendarTab.value = tab
}

const componets = {
  1: UpcomingEvents,
  2: Calendar,
}

onMounted(() => {
  const tg = window.Telegram.WebApp

  if (tg.platform === "android") {
    tg.disableVerticalSwipes()
  }
})
</script>

<template>
  <TabPageLayout>
    <div class="calendar">
      <div class="calendar_nav-bar nav-bar">
        <ul class="nav-bar_list">
          <template v-for="(tab, i) in navTabs" :key="i">
            <li :class="`nav-bar_item ${calendarTab === tab.value ? 'active' : ''}`" @click="toggleCalendarTab(tab.value)">
              <p class="nav-bar_text">{{ tab.title }}</p>
            </li>
          </template>
        </ul>
      </div>
      <div class="calendar_content">
        <component :is="componets[calendarTab]" />
      </div>
    </div>
  </TabPageLayout>
</template>

<style scoped lang="scss">
.calendar {

  .nav-bar {
    position: relative;
    z-index: 2;
    margin-bottom: 33px;
    border-radius: 8px;
    background: rgba(118, 118, 128, 0.2);

    &_list {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 2px;
    }

    &_item {
      cursor: pointer;
      display: flex;
      justify-content: center;
      width: 100%;

      &.active {
        border-radius: 7px;
        border: 0.5px solid rgba(0, 0, 0, 0.04);
        background: #454548;
      }
    }

    &_text {
      text-align: center;
      padding: 6px;
      width: 100%;
      color: var(--Label-Color-Dark-Primary, var(--system-white, #fff));
      font-size: 13px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }
  }

  &_content {
    position: relative;
    z-index: 2;
  }
}
</style>
