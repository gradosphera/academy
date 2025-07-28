import { ref } from 'vue'

export function useMediaLoader(delay = 300) {
    const isMediaLoading = ref(true)

    const waitForMediaLoad = async (containerRef) => {
        const container = containerRef
        if (!container) return

        const mediaElements = container.querySelectorAll('img, video')

        const loadPromises = Array.from(mediaElements).map(el => {
            if (el.complete) return Promise.resolve()
            return new Promise(resolve => {
                el.addEventListener('load', resolve)
                el.addEventListener('error', resolve)
            })
        })

        await Promise.all(loadPromises)

        setTimeout(() => {
            isMediaLoading.value = false
        }, delay)
    }

    return {
        isMediaLoading,
        waitForMediaLoad
    }
}
