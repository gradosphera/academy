<template>
  <div class="edit-profile">
    <div class="edit-profile__inner">
      <span class="edit-profile__title">{{t('general.profile.menu.edit_profile.title')}}</span>

      <div @click="toggleAvatarPopUp" class="edit-profile__avatar_wrapper">
        <div class="edit-profile__avatar">
          <img v-if="singleInputs.avatar?.length" :src="avatarLink" alt="">
          <SVGProfileAvatar v-else/>
        </div>
        <span>{{t('general.profile.menu.edit_profile.set_profile_photo')}}</span>
      </div>

      <div class="edit-profile__input_wrapper">
        <span class="edit-profile__input_title">{{t('general.profile.menu.edit_profile.name')}}</span>
        <div>
          <div class="edit-profile__input name">
            <input @focus="hideSaveButtonOnMobile" @blur="showSaveButtonOnMobile" class="name" v-model="singleInputs.name" :placeholder="t('general.profile.menu.edit_profile.name_placeholder')" maxlength="30" type="text">
          </div>
          <UISymbols v-if="isSymbolsCounter" :max-length="30" :current-length="singleInputs.name.length" :is-border="false" />
        </div>
      </div>

      <div v-if="isSaveButtonVisible" class="edit-profile__bottom">
        <UIButton :height="'48px'" @click="clickSave" :bg="miniAppStore.accentedColor">{{t('general.buttons.save')}}</UIButton>
      </div>

      <div v-if="isAvatarPopUpOpen" class="edit-profile__popup_wrapper">
        <Transition name="slide-up">
          <div class="edit-profile__popup">
            <div class="edit-profile__popup_top">
              <h3 class="edit-profile__popup_title">{{t('general.profile.menu.edit_profile.set_profile_photo')}}</h3>
              <button @click="toggleAvatarPopUp">
                <SVGClose/>
              </button>
            </div>
            <div class="edit-profile__popup_content">
              <label for="avatarInput">
                {{t('general.profile.menu.edit_profile.gallery')}}
                <SVGGallery/>
                <input id="avatarInput" type="file" @change="uploadAvatarImage" accept="image/png, image/jpeg">
              </label>
              <button @click="removeAvatar">{{t('general.buttons.delete')}}
                <SVGDelete/>
              </button>
            </div>
          </div>
        </Transition>
      </div>
      <div class="edit-profile__cropper_wrapper" v-if="isCopperOpen">
        <div class="edit-profile__cropper_inner">
          <VuePictureCropper
              class="edit-profile__cropper"
              :boxStyle="{
                width: '100%',
                height: '100%',
                backgroundColor: '#D9D9D9',
              }"
              :img="singleInputs.temporary_avatar"
              :options="{
                viewMode: 1,
                aspectRatio: 1,
                dragMode: 'move',
                cropBoxMovable: true,
                cropBoxResizable: false,
              }"
              :presetMode="{
                mode: 'round',
                width: 342,
                height: 342,
                margin: 'auto'
              }"
          />
          <div class="edit-profile__cropper_btns">
            <button @click="closeCropper">{{t('general.buttons.cancel')}}</button>
            <button @click="uploadSelectedAvatar">{{t('general.buttons.select_photo')}}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import SVGDelete from "../svg/SVGDelete.vue";
import {computed, onMounted, ref, watch} from "vue";
import {useNavTabsStore} from "../../store/tabsStore.js";
import {useMiniAppStore} from "../../store/miniAppStore.js";
import SVGProfileAvatar from "../svg/SVGProfileAvatar.vue";
import SVGClose from "../svg/SVGClose.vue";
import SVGGallery from "../svg/SVGGallery.vue";
import {getFile} from "../../helpers/index.js";
import VuePictureCropper, {cropper} from 'vue-picture-cropper';
import {useToastStore} from "../../store/toastStore.js";
import {editUserProfile} from "../../api/api.js";
import UIButton from "../UI/UIButton.vue";
import {useI18n} from "vue-i18n";
import UISymbols from "../UI/UISymbols.vue";

const {t} = useI18n();
const toastStore = useToastStore();
const miniAppStore = useMiniAppStore();
const navBarStore = useNavTabsStore();
const isAvatarPopUpOpen = ref(false);
const isSaveButtonVisible = ref(true);
const isSymbolsCounter = ref(false);
const singleInputs = ref({
  name: '',
  temporary_avatar: '',
  avatar: '',
  file: null,
})


//cropper logic
const toggleAvatarPopUp = () => {
  isAvatarPopUpOpen.value = !isAvatarPopUpOpen.value;
  navBarStore.toggleNavBarVisibility();
}

