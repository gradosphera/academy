<template>
  <div class="courses_course course">
    <div class="course_image">
      <AnimatedSkeleton height="86px"/>
    </div>
    <div class="course_body">
      <div class="course_type">
        <AnimatedSkeleton height="15px" width="80px"/>
      </div>
      <div class="course_title">
        <AnimatedSkeleton height="15px" width="100%" border-radius="12px"/>
      </div>
      <div class="course_bottom">
        <div class="course_item">
          <AnimatedSkeleton height="15px" width="80px" border-radius="12px"/>
        </div>
        <button class="course_explore">
          <AnimatedSkeleton height="15px" width="80px" border-radius="12px"/>
        </button>
      </div>
    </div>
  </div>
</template>
<script setup>
import AnimatedSkeleton from "../UI/AnimatedSkeleton.vue";
</script>
<style scoped lang="scss">
.course {
  position: relative;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.07);
  backdrop-filter: blur(12px);
  cursor: pointer;
  overflow: hidden;

  &_ban {
    position: absolute;
    z-index: 5;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;

    border-radius: 16px;
    opacity: 0.5;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(12px);
  }

  &_image {
    position: relative;
    height: 86px;
    background: rgba(255, 255, 255, 0.12);

    & img {
      height: 100%;
      width: 100%;
      object-fit: cover;
    }
  }

  &_product_date {
    position: absolute;
    z-index: 6;
    top: 8px;
    right: 10px;
    padding: 4px 8px;
    display: flex;
    gap: 4px;
    background-color: #FFF;
    border-radius: 10px;
    align-items: center;

    & span {
      color: #0A0A0A;

      font-size: 11px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    & svg {
      width: 12px;
      height: 12px;

      ::v-deep(path) {
        stroke: #0A0A0A;
      }
    }
  }

  &_body {
    padding: 16px;
  }

  &_type {
    margin-bottom: 8px;
    display: inline-block;
    border-radius: 10px;
    overflow: hidden;
    color: #a5a5a5;
    font-size: 11px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  &_title {
    margin-bottom: 30px;
    color: var(--White, var(--system-white, #fff));
    font-size: 15px;
    font-style: normal;
    font-weight: 600;
    line-height: 120%; /* 18px */
  }

  &_bottom {
    margin-top: 15px;
    display: flex;
  }

  &_item {
    display: flex;
    gap: 4px;

    color: #FFF;

    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;

    &.price {
      margin-right: 25px;
    }
  }

  &_icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.07);

    & svg {
      width: 70%;
    }
  }

  &_explore {
    margin-left: auto;
    background-color: transparent;
    color: rgba(255, 255, 255, 0.70);
    font-size: 13px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
    display: flex;
    align-items: center;
    gap: 4px;
  }
}
</style>