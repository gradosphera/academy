import Cookies from 'js-cookie';

class CookieManager {
    static setItem(
        name,
        value,
        expires = Date.now() + 31449600, // 1 year by default
    ) {
        Cookies.set(name, value,
            {
                domain: window.location.hostname === 'localhost' ? 'localhost' : '.academy.gradosphera.org',
                expires: new Date(expires * 1000),
                secure: true,
                sameSite: 'strict',
            }
        );
    }

    static getItem(name) {
        return Cookies.get(name);
    }

    static removeItem(name) {
        Cookies.remove(name, { domain: window.location.hostname === 'localhost' ? 'localhost' : '.academy.gradosphera.org' });
    }

    static clear() {
        const cookies = Object.keys(Cookies.get());
        for (const cookie of cookies) {
            Cookies.remove(cookie, { domain: window.location.hostname === 'localhost' ? 'localhost' : '.academy.gradosphera.org' });
        }
    }

    static getAll() {
        return Cookies.get();
    }
}

export default CookieManager;
