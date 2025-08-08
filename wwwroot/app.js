class WebClipboard {
    constructor() {
        this.recentItems = JSON.parse(localStorage.getItem('recentItems') || '[]');
        this.init();
    }

    init() {
        this.bindEvents();
        this.updateRecentItems();
    }

    bindEvents() {
        // Text operations
        document.getElementById('saveText').addEventListener('click', () => this.saveText());
        document.getElementById('copyText').addEventListener('click', () => this.copyTextToClipboard());
        document.getElementById('loadText').addEventListener('click', () => this.loadText());

        // File operations
        document.getElementById('selectFile').addEventListener('click', () => {
            document.getElementById('fileInput').click();
        });
        document.getElementById('fileInput').addEventListener('change', (e) => this.handleFileSelect(e));
        document.getElementById('uploadFile').addEventListener('click', () => this.uploadFile());
        document.getElementById('downloadFile').addEventListener('click', () => this.downloadFile());

        // Drag and drop for files
        const fileArea = document.getElementById('selectFile');
        fileArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            fileArea.classList.add('border-blue-500', 'text-blue-500');
        });
        fileArea.addEventListener('dragleave', () => {
            fileArea.classList.remove('border-blue-500', 'text-blue-500');
        });
        fileArea.addEventListener('drop', (e) => {
            e.preventDefault();
            fileArea.classList.remove('border-blue-500', 'text-blue-500');
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                document.getElementById('fileInput').files = files;
                this.handleFileSelect({ target: { files } });
            }
        });

        // Enter key handlers
        document.getElementById('textId').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.loadText();
        });
        document.getElementById('fileId').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.downloadFile();
        });
    }

    async saveText() {
        const content = document.getElementById('textContent').value.trim();
        if (!content) {
            this.showMessage(i18n.t('please-enter-text'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/text', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content })
            });

            if (response.ok) {
                const data = await response.json();
                this.showMessage(i18n.t('text-saved', data.id), 'success');
                this.addToRecent('text', data.id, content.substring(0, 50) + '...', data.expiresAt);
                document.getElementById('textId').value = data.id;
            } else {
                throw new Error(i18n.t('failed-save-text'));
            }
        } catch (error) {
            this.showMessage(i18n.t('error-saving-text', error.message), 'error');
        }
    }

    async copyTextToClipboard() {
        const content = document.getElementById('textContent').value;
        if (!content) {
            this.showMessage('No text to copy', 'error');
            return;
        }

        try {
            await navigator.clipboard.writeText(content);
            this.showMessage('Text copied to clipboard!', 'success');
        } catch (error) {
            this.showMessage('Failed to copy text to clipboard', 'error');
        }
    }

    async loadText() {
        const id = document.getElementById('textId').value.trim();
        if (!id) {
            this.showMessage('Please enter a text ID', 'error');
            return;
        }

        try {
            const response = await fetch(`/api/text/${id}`);
            if (response.ok) {
                const data = await response.json();
                document.getElementById('textContent').value = data.content;
                this.showMessage('Text loaded successfully!', 'success');
            } else if (response.status === 404) {
                this.showMessage('Text not found or expired', 'error');
            } else {
                throw new Error('Failed to load text');
            }
        } catch (error) {
            this.showMessage('Error loading text: ' + error.message, 'error');
        }
    }

    handleFileSelect(event) {
        const file = event.target.files[0];
        const fileDisplay = document.getElementById('selectedFile');
        const uploadBtn = document.getElementById('uploadFile');

        if (file) {
            const fileSize = (file.size / 1024 / 1024).toFixed(2);
            fileDisplay.textContent = `Selected: ${file.name} (${fileSize} MB)`;
            uploadBtn.disabled = false;
        } else {
            fileDisplay.textContent = '';
            uploadBtn.disabled = true;
        }
    }

    async uploadFile() {
        const fileInput = document.getElementById('fileInput');
        const file = fileInput.files[0];
        
        if (!file) {
            this.showMessage('Please select a file', 'error');
            return;
        }

        const formData = new FormData();
        formData.append('file', file);

        try {
            const response = await fetch('/api/file', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                const data = await response.json();
                this.showMessage(`File uploaded! ID: ${data.id}`, 'success');
                this.addToRecent('file', data.id, data.fileName, data.expiresAt);
                document.getElementById('fileId').value = data.id;
            } else {
                throw new Error('Failed to upload file');
            }
        } catch (error) {
            this.showMessage('Error uploading file: ' + error.message, 'error');
        }
    }

    async downloadFile() {
        const id = document.getElementById('fileId').value.trim();
        if (!id) {
            this.showMessage(i18n.t('please-enter-file-id'), 'error');
            return;
        }

        try {
            const response = await fetch(`/api/file/${id}`);
            if (response.ok) {
                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                
                const contentDisposition = response.headers.get('content-disposition');
                let filename = 'download';
                if (contentDisposition) {
                    const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
                    if (filenameMatch) {
                        filename = filenameMatch[1].replace(/['"]/g, '');
                    }
                }

                a.href = url;
                a.download = filename;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);
                
                this.showMessage(i18n.t('file-downloaded'), 'success');
            } else if (response.status === 404) {
                this.showMessage(i18n.t('file-not-found'), 'error');
            } else {
                throw new Error(i18n.t('failed-download-file'));
            }
        } catch (error) {
            this.showMessage(i18n.t('error-downloading-file', error.message), 'error');
        }
    }

    addToRecent(type, id, description, expiresAt) {
        const item = {
            type,
            id,
            description,
            createdAt: new Date().toISOString(),
            expiresAt
        };

        this.recentItems = [item, ...this.recentItems.filter(i => i.id !== id)].slice(0, 10);
        localStorage.setItem('recentItems', JSON.stringify(this.recentItems));
        this.updateRecentItems();
    }

    updateRecentItems() {
        const container = document.getElementById('recentItems');
        const now = new Date();
        
        const validItems = this.recentItems.filter(item => new Date(item.expiresAt) > now);
        
        if (validItems.length !== this.recentItems.length) {
            this.recentItems = validItems;
            localStorage.setItem('recentItems', JSON.stringify(this.recentItems));
        }

        if (validItems.length === 0) {
            container.innerHTML = `<p class="text-gray-500 text-center text-sm">${i18n.t('no-recent-items')}</p>`;
            return;
        }

        container.innerHTML = validItems.map(item => `
            <div class="flex items-center justify-between p-3 bg-gray-50 rounded border">
                <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                        <span class="text-sm">${item.type === 'text' ? '📝' : '📁'}</span>
                        <span class="font-medium text-sm truncate">${item.description}</span>
                    </div>
                    <div class="text-xs text-gray-500 mt-1">
                        ID: ${item.id} • ${i18n.t('created', new Date(item.createdAt).toLocaleString())}
                    </div>
                </div>
                <div class="flex gap-1 ml-2">
                    <button onclick="app.copyId('${item.id}')" 
                            class="px-2 py-1 bg-blue-100 hover:bg-blue-200 text-blue-600 rounded text-xs"
                            title="Copy ID">
                        📋
                    </button>
                    <button onclick="app.loadItem('${item.type}', '${item.id}')" 
                            class="px-2 py-1 bg-green-100 hover:bg-green-200 text-green-600 rounded text-xs"
                            title="Load">
                        📥
                    </button>
                </div>
            </div>
        `).join('');
    }

    async copyId(id) {
        try {
            await navigator.clipboard.writeText(id);
            this.showMessage(i18n.t('id-copied'), 'success');
        } catch (error) {
            this.showMessage(i18n.t('failed-copy-id'), 'error');
        }
    }

    async loadItem(type, id) {
        if (type === 'text') {
            document.getElementById('textId').value = id;
            await this.loadText();
        } else {
            document.getElementById('fileId').value = id;
            await this.downloadFile();
        }
    }

    showMessage(message, type) {
        const messageEl = document.getElementById('statusMessage');
        messageEl.textContent = message;
        messageEl.className = `mt-4 p-3 rounded-lg text-sm ${
            type === 'success' ? 'bg-green-100 text-green-700 border border-green-200' : 
            'bg-red-100 text-red-700 border border-red-200'
        }`;
        messageEl.classList.remove('hidden');
        
        setTimeout(() => {
            messageEl.classList.add('hidden');
        }, 5000);
    }
}

// Wait for both DOM and i18n to be ready
let app;
function initApp() {
    if (typeof i18n !== 'undefined' && document.readyState !== 'loading') {
        app = new WebClipboard();
    } else {
        // Wait a bit more
        setTimeout(initApp, 10);
    }
}

initApp();