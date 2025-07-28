<template>
  <div class="author">
    <div>
      <Header />
      <div class="author_text">
        <span v-for="(line, index) in parseMessage(miniAppStore.miniAppData?.teacher_bio?.length ? miniAppStore.miniAppData?.teacher_bio : DEFAULT_DESCRIPTION)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(miniAppStore.miniAppData?.teacher_bio?.length ? miniAppStore.miniAppData?.teacher_bio : DEFAULT_DESCRIPTION).length - 1"/>
        </span>
      </div>
      <div class="author_socials">
        <a v-for="(link, i) in socialMedias.filter(item => item.link.length)" :key="i" :href="link.link" target="_blank" class="author_social">
          <component v-if="link.icon" :is="link.icon" />
        </a>
      </div>
    </div>
    <div v-if="miniAppStore.miniAppData?.tos && miniAppStore.miniAppData?.tos.length" class="author_navigation">
      <a :href="getAgreementLink(miniAppStore.miniAppData?.tos[0])" target="_blank"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="17" height="16" viewBox="0 0 17 16" fill="none">
          <path d="M6.33333 8L7.77778 9.5L10.6667 6.5M3 8.66667V4.27468C3 4.00711 3.15998 3.76546 3.4063 3.66096L8.23963 1.61046C8.40603 1.53986 8.59397 1.53986 8.76037 1.61046L13.5937 3.66096C13.84 3.76546 14 4.00711 14 4.27469V8.66667C14 11.7042 11.5376 14.1667 8.5 14.1667C5.46243 14.1667 3 11.7042 3 8.66667Z" stroke="white" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        {{t('general.author.terms_of_use')}}
      </a>
    </div>
    <div class="author_bottom">
      <button type="button" @click="toggleAuthorVisibility" class="author_hide-btn">
        <span>{{t('general.buttons.close')}}</span>
      </button>
    </div>
  </div>
</template>
<script setup>
import Header from "./Header.vue";
import SVGGlobe from "../components/svg/SVGGlobus.vue";
import SVGTelegram from "./svg/SVGTelegram.vue";
import SVGYouTube from "./svg/SVGYouTube.vue";
import SVGDiscord from "./svg/SVGDiscord.vue";
import SVGInstagram from "./svg/SVGInstagram.vue";
import { useAuthorStore } from "../store/authorStore.js";
import {useI18n} from "vue-i18n";
import {computed, markRaw, onMounted, ref} from "vue";
import {useMiniAppStore} from "../store/miniAppStore.js";
import SVGTikTok from "./svg/SVGTikTok.vue";
import SVGFacebook from "./svg/SVGFacebook.vue";
import SVGTwitter from "./svg/SVGTwitter.vue";
import SVGLink from "./svg/SVGLink.vue";
import {getAgreementLink, parseMessage} from "../helpers/index.js";

const { toggleAuthorVisibility } = useAuthorStore();
const miniAppStore = useMiniAppStore();
const {t} = useI18n();
const teachersData = computed(() => miniAppStore.miniAppData?.owner)
const DEFAULT_DESCRIPTION = computed(() => {
  return t('general.author.default_text');
})
const socialMedias = ref([
  {id: 0, icon: markRaw(SVGGlobe), link: ''},
  {id: 1, icon: markRaw(SVGInstagram), link: ''},
  {id: 2, icon: markRaw(SVGTikTok), link: ''},
  {id: 3, icon: markRaw(SVGTelegram), link: ''},
  {id: 4, icon: markRaw(SVGYouTube), link: ''},
  {id: 5, icon: markRaw(SVGDiscord), link: ''},
  {id: 6, icon: markRaw(SVGFacebook), link: ''},
  {id: 7, icon: markRaw(SVGTwitter), link: ''},
]);

const addNetwork = (link = '') => {
  const lastItemId = socialMedias.value[socialMedias.value.length - 1].id;
  const newNetwork = {id: lastItemId + 1, icon: markRaw(SVGLink), link};

  socialMedias.value.push(newNetwork);
}

const sortLinksFromStore = (data) => {

  function checkLink(address, value) {
    if (value?.link === address) return;

    if (value?.link?.length) {
      addNetwork(address)
    } else {
      value.link = address;
    }
  }

  if (data) {
    data.forEach((item, index) => {
      if (index === 0) {
        socialMedias.value[0].link = item;
      } else {
        switch (true) {
          case item.includes('https://www.instagram.com'):
            const instagram = socialMedias.value.find(item => item.id === 1);
            checkLink(item, instagram)
            break;
          case item.includes('https://www.tiktok.com'):
            const tiktok = socialMedias.value.find(item => item.id === 2);
            checkLink(item, tiktok)
            break;
          case item.includes('https://t.me'):
            const tg = socialMedias.value.find(item => item.id === 3);
            checkLink(item, tg)
            break;
          case item.includes('https://www.youtube.com'):
            const youtube = socialMedias.value.find(item => item.id === 4);
            checkLink(item, youtube)
            break;
          case item.includes('https://discord.com'):
            const discord = socialMedias.value.find(item => item.id === 5);
            checkLink(item, discord)
            break;
          case item.includes('https://www.facebook.com'):
            const facebook = socialMedias.value.find(item => item.id === 6);
            checkLink(item, facebook)
            break;
          case item.includes('https://x.com'):
            const twitter = socialMedias.value.find(item => item.id === 7);
            checkLink(item, twitter)
            break;
          default:
            addNetwork(item)
        }
      }
    })
  }
}

onMounted(() => {
  sortLinksFromStore(miniAppStore.miniAppData?.teacher_links)
})
</script>
<style scoped lang="scss">
.author {
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  color: var(--White, var(--system-white, #fff));

  &_text {
    margin-bottom: 24px;
    & span {
      font-size: 17px;
      font-style: normal;
      font-weight: 500;
      line-height: 125%;
    }
  }

  &_socials {
    display: flex;
    flex-wrap: wrap;
    gap: 32px;

    & svg {
      ::v-deep(path) {
        stroke-opacity: 1;
      }
    }
  }

  &_social {
    -webkit-tap-highlight-color: transparent;
    cursor: pointer;
    min-width: 48px;
    min-height: 48px;
    border-radius: 120px;
    background: rgba(255, 255, 255, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &_navigation {
    margin-top: auto;
    -webkit-tap-highlight-color: transparent;
    margin-bottom: 25px;
    display: flex;
    justify-content: center;

    a {
      display: flex;
      align-items: center;
      gap: 6px;
      cursor: pointer;
      color: #FFF;
      font-size: 13px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
      width: fit-content;
    }
  }

  &_hide-btn {
    -webkit-tap-highlight-color: transparent;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    height: 44px;
    width: 100%;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.15);

    span {
      font-size: 17px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
      color: var(--White, var(--system-white, #fff));
    }
  }
}
</style>