const avatarLink = computed(() => {
  if (singleInputs.value.avatar.startsWith('data')) {
    return singleInputs.value.avatar;
  } else {
    return getFile(miniAppStore.userData?.avatar)
  }
})

const removeAvatar = async () => {
  miniAppStore.userData.delete_avatar = true;
  const formData = new FormData();

  formData.append('user', JSON.stringify(miniAppStore.userData));

  const resp = await editUserProfile(formData);
  if (resp.data) {
    miniAppStore.setUserData(resp.data.user)
  }

  delete miniAppStore.userData.delete_avatar;
  toggleAvatarPopUp();
}

const isCopperOpen = computed(() => {
  const result = navBarStore.activeState.at(-1) === 'profile-copper';

  if (!result) navBarStore.isNavBarVisible = true;

  return result;
})

const hideSaveButtonOnMobile = () => {
  isSymbolsCounter.value = true;

  if (!miniAppStore.isMobileDevice) return;

  isSaveButtonVisible.value = false;
}

const showSaveButtonOnMobile = () => {
  isSymbolsCounter.value = false;
  isSaveButtonVisible.value = true;
}

const uploadAvatarImage = (event) => {
  const file = event.target.files[0];
  if (file) {
    if (file.size > 3 * 1024 * 1024) {
      toastStore.error({text: t('general.toast_notifications.image_size_too_large')});
      return;
    }

    if (!['image/png', 'image/jpeg', 'image/jpg'].includes(file.type)) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      singleInputs.value.temporary_avatar = e.target.result;
      singleInputs.value.file = file;

      isAvatarPopUpOpen.value = !isAvatarPopUpOpen.value;
      navBarStore.setActiveState('profile-copper');
    };
    reader.readAsDataURL(file);
  }
}

const closeCropper = () => {
  singleInputs.value.file = null;
  singleInputs.value.temporary_avatar = '';
  navBarStore.setPreviousTab();
  navBarStore.toggleNavBarVisibility();
  if (!cropper) return;
  cropper.clear();
}

const uploadSelectedAvatar = async () => {
  if (!cropper) return
  const base64 = cropper.getDataURL()

  const file = await cropper.getFile();

  singleInputs.value.temporary_avatar = '';
  singleInputs.value.avatar = base64;
  singleInputs.value.file = file;
  isCopperOpen.value = false;
  navBarStore.setPreviousTab();
  navBarStore.toggleNavBarVisibility();
  cropper.clear();
}

const safeProfileChanges = async () => {

  miniAppStore.userData.first_name = singleInputs.value.name || '';

  const formData = new FormData();
  if (singleInputs.value.file) {
    formData.append("avatar", singleInputs.value.file);
  }
  formData.append("user", JSON.stringify(miniAppStore.userData));

  try {
    const resp = await editUserProfile(formData);

    if (resp.data) {
      miniAppStore.setUserData(resp.data.user)
    }

    toastStore.success({text: t('general.toast_notifications.changes_saved')})
  } catch (e) {
    toastStore.error({text: t('general.toast_notifications.something_went_wrong')})
  }
}

const clickSave = async () => {
  await safeProfileChanges();
}

watch(() => miniAppStore.userData, (newVal) => {
  if (newVal) {
    if (newVal.avatar) {
      singleInputs.value.avatar = newVal.avatar;
    } else {
      singleInputs.value.avatar = '';
    }
  }

})

onMounted(() => {
  if (miniAppStore.userData) {
    const data = miniAppStore.userData;
    singleInputs.value.name = data.first_name;
    singleInputs.value.avatar = data.avatar;
  }
})
</script>
<style scoped lang="scss">
@use "../../assets/styles/main.scss" as *;

