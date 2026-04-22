// config.js

const isProduction = process.env.NODE_ENV === 'prod';

// 从 window 对象获取路径前缀（由后端注入或通过 API 获取）
const getPathPrefix = () => {
    return window.__URL_PATH_PREFIX__ || '';
};

// 获取 API 基础 URL
// 生产环境使用相对路径（同域名），开发环境使用本地后端地址
// 支持通过环境变量 VITE_API_URL 覆盖
const getApiUrl = () => {
    // 如果明确指定了 API URL，使用指定的
    if (import.meta.env?.VITE_API_URL) {
        return import.meta.env.VITE_API_URL;
    }
    // 生产环境使用相对路径
    if (isProduction) {
        return '';
    }
    // 开发环境默认使用本地后端
    return 'http://127.0.0.1:8081';
};

const config = {
    apiUrl: getApiUrl(),
    pathPrefix: getPathPrefix(),
};

export default config;

