class I18n {
    constructor() {
        this.translations = {
            en: {
                'title': 'Web Clipboard',
                'expiry-notice': '⏰ Content expires after 10 minutes',
                'text-clipboard': '📝 Text Clipboard',
                'file-clipboard': '📁 File Clipboard',
                'recent-items': '📋 Recent Items',
                'text-placeholder': 'Paste or type your text here...',
                'text-id-placeholder': 'Enter 4-char ID (e.g. A1B2)',
                'file-id-placeholder': 'Enter 4-char ID (e.g. X9Y8)',
                'save-text': '💾 Save Text',
                'copy-text': '📋 Copy Text',
                'load-text': '📥 Load Text',
                'select-file': '📎 Click to select file or drag & drop',
                'upload-file': '⬆️ Upload File',
                'download-file': '📥 Download File',
                
                // Messages from app.js
                'please-enter-text': 'Please enter some text to save',
                'text-saved': 'Text saved! ID: {0}',
                'failed-save-text': 'Failed to save text',
                'error-saving-text': 'Error saving text: {0}',
                'no-text-to-copy': 'No text to copy',
                'text-copied': 'Text copied to clipboard!',
                'failed-copy-text': 'Failed to copy text to clipboard',
                'please-enter-text-id': 'Please enter a text ID',
                'text-loaded': 'Text loaded successfully!',
                'text-not-found': 'Text not found or expired',
                'failed-load-text': 'Failed to load text',
                'error-loading-text': 'Error loading text: {0}',
                'selected-file': 'Selected: {0} ({1} MB)',
                'please-select-file': 'Please select a file',
                'file-uploaded': 'File uploaded! ID: {0}',
                'failed-upload-file': 'Failed to upload file',
                'error-uploading-file': 'Error uploading file: {0}',
                'please-enter-file-id': 'Please enter a file ID',
                'file-downloaded': 'File downloaded successfully!',
                'file-not-found': 'File not found or expired',
                'failed-download-file': 'Failed to download file',
                'error-downloading-file': 'Error downloading file: {0}',
                'id-copied': 'ID copied to clipboard!',
                'failed-copy-id': 'Failed to copy ID',
                'no-recent-items': 'No recent items',
                'created': 'Created: {0}'
            },
            zh: {
                'title': '网页剪贴板',
                'expiry-notice': '⏰ 内容10分钟后过期',
                'text-clipboard': '📝 文本剪贴板',
                'file-clipboard': '📁 文件剪贴板',
                'recent-items': '📋 最近项目',
                'text-placeholder': '在此粘贴或输入文本...',
                'text-id-placeholder': '输入4位ID（如 A1B2）',
                'file-id-placeholder': '输入4位ID（如 X9Y8）',
                'save-text': '💾 保存文本',
                'copy-text': '📋 复制文本',
                'load-text': '📥 加载文本',
                'select-file': '📎 点击选择文件或拖放',
                'upload-file': '⬆️ 上传文件',
                'download-file': '📥 下载文件',
                
                // Messages from app.js
                'please-enter-text': '请输入要保存的文本',
                'text-saved': '文本已保存！ID：{0}',
                'failed-save-text': '保存文本失败',
                'error-saving-text': '保存文本时出错：{0}',
                'no-text-to-copy': '没有文本可复制',
                'text-copied': '文本已复制到剪贴板！',
                'failed-copy-text': '复制文本到剪贴板失败',
                'please-enter-text-id': '请输入文本ID',
                'text-loaded': '文本加载成功！',
                'text-not-found': '文本未找到或已过期',
                'failed-load-text': '加载文本失败',
                'error-loading-text': '加载文本时出错：{0}',
                'selected-file': '已选择：{0} ({1} MB)',
                'please-select-file': '请选择一个文件',
                'file-uploaded': '文件上传成功！ID：{0}',
                'failed-upload-file': '上传文件失败',
                'error-uploading-file': '上传文件时出错：{0}',
                'please-enter-file-id': '请输入文件ID',
                'file-downloaded': '文件下载成功！',
                'file-not-found': '文件未找到或已过期',
                'failed-download-file': '下载文件失败',
                'error-downloading-file': '下载文件时出错：{0}',
                'id-copied': 'ID已复制到剪贴板！',
                'failed-copy-id': '复制ID失败',
                'no-recent-items': '暂无最近项目',
                'created': '创建时间：{0}'
            }
        };
        
        this.currentLang = this.detectLanguage();
        this.init();
    }
    
    detectLanguage() {
        // Check localStorage first
        const savedLang = localStorage.getItem('language');
        if (savedLang && this.translations[savedLang]) {
            return savedLang;
        }
        
        // Detect browser language
        const browserLang = navigator.language || navigator.userLanguage;
        if (browserLang.startsWith('zh')) {
            return 'zh';
        }
        
        return 'en'; // Default to English
    }
    
    init() {
        // Only initialize DOM-dependent features if DOM is ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
                this.initDOM();
            });
        } else {
            this.initDOM();
        }
    }
    
    initDOM() {
        this.updateLanguageButtons();
        this.translatePage();
        this.bindLanguageButtons();
    }
    
    bindLanguageButtons() {
        const enBtn = document.getElementById('langEn');
        const zhBtn = document.getElementById('langZh');
        
        if (enBtn && zhBtn) {
            enBtn.addEventListener('click', () => {
                this.setLanguage('en');
            });
            
            zhBtn.addEventListener('click', () => {
                this.setLanguage('zh');
            });
        }
    }
    
    setLanguage(lang) {
        if (this.translations[lang]) {
            this.currentLang = lang;
            localStorage.setItem('language', lang);
            document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';
            this.translatePage();
            this.updateLanguageButtons();
        }
    }
    
    updateLanguageButtons() {
        const enBtn = document.getElementById('langEn');
        const zhBtn = document.getElementById('langZh');
        
        if (enBtn && zhBtn) {
            // Reset styles
            enBtn.className = 'px-3 py-1 rounded text-sm font-medium transition-colors';
            zhBtn.className = 'px-3 py-1 rounded text-sm font-medium transition-colors';
            
            // Highlight current language
            if (this.currentLang === 'en') {
                enBtn.className += ' bg-blue-100 text-blue-600';
                zhBtn.className += ' text-gray-600 hover:text-gray-800';
            } else {
                zhBtn.className += ' bg-blue-100 text-blue-600';
                enBtn.className += ' text-gray-600 hover:text-gray-800';
            }
        }
    }
    
    translatePage() {
        // Translate elements with data-i18n attribute
        document.querySelectorAll('[data-i18n]').forEach(element => {
            const key = element.getAttribute('data-i18n');
            const translation = this.translations[this.currentLang][key];
            if (translation) {
                element.textContent = translation;
            }
        });
        
        // Translate placeholders
        document.querySelectorAll('[data-i18n-placeholder]').forEach(element => {
            const key = element.getAttribute('data-i18n-placeholder');
            const translation = this.translations[this.currentLang][key];
            if (translation) {
                element.placeholder = translation;
            }
        });
    }
    
    t(key, ...args) {
        let translation = this.translations[this.currentLang][key] || key;
        
        // Replace placeholders {0}, {1}, etc.
        args.forEach((arg, index) => {
            translation = translation.replace(`{${index}}`, arg);
        });
        
        return translation;
    }
    
    getCurrentLanguage() {
        return this.currentLang;
    }
}

// Initialize i18n immediately
let i18n = new I18n();