// request.js
import axios from 'axios';
import router from '../router';

import { usePageState } from '../store/page_sate';
import { CONSTANT } from '../constant';
import config from '../../config.js';
import { toast } from "vue-sonner"
import { clearAuthzDataStorage } from '@/util/rbacAuthz'


const ERR_NETWORK = "ERR_NETWORK";

// 获取路径前缀
const getPathPrefix = () => {
    return config.pathPrefix || '';
};

const request = axios.create({
    baseURL: config.apiUrl + getPathPrefix(),
    timeout: 50000,
    withCredentials: true,
});



// 请求拦截器
request.interceptors.request.use(
    (config) => {
        const pageState = usePageState();
        const token = pageState.Token || localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || '';
        if (!CONSTANT.NO_AUTH_URL.includes(config.url)) {
            config.url = '/api/v1' + config.url;
        }
        if (token && token.trim() !== '' && !CONSTANT.NO_AUTH_URL.includes(config.url)) {
            config.headers = {
                ...config.headers,
                'm-token': token,
            };
        }
        return config;
    },
    (error) => {
        handleException(error);
    }
);

// 响应拦截器
request.interceptors.response.use(
    (response) => {
        // 检查业务逻辑错误
        if (response && response.data.code != 200) {
            // 检查是否是token相关的错误码
            const tokenErrorCodes = [20001, 20002, 20003, 20004, 20005];
            if (tokenErrorCodes.includes(response.data.code)) {
                // Token失效，执行登出
                logout();
                return Promise.reject(response);
            }
            
            // 其他业务错误显示toast（允许单次请求静默）
            if (!response.config?.meta?.silentBizToast) {
                toast.error(response.data.msg, {
                    description: '接口逻辑错误'
                });
            }
        }
        return response;
    },
    (error) => {
        // HTTP状态码401表示未授权
        if (error.response && error.response.status == 401) {
            // 401错误静默处理，不显示错误提示，直接执行登出
            // 检查响应体中的业务错误码
            if (error.response.data && error.response.data.code) {
                const tokenErrorCodes = [20001, 20002, 20003, 20004, 20005];
                if (tokenErrorCodes.includes(error.response.data.code)) {
                    logout();
                    return Promise.reject(error);
                }
            }
            
            // 其他401错误也执行登出
            logout();
        } else {
            if (!error.config?.meta?.silentErrorToast) {
                handleException(error);
            }
        }
        return Promise.reject(error);
    }
);

// 异常处理
const handleException = (error) => {
    if (!error.response) {
        toast.error('网络请求失败，请确认后端服务已启动', {
            description: error?.message || '无法连接到服务'
        })
        return
    };

    if (error.code == ERR_NETWORK) {
        toast(`网络错误！`)
    } else {
        let msg = `未知错误：${error.response.status}, ${error.response.data.msg}`;
        toast(msg)
    };

};

const clearClientAuthState = (pageState) => {
    pageState.setToken('');
    localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
    localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME);
    localStorage.removeItem(CONSTANT.STORE_OPEN_TABS_NAME || '__message_nest_open_tabs_v1');
    clearAuthzDataStorage();
};
// 登出系统 - 添加防抖标记
let isLoggingOut = false;

const logout = async () => {
    // 防止重复执行登出
    if (isLoggingOut) {
        return;
    }
    isLoggingOut = true;
    
    const pageState = usePageState();
    
    // 检查登录来源
    const authSource = localStorage.getItem(CONSTANT.STORE_AUTH_SOURCE_NAME);
    console.log('[Logout] authSource:', authSource);
    
    // Casdoor 登录用户，调用统一登出
    if (authSource === 'casdoor') {
        try {
            const redirectUri = encodeURIComponent(window.location.origin + '/#/login?casdoor_logout=1');
            const logoutUrlRes = await request.post(`/auth/casdoor/logout?redirect_uri=${redirectUri}`);
            console.log('[Casdoor] 登出接口响应:', logoutUrlRes.data);
            
            if (logoutUrlRes.data?.data?.logout_url) {
                console.log('[Casdoor] 跳转到统一登出 URL');
                clearClientAuthState(pageState);
                window.location.href = logoutUrlRes.data.data.logout_url;
                return;
            }
        } catch (error) {
            console.error('[Casdoor] 统一登出失败:', error);
        }
    }
    
    // 清除token和登录状态
    clearClientAuthState(pageState);
    
    // 立即跳转到登录页
    router.push('/login');
};

export { request, handleException, logout };
