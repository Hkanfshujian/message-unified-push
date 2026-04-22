import { CodeTemplates } from './code-templates';

const gethttpOrigin = () => {
    return window.location.origin
}

export class TokenEncryption {
    // 根据字符串内容生成确定性 salt（范围 0~255）
    static getDeterministicSalt(text: string) {
        let sum = 0;
        for (let i = 0; i < text.length; i++) {
            sum = (sum + text.charCodeAt(i) * (i + 1)) & 0xFF;
        }
        return sum;
    }

    static encryptHex(text: string, key: number) {
        const salt = TokenEncryption.getDeterministicSalt(text);
        let result = salt.toString(16).padStart(2, '0');
        for (let i = 0; i < text.length; i++) {
            const code = text.charCodeAt(i) ^ (key & 0xFF) ^ ((salt + i) & 0xFF);
            result += code.toString(16).padStart(2, '0');
        }
        return result;
    }

    static decryptHex(enc: string, key: number) {
        if (!enc) {
            throw new Error('empty token');
        }
        if (enc.length % 2 !== 0) {
            throw new Error('invalid hex length');
        }
        const bytes: number[] = [];
        for (let i = 0; i < enc.length; i += 2) {
            const part = enc.slice(i, i + 2);
            const v = parseInt(part, 16);
            if (Number.isNaN(v)) {
                throw new Error('invalid hex token');
            }
            bytes.push(v);
        }
        if (bytes.length < 1) {
            throw new Error('invalid token length');
        }
        const salt = bytes[0];
        const out: number[] = [];
        for (let i = 1; i < bytes.length; i++) {
            const v = bytes[i] ^ (key & 0xFF) ^ ((salt + (i - 1)) & 0xFF);
            out.push(v);
        }
        return String.fromCharCode(...out);
    }
}

// ==================== 公共代码模板生成器 (已移至分离的模块) ====================

// ==================== 模板 API (V2) ====================

export class TemplateApiStrGenerate {
    static getTemplateDataString(template_id: string, placeholders_json: string, options: any = {}) {
        // 解析占位符配置
        let placeholders: any = {};
        try {
            const placeholdersList = JSON.parse(placeholders_json || '[]');
            // 根据占位符配置生成示例值
            placeholdersList.forEach((p: any) => {
                placeholders[p.key] = p.default || `mock_${p.key}`;
            });
        } catch (e) {
            // 如果解析失败，使用默认示例
            placeholders = {
                'username': 'John Doe',
                'email': 'john@example.com',
                'phone': '13800138000'
            };
        }

        let data: any = {
            token: TokenEncryption.encryptHex(template_id, 71),
            title: 'message title',
            placeholders: placeholders
        };

        // 添加动态接收者字段（如果需要）
        if (options.recipients) {
            data.recipients = Array.isArray(options.recipientExample) && options.recipientExample.length > 0
                ? options.recipientExample
                : ['user1@example.com', 'user2@example.com'];
        }
        if (options.waitResult) {
            data.wait_result = true;
        }

        return JSON.stringify(data, null, 4);
    }

    static getApiUrl() {
        return `${gethttpOrigin()}/api/v2/message/send`;
    }

    static getCurlString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getCurl(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getGolangString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getGolang(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getPythonString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getPython(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getJavaString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getJava(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getRustString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getRust(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getPHPString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getPHP(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getNodeString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getNode(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }
}
