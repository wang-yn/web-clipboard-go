class I18n {
    constructor() {
        this.translations = {
            en: {
                'title': 'Web Clipboard',
                'expiry-notice': 'Content expires after 10 minutes',
                'text-clipboard': 'Text Clipboard',
                'file-clipboard': 'File Clipboard',
                'recent-items': 'Recent Items',
                'text-placeholder': 'Paste or type your text here...',
                'save-text': 'Save Text',
                'copy-text': 'Copy Text',
                'select-file': 'Click to select file or drag & drop',
                'upload-file': 'Upload File',
                'item-action-copy-text': 'Copy text',
                'item-action-download-file': 'Download file',
                'change-password': 'Change Password',
                'user-management': 'User Management',
                'create-user': 'Create User',
                'reset-password': 'Reset Password',
                'edit-user': 'Edit User',
                'delete-user': 'Delete User',
                'logout': 'Logout',
                'username': 'Username',
                'password': 'Password',
                'new-password': 'New Password',
                'confirm-password': 'Confirm Password',
                'email': 'Email',
                'role': 'Role',
                'status': 'Status',
                'active': 'Active',
                'inactive': 'Inactive',
                'admin': 'Admin',
                'user': 'User',
                'save': 'Save',
                'cancel': 'Cancel',
                'close': 'Close',
                'please-enter-text': 'Please enter some text to save',
                'text-saved': 'Text saved. Use Recent Items to copy it.',
                'failed-save-text': 'Failed to save text',
                'error-saving-text': 'Error saving text: {0}',
                'no-text-to-copy': 'No text to copy',
                'text-copied': 'Text copied to clipboard!',
                'failed-copy-text': 'Failed to copy text to clipboard',
                'text-loaded': 'Text loaded successfully!',
                'text-not-found': 'Text not found or expired',
                'failed-load-text': 'Failed to load text',
                'error-loading-text': 'Error loading text: {0}',
                'selected-file': 'Selected: {0} ({1} MB)',
                'please-select-file': 'Please select a file',
                'file-uploaded': 'File uploaded. Use Recent Items to download it.',
                'failed-upload-file': 'Failed to upload file',
                'error-uploading-file': 'Error uploading file: {0}',
                'file-downloaded': 'File downloaded successfully!',
                'file-not-found': 'File not found or expired',
                'failed-download-file': 'Failed to download file',
                'error-downloading-file': 'Error downloading file: {0}',
                'no-recent-items': 'No recent items',
                'created': 'Created: {0}',
                'login-title': 'Please login to continue',
                'remember-me': 'Remember me (7 days)',
                'login': 'Login',
                'login-loading': 'Logging in...',
                'login-success': 'Login successful! Redirecting...',
                'login-failed': 'Login failed: {0}',
                'default-credentials': 'Default credentials:',
                'password-mismatch': 'Passwords do not match',
                'password-too-short': 'Password must be at least 6 characters',
                'password-changed': 'Password changed. Please login again.',
                'users-loaded': 'Users loaded',
                'user-created': 'User created',
                'user-updated': 'User updated',
                'user-deleted': 'User deleted',
                'confirm-delete-user': 'Delete user {0}?',
                'cannot-delete-self': 'You cannot delete your own account here',
                'load-users-failed': 'Failed to load users: {0}',
                'save-user-failed': 'Failed to save user: {0}',
                'delete-user-failed': 'Failed to delete user: {0}'
            },
            zh: {
                'title': '网页剪贴板',
                'expiry-notice': '内容10分钟后过期',
                'text-clipboard': '文本剪贴板',
                'file-clipboard': '文件剪贴板',
                'recent-items': '最近项目',
                'text-placeholder': '在此粘贴或输入文本...',
                'save-text': '保存文本',
                'copy-text': '复制文本',
                'select-file': '点击选择文件或拖放',
                'upload-file': '上传文件',
                'item-action-copy-text': '复制文本',
                'item-action-download-file': '下载文件',
                'change-password': '修改密码',
                'user-management': '用户管理',
                'create-user': '创建用户',
                'reset-password': '重置密码',
                'edit-user': '编辑用户',
                'delete-user': '删除用户',
                'logout': '退出登录',
                'username': '用户名',
                'password': '密码',
                'new-password': '新密码',
                'confirm-password': '确认密码',
                'email': '邮箱',
                'role': '角色',
                'status': '状态',
                'active': '启用',
                'inactive': '停用',
                'admin': '管理员',
                'user': '用户',
                'save': '保存',
                'cancel': '取消',
                'close': '关闭',
                'please-enter-text': '请输入要保存的文本',
                'text-saved': '文本已保存，可在最近项目中复制。',
                'failed-save-text': '保存文本失败',
                'error-saving-text': '保存文本时出错：{0}',
                'no-text-to-copy': '没有文本可复制',
                'text-copied': '文本已复制到剪贴板！',
                'failed-copy-text': '复制文本到剪贴板失败',
                'text-loaded': '文本加载成功！',
                'text-not-found': '文本未找到或已过期',
                'failed-load-text': '加载文本失败',
                'error-loading-text': '加载文本时出错：{0}',
                'selected-file': '已选择：{0} ({1} MB)',
                'please-select-file': '请选择一个文件',
                'file-uploaded': '文件上传成功，可在最近项目中下载。',
                'failed-upload-file': '上传文件失败',
                'error-uploading-file': '上传文件时出错：{0}',
                'file-downloaded': '文件下载成功！',
                'file-not-found': '文件未找到或已过期',
                'failed-download-file': '下载文件失败',
                'error-downloading-file': '下载文件时出错：{0}',
                'no-recent-items': '暂无最近项目',
                'created': '创建时间：{0}',
                'login-title': '请登录后继续',
                'remember-me': '记住我（7天）',
                'login': '登录',
                'login-loading': '登录中...',
                'login-success': '登录成功，正在跳转...',
                'login-failed': '登录失败：{0}',
                'default-credentials': '默认账号：',
                'password-mismatch': '两次输入的密码不一致',
                'password-too-short': '密码至少需要6个字符',
                'password-changed': '密码已修改，请重新登录。',
                'users-loaded': '用户列表已加载',
                'user-created': '用户已创建',
                'user-updated': '用户已更新',
                'user-deleted': '用户已删除',
                'confirm-delete-user': '确认删除用户 {0}？',
                'cannot-delete-self': '不能在这里删除自己的账号',
                'load-users-failed': '加载用户失败：{0}',
                'save-user-failed': '保存用户失败：{0}',
                'delete-user-failed': '删除用户失败：{0}'
            }
        };

        this.currentLang = this.detectLanguage();
    }

    detectLanguage() {
        const savedLang = localStorage.getItem('language');
        if (savedLang && this.translations[savedLang]) {
            return savedLang;
        }

        const browserLang = navigator.language || navigator.userLanguage || '';
        return browserLang.startsWith('zh') ? 'zh' : 'en';
    }

    setLanguage(lang) {
        if (!this.translations[lang]) {
            return;
        }
        this.currentLang = lang;
        localStorage.setItem('language', lang);
        document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';
    }

    t(key, ...args) {
        let translation = this.translations[this.currentLang][key] || key;
        args.forEach((arg, index) => {
            translation = translation.replace(`{${index}}`, arg);
        });
        return translation;
    }

    getCurrentLanguage() {
        return this.currentLang;
    }
}

let i18n = new I18n();