.edit-profile {
  padding: 24px 24px 185px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  color: #fff;
  background: #0a0a0a;
  height: calc(100vh - 1px);

  &__title {
    color: rgba(255, 255, 255, 0.5);
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  &__avatar {
    min-height: 138px;
    max-height: 138px;
    aspect-ratio: 1/1;
    border-radius: 50%;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #212121;

    & img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    &_wrapper {
      margin: 24px 0;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-direction: column;
      gap: 18px;

      & span {
        font-size: 16px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }
    }
  }

  &__input {
    height: 50px;
    display: flex;
    align-items: center;
    padding: 0 16px;
    border-radius: 12px;
    background-color: #212121;

    &.name {
      height: 44px;
    }

    & input {
      width: 100%;
      margin-right: 15px;
      color: #FFF;
      font-size: 15px;
      font-style: normal;
      font-weight: 400;
      line-height: normal;
      background-color: transparent;

      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      display: inline-block;

      &::placeholder {
        color: rgba(255, 255, 255, 0.50);
        font-size: 15px;
        font-style: normal;
        font-weight: 400;
        line-height: normal;
      }
    }

    &_wrapper {
      display: flex;
      flex-direction: column;
      gap: 12px;

      & textarea {
        width: 100%;
        height: auto;
        background-color: #212121;
        border: none;
        outline: none;
        padding: 14px 16px;
        border-radius: 12px;

        color: #FFF;
        font-size: 15px;
        font-style: normal;
        font-weight: 400;
        line-height: normal;
        resize: none;
      }
    }

    &_title {
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
    }

    &_icon {
      min-width: 24px;
      max-width: 24px;
      min-height: 24px;
      max-height: 24px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(255, 255, 255, 0.10);
      margin-right: 8px;

      & svg {
        width: 50%;
        height: 50%;
      }
    }

    &_description {
      color: rgba(255, 255, 255, 0.50);
      font-size: 13px;
      font-style: normal;
      font-weight: 400;
      line-height: normal;
    }

    &_btns {
      display: flex;
      gap: 8px;
    }

    & .divider {
      height: 20px;
      width: 1px;
      background-color: rgba(255, 255, 255, 0.2);
    }

    & svg {
      cursor: pointer;
    }

    & .paste {
      background-color: transparent;
      color: #FFF;
      font-size: 13px;
      font-style: normal;
      font-weight: 400;
      line-height: normal;
    }
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &__bottom {
    position: fixed;
    z-index: 3;
    left: 0;
    bottom: 0;
    width: 100%;
    padding: 0 24px 100px;
    display: flex;
    gap: 8px;
  }

  &__author-preview {
    padding: 24px 24px 45px;
    position: fixed;
    z-index: 4;
    display: flex;
    top: 0;
    left: 0;
    min-height: 100vh;
    height: 100%;
    overflow: scroll;
    width: 100%;
    background: linear-gradient(
            180deg,
            #0e73e1 61.8%,
            #0e73e1 65.95%,
            #0e73e1 69.3%,
            #0e73e1 74.17%,
            #1276e1 75.99%,
            #177ae0 77.58%,
            #1b7ee0 79.11%,
            #2082df 80.7%,
            #2586df 82.52%,
            #2a8ade 84.7%,
            #2f8ede 87.39%,
            #3492dd 90.74%,
            #3996dd 94.89%,
            #3d9adc 100%
    );
  }

  &__popup {
    width: 100%;
    border-radius: 12px 12px 0 0;
    background: #181818;
    padding: 24px;

    &_wrapper {
      position: fixed;
      bottom: 0;
      width: 100%;
      left: 0;
      z-index: 10;
      min-height: 100vh;
      background: rgba(0, 0, 0, 0.40);
      backdrop-filter: blur(2px);
      display: flex;
      align-items: flex-end;
    }

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 24px;

      & button {
        background-color: transparent;
      }
    }

    &_title {
      font-size: 17px;
      font-style: normal;
      font-weight: 500;
      line-height: 125%;
    }

    &_content {
      display: flex;
      flex-direction: column;

      & label {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding-bottom: 12px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);

        font-size: 16px;
        font-style: normal;
        font-weight: 400;
        line-height: 130%;
        letter-spacing: 0.16px;
      }

      & input {
        position: absolute;
        opacity: 0;
        pointer-events: none;
      }

      & button {
        background: transparent;
        outline: none;
        border: none;
        padding: 12px 0;
        display: flex;
        align-items: center;
        justify-content: space-between;

        color: #FA4851;
        font-size: 16px;
        font-style: normal;
        font-weight: 400;
        line-height: 130%;
        letter-spacing: 0.16px;
      }
    }
  }

  &__cropper {
    ::v-deep(.cropper-bg) {
      background-image: none;
      background-color: #000;
    }

    &_wrapper {
      position: fixed;
      z-index: 5;
      left: 0;
      top: 0;
      height: 100vh;
      background: #0A0A0A;
      width: 100%;
    }

    &_inner {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100%;
      position: relative;

      & img {
        width: 100%;
      }
    }

    &_btns {
      position: absolute;
      width: 100%;
      left: 0;
      bottom: 50px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;

      & button {
        background: transparent;
        color: #FFF;
        font-size: 16px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
        letter-spacing: -0.3px;
      }
    }
  }

  &__modal {
    margin-top: 24px;
    display: flex;
    gap: 24px;
    flex-direction: column;

    &_text {
      color: rgba(255, 255, 255, 0.8);
      font-size: 17px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.17px;
    }

    &_btns {
      display: flex;
      gap: 8px;
    }
  }
}
</style>
