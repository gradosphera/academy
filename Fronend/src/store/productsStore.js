import {defineStore} from "pinia";
import {computed, ref, watch} from "vue";

export const useProductsStore = defineStore("products", () => {
    //products data
    const allProducts = ref(null);
    const selectedProduct = ref(null);
    const selectedProductModules = ref(null);
    const productProgress = ref([]);
    const productReviews = ref([]);
    const productsAccess = ref([]);
    const productInvite = ref(null);
    const selectedPaymentTariff = ref(null);

    //chapters data
    const selectedChapter = ref(null);

    //lessons data
    const paidLessons = ref([]);
    const selectedLesson = ref(null);
    const selectedContentType = ref(null); //format 'text' || 'audio' || 'video'
    const selectedLessonTab = ref(null);
    const lessonMaterials = ref([]);
    const lessonMaterialsFromBackend = ref({content: [], materials: []});
    const lessonContentData = ref({
        video_material_id: '',
        title: '',
        description: '',
        cover: '',
        cover_file: null,
        cover_name: '',
        cover_size: 0,
        filename: '',
        metadata: null,
        file: null,
        circle_filename: '',
        circle_metadata: null,
        circle_file: null,
        youtube: '',
    });
    const lessonHomeworkType = ref(null); //format {id, label: 'quiz' || 'task'}
    const lessonHomeworkTask = ref({description: '', files: [], links: []});
    const lessonHomeworkData = ref({
        description: '',
        task: {files: [], links: [], to_delete: []},
        quiz: []
    });
    const selectedVideoType = ref('');
    const isMediaLoading = ref(false);
    const isLessonExitModalOpen = ref(false);
    const lessonTabs = ref([]);
    const isLessonCompleted = ref(false);
    const isLessonReviewed = ref(false);


    //products logic
    const setAllProducts = (data) => allProducts.value = data;
    const setSelectedProduct = (val) => selectedProduct.value = val;
    const setProductProgress = (el) => productProgress.value = el;
    const setProductReviews = (el) => productReviews.value = el;
    const setProductAccess = (el) => productsAccess.value = el;
    const setProductInvite = (el) => productInvite.value = el;
    const setSelectedPaymentTariff = (el) => selectedPaymentTariff.value = el;
    const removeSelectedProduct = () => {
        selectedProduct.value = null;
        selectedLesson.value = null;
        selectedChapter.value = null;

        if (selectedContentType.value) {
            selectedContentType.value = null;
        }
    }
    const availableProducts = computed(() => {
        const arr = [];

        if (productsAccess.value?.length > 0) {
            allProducts.value.forEach(item => {
                const findProduct = productsAccess.value.find(el => el.product_id === item.id);

                if (findProduct) {
                    arr.push({...item, deleted_at: findProduct.deleted_at});
                } else {
                    arr.push(item);
                }
            })

            return arr.filter(product => product.is_active);
        } else {
            return allProducts.value?.filter(product => product.is_active);
        }
    })


    watch(() => selectedProduct.value, () => {
        if (!selectedProduct.value?.lessons) {
            selectedProductModules.value = [{
                name: '',
                items: [],
                other_keys: {is_edit_options_open: false, is_edit_title: false},
            }];

            return;
        }

        const modules = [];

        selectedProduct.value.lessons.forEach(lesson => {
            const newLesson = {...lesson};

            if (selectedProduct.value?.product_levels?.length) {
                selectedProduct.value.product_levels.forEach(tariff => {
                    tariff?.product_level_lessons?.forEach(el => {
                        if (el.lesson_id === lesson.id) {
                            newLesson.product_level_id = el.product_level_id;
                        }
                    })
                })
            }

            if (lesson.module_name === '') {
                modules.push({
                    name: '',
                    items: [newLesson],
                    other_keys: {is_edit_options_open: false, is_edit_title: false},
                });
            } else {
                let module = modules.find(m => m.name === newLesson.module_name);
                if (!module) {
                    module = {
                        name: newLesson.module_name,
                        items: [],
                        other_keys: {is_edit_options_open: false, is_edit_title: false},
                    };
                    modules.push(module);
                }
                module.items.push(newLesson);
            }
        });

        selectedProductModules.value = modules;
    });


    //chapters logic
    const setChapter = (chapter) => {
        selectedChapter.value = chapter;
    }

    //lessons logic
    const setContentType = (val) => selectedContentType.value = val;
    const setLessonTab = (tab) => selectedLessonTab.value = tab;
    const setHomeworkType = (type) => lessonHomeworkType.value = type;
    const setPaidLessons = (paid) => paidLessons.value = paid;
    const setSelectedLesson = (lesson) => selectedLesson.value = lesson;

    const openLesson = (item) => {
        if (item.material) {
            lessonMaterials.value = item.material;
        }

        const contentKey = Object.keys(item.content)[0];
        lessonContentData.value[contentKey] = item.content[contentKey];

        selectedContentType.value = item.type;
    }

    watch(() => selectedLesson.value, (newVal) => {
        if (newVal && newVal.materials) {
            lessonMaterials.value = [];
            lessonContentData.value = {
                title: '',
                description: '',
                cover: '',
                cover_file: null,
                cover_name: '',
                cover_size: 0,
                filename: '',
                file: null,
                circle_filename: '',
                circle_file: null,
                youtube: '',
            };
            lessonTabs.value = [];
            const content = [];

            newVal.materials.forEach((material) => {
                if (material.category === 'lesson_cover' || material.category === 'lesson_content') {
                    lessonMaterialsFromBackend.value.content.push(material);
                    content.push(material);

                    if (!lessonTabs.value.includes('content')) {
                        lessonTabs.value.push('content');
                    }
                } else if (material.category === 'materials') {
                    lessonMaterialsFromBackend.value.materials.push(material);
                    lessonMaterials.value.push(material);

                    if (!lessonTabs.value.includes('materials')) {
                        lessonTabs.value.push('materials');
                    }
                }
            })

            if (content.length) {
                selectedVideoType.value = content[0]?.content_type === 'video' ? 'default' : 'circular';
                lessonContentData.value.title = content[0]?.title;
                lessonContentData.value.description = content[0]?.description;
                lessonContentData.value.cover = content[1]?.filename || '';
                lessonContentData.value.cover_name = content[1]?.original_filename || '';
                lessonContentData.value.cover_size = content[1]?.size || 0;
                lessonContentData.value.youtube = content[0]?.url || '';

                if (content[0].content_type === 'video' || content[0]?.content_type === 'audio') {
                    lessonContentData.value.video_material_id = content[0].id;
                    lessonContentData.value.filename = content[0]?.filename || '';
                    lessonContentData.value.metadata = content[0]?.metadata;

                } else {
                    lessonContentData.value.video_material_id = content[0].id;
                    lessonContentData.value.circle_filename = content[0]?.filename || '';
                    lessonContentData.value.circle_metadata = content[0]?.metadata;
                }
            }

            if (!isLessonCompleted.value) {
              isLessonCompleted.value = productProgress.value.find(item => item.lesson_id === selectedLesson.value?.id);
            }

            if (!isLessonReviewed.value) {
              isLessonReviewed.value = productReviews.value.find(review => review.lesson_id === selectedLesson.value?.id);
            }
        }
    })

    const resetLessonData = () => {
        selectedVideoType.value = '';
        lessonMaterials.value = [];
        lessonContentData.value = {
            title: '',
            description: '',
            cover: '',
            cover_file: null,
            cover_name: '',
            cover_size: 0,
            filename: '',
            file: null,
            circle_filename: '',
            circle_file: null,
            youtube: '',
        };
        selectedLessonTab.value = null;
        selectedContentType.value = null;
        selectedLesson.value = null;
        isLessonCompleted.value = false;
        isLessonReviewed.value = false;
    }

    return {
        allProducts,
        selectedProduct,
        selectedContentType,
        lessonMaterials,
        lessonHomeworkTask,
        selectedLessonTab,
        lessonHomeworkData,
        selectedChapter,
        lessonHomeworkType,
        lessonContentData,
        selectedProductModules,
        selectedLesson,
        selectedVideoType,
        isMediaLoading,
        lessonMaterialsFromBackend,
        isLessonExitModalOpen,
        lessonTabs,
        productProgress,
        productReviews,
        isLessonReviewed,
        isLessonCompleted,
        availableProducts,
        paidLessons,
        productsAccess,
        productInvite,
        selectedPaymentTariff,

        setSelectedProduct,
        removeSelectedProduct,
        setContentType,
        setLessonTab,
        setSelectedLesson,
        setChapter,
        setHomeworkType,
        openLesson,
        setAllProducts,
        resetLessonData,
        setProductProgress,
        setProductReviews,
        setPaidLessons,
        setProductAccess,
        setProductInvite,
        setSelectedPaymentTariff,
    }
})