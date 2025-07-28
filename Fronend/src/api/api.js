import { apiInstance } from "./instance.js";


// auth
export const signIn = async (data, start_param=false) => apiInstance.post(`/auth/signin${start_param ? '/invite' : ''}`, data);

// profile
export const getMe = async () => apiInstance.get(`/user/me`);
export const editUserProfile = async (formData) => apiInstance.post(`/user/edit`, formData);

// mini-app
export const getMiniApp = async () => apiInstance.get(`/app`);

// product
export const getProduct = async (id) => apiInstance.get(`/app/product/${id}`);

// lessons
export const getLesson = async (id) => apiInstance.get(`/app/lesson/${id}`);
export const submitLesson = async (id, data) => apiInstance.post(`/app/lesson/${id}/submit`, data);
export const rateLesson = async (id, data) => apiInstance.post(`/app/lesson/${id}/review`, data);
export const buyLessonsByTon = async(id) => apiInstance.get(`/app/level/${id}/buy/ton`);
export const buyLessonsByWayForPay = async(id) => apiInstance.get(`/app/level/${id}/buy/wayforpay`);

//payments
export const getPayments = async (data) => apiInstance.post(`/app/payments`, data);

//Mux player
export const getMaterialIdForMuxVideo = async(id) => apiInstance.get(`/app/material/${id}/token`);
