import axios from "axios";
import {BASE_URL} from "../constants/const.js";
import CookieManager from "../helpers/cookieManager.js";
import {decodeJwt} from "../helpers/index.js";

export function useToken() {
  const getToken = async () => {
    let token = CookieManager.getItem("student_access_token");

    if (!token) token = await refreshToken();

    return token;
  };

  const refreshToken = async () => {
    const rt = CookieManager.getItem("student_refresh_token");

    const response = await axios.post(
      `${BASE_URL}/auth/refresh`,
      {refresh_token: rt},
    );

    const { access_token, refresh_token } = response.data?.jwt_info;

    setToken(access_token, refresh_token);

    return access_token;
  }

  const setToken = (access_token, refresh_token) => {
    CookieManager.setItem("student_access_token", access_token, decodeJwt(access_token).exp);
    CookieManager.setItem("student_refresh_token", refresh_token, decodeJwt(refresh_token).exp);
  }

  return { getToken, setToken, refreshToken };
}
