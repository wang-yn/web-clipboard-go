import React, { useEffect, useMemo, useState } from 'react';
import {
    Copy,
    Download,
    FileIcon,
    FileText,
    FolderOpen,
    Image as ImageIcon,
    Save,
    Upload,
    X
} from 'lucide-react';
import { AccountMenu } from './account.jsx';
import { Auth } from './auth.js';
import { i18n } from './i18n.js';
import { IconLabel, StatusMessage, useMessage } from './shared.jsx';
import './styles.css';

const e = React.createElement;

function isImageItem(item) {
    return item.contentType?.startsWith('image/');
}

function RecentTypeIcon({ type, contentType }) {
    const image = contentType?.startsWith('image/');
    const Icon = type === 'text' ? FileText : image ? ImageIcon : FileIcon;
    return e('span', {
        className: type === 'text'
            ? 'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-blue-700'
            : image
                ? 'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-700'
                : 'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-amber-100 text-amber-700',
        'aria-label': type === 'text' ? 'Text item' : image ? 'Image item' : 'File item'
    },
        e(Icon, { size: 18, 'aria-hidden': true }),
        e('span', { className: 'sr-only' }, type === 'text' ? 'Text item' : image ? 'Image item' : 'File item')
    );
}

export function AppShell() {
    const [user, setUser] = useState(Auth.getCurrentUser());
    const [ready, setReady] = useState(false);
    const [language, setLanguage] = useState(i18n.getCurrentLanguage());
    const [message, showMessage] = useMessage();

    useEffect(() => {
        Auth.requireAuth().then((authenticated) => {
            if (!authenticated) {
                return;
            }
            setUser(Auth.getCurrentUser());
            setReady(true);
        });
    }, []);

    function switchLanguage(lang) {
        i18n.setLanguage(lang);
        setLanguage(lang);
    }

    if (!ready) {
        return e('main', { className: 'min-h-screen flex items-center justify-center text-gray-600' }, 'Loading...');
    }

    return e('div', { className: 'container mx-auto max-w-6xl py-4 px-2 sm:px-4' },
        e(AccountMenu, {
            user,
            language,
            onLanguageChange: switchLanguage,
            settingsHref: '/settings.html'
        }),
        e('h1', { className: 'text-2xl sm:text-3xl font-bold text-center text-gray-800 mt-6 mb-3' }, i18n.t('title')),
        e('p', { className: 'text-center text-sm text-gray-600 mb-6' }, i18n.t('expiry-notice')),
        e(ClipboardPanel, { showMessage }),
        message && e(StatusMessage, { message })
    );
}

function ClipboardPanel({ showMessage }) {
    const [textContent, setTextContent] = useState('');
    const [selectedFile, setSelectedFile] = useState(null);
    const [dragActive, setDragActive] = useState(false);
    const [recentItems, setRecentItems] = useState([]);

    useEffect(() => {
        const timer = setInterval(() => {
            loadRecentItems(false);
            cleanupExpiredItems();
        }, 60000);
        loadRecentItems();
        cleanupExpiredItems();
        return () => clearInterval(timer);
    }, []);

    async function loadRecentItems(showErrors = true) {
        try {
            const data = await Auth.json('/api/items');
            setRecentItems(data.items || []);
        } catch (error) {
            if (showErrors) {
                showMessage(i18n.t('load-recent-failed', error.message), 'error');
            }
        }
    }

    function setRecent(items) {
        setRecentItems(items);
    }

    function cleanupExpiredItems() {
        const now = new Date();
        setRecentItems((currentItems) => currentItems.filter((item) => new Date(item.expiresAt) > now));
    }

    function addToRecent(type, id, description, expiresAt, contentType) {
        const item = {
            type,
            id,
            description,
            contentType,
            createdAt: new Date().toISOString(),
            expiresAt
        };
        setRecentItems((currentItems) => {
            const nextItems = [item, ...currentItems.filter((current) => current.id !== id)].slice(0, 10);
            return nextItems;
        });
    }

    async function saveText() {
        const content = textContent.trim();
        if (!content) {
            showMessage(i18n.t('please-enter-text'), 'error');
            return;
        }

        try {
            const response = await Auth.fetch('/api/text', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content })
            });
            if (!response.ok) {
                throw new Error(i18n.t('failed-save-text'));
            }
            const data = await response.json();
            addToRecent('text', data.id, `${content.substring(0, 50)}...`, data.expiresAt);
            loadRecentItems();
            showMessage(i18n.t('text-saved'));
        } catch (error) {
            showMessage(i18n.t('error-saving-text', error.message), 'error');
        }
    }

    async function copyCurrentText() {
        if (!textContent) {
            showMessage(i18n.t('no-text-to-copy'), 'error');
            return;
        }
        try {
            await navigator.clipboard.writeText(textContent);
            showMessage(i18n.t('text-copied'));
        } catch (error) {
            showMessage(i18n.t('failed-copy-text'), 'error');
        }
    }

    async function uploadFile() {
        if (!selectedFile) {
            showMessage(i18n.t('please-select-file'), 'error');
            return;
        }

        const formData = new FormData();
        formData.append('file', selectedFile);
        try {
            const response = await Auth.fetch('/api/file', {
                method: 'POST',
                body: formData
            });
            if (!response.ok) {
                throw new Error(i18n.t('failed-upload-file'));
            }
            const data = await response.json();
            addToRecent('file', data.id, data.fileName, data.expiresAt, data.contentType);
            loadRecentItems();
            showMessage(i18n.t('file-uploaded'));
        } catch (error) {
            showMessage(i18n.t('error-uploading-file', error.message), 'error');
        }
    }

    function handleDroppedFile(event) {
        event.preventDefault();
        setDragActive(false);
        const file = event.dataTransfer.files[0];
        if (file) {
            setSelectedFile(file);
        }
    }

    return e(React.Fragment, null,
        e('section', { className: 'grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6' },
            e('div', { className: 'bg-white rounded-lg shadow-md p-4 sm:p-6' },
                e('h2', { className: 'text-lg sm:text-xl font-semibold mb-4 text-gray-700' }, i18n.t('text-clipboard')),
                e('textarea', {
                    className: 'w-full h-32 sm:h-40 p-3 border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm',
                    value: textContent,
                    placeholder: i18n.t('text-placeholder'),
                    onChange: (event) => setTextContent(event.target.value)
                }),
                e('div', { className: 'flex flex-col sm:flex-row gap-2 mt-4' },
                    e('button', { className: 'flex-1 bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded-lg font-medium text-sm inline-flex items-center justify-center gap-2', onClick: saveText }, e(IconLabel, { icon: Save, label: i18n.t('save-text') })),
                    e('button', { className: 'flex-1 bg-green-500 hover:bg-green-600 text-white py-2 px-4 rounded-lg font-medium text-sm inline-flex items-center justify-center gap-2', onClick: copyCurrentText }, e(IconLabel, { icon: Copy, label: i18n.t('copy-text') }))
                )
            ),
            e('div', { className: 'bg-white rounded-lg shadow-md p-4 sm:p-6' },
                e('h2', { className: 'text-lg sm:text-xl font-semibold mb-4 text-gray-700' }, i18n.t('file-clipboard')),
                e('input', {
                    type: 'file',
                    id: 'fileInput',
                    className: 'hidden',
                    onChange: (event) => setSelectedFile(event.target.files[0] || null)
                }),
                e('button', {
                    className: `w-full p-4 border-2 border-dashed rounded-lg text-sm ${dragActive ? 'border-blue-500 text-blue-500 bg-blue-50' : 'border-gray-300 text-gray-600 hover:border-blue-500 hover:text-blue-500'}`,
                    onClick: () => document.getElementById('fileInput').click(),
                    onDragOver: (event) => {
                        event.preventDefault();
                        setDragActive(true);
                    },
                    onDragLeave: () => setDragActive(false),
                    onDrop: handleDroppedFile
                }, e(IconLabel, { icon: FolderOpen, label: i18n.t('select-file') })),
                selectedFile && e('div', { className: 'mt-2 text-sm text-gray-600' },
                    i18n.t('selected-file', selectedFile.name, (selectedFile.size / 1024 / 1024).toFixed(2))
                ),
                e('button', {
                    className: 'w-full mt-4 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white py-2 px-4 rounded-lg font-medium text-sm',
                    disabled: !selectedFile,
                    onClick: uploadFile
                }, e(IconLabel, { icon: Upload, label: i18n.t('upload-file') }))
            )
        ),
        e(RecentItems, { items: recentItems, setRecent, showMessage })
    );
}

function RecentItems({ items, setRecent, showMessage }) {
    const [imagePreview, setImagePreview] = useState(null);
    const validItems = useMemo(() => {
        const now = new Date();
        return items.filter((item) => new Date(item.expiresAt) > now);
    }, [items]);

    async function copyTextItem(id) {
        try {
            const response = await Auth.fetch(`/api/text/${id}`);
            if (response.status === 404) {
                showMessage(i18n.t('text-not-found'), 'error');
                return;
            }
            if (!response.ok) {
                throw new Error(i18n.t('failed-load-text'));
            }
            const data = await response.json();
            await navigator.clipboard.writeText(data.content);
            showMessage(i18n.t('text-copied'));
        } catch (error) {
            showMessage(i18n.t('failed-copy-text'), 'error');
        }
    }

    async function downloadFile(id) {
        try {
            const response = await Auth.fetch(`/api/file/${id}`);
            if (response.status === 404) {
                showMessage(i18n.t('file-not-found'), 'error');
                return;
            }
            if (!response.ok) {
                throw new Error(i18n.t('failed-download-file'));
            }
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = getDownloadFilename(response.headers.get('content-disposition'));
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            window.URL.revokeObjectURL(url);
            showMessage(i18n.t('file-downloaded'));
        } catch (error) {
            showMessage(i18n.t('error-downloading-file', error.message), 'error');
        }
    }

    async function previewImage(item) {
        try {
            const response = await Auth.fetch(`/api/file/${item.id}`);
            if (response.status === 404) {
                showMessage(i18n.t('file-not-found'), 'error');
                return;
            }
            if (!response.ok) {
                throw new Error(i18n.t('failed-load-image'));
            }
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            setImagePreview((current) => {
                if (current?.url) {
                    URL.revokeObjectURL(current.url);
                }
                return {
                    url,
                    fileName: item.fileName || item.description
                };
            });
        } catch (error) {
            showMessage(i18n.t('error-loading-image', error.message), 'error');
        }
    }

    function closeImagePreview() {
        setImagePreview((current) => {
            if (current?.url) {
                URL.revokeObjectURL(current.url);
            }
            return null;
        });
    }

    function loadItem(type, id) {
        if (type === 'text') {
            return copyTextItem(id);
        }
        return downloadFile(id);
    }

    useEffect(() => {
        if (validItems.length !== items.length) {
            setRecent(validItems);
        }
    }, [validItems.length]);

    return e('section', { className: 'mt-6 sm:mt-8 bg-white rounded-lg shadow-md p-4 sm:p-6' },
        e('h2', { className: 'text-lg sm:text-xl font-semibold mb-4 text-gray-700' }, i18n.t('recent-items')),
        validItems.length === 0
            ? e('p', { className: 'text-gray-500 text-center text-sm' }, i18n.t('no-recent-items'))
            : e('div', { className: 'space-y-2' }, validItems.map((item) =>
                e('div', { key: item.id, className: 'flex items-center justify-between p-3 bg-gray-50 rounded border' },
                    e('div', { className: 'flex-1 min-w-0' },
                        e('div', { className: 'flex items-center gap-2' },
                            e(RecentTypeIcon, { type: item.type, contentType: item.contentType }),
                            e('span', { className: 'font-medium text-sm truncate' }, item.description)
                        ),
                        e('div', { className: 'text-xs text-gray-500 mt-1' }, i18n.t('created', new Date(item.createdAt).toLocaleString()))
                    ),
                    e('div', { className: 'flex shrink-0 items-center gap-2' },
                        isImageItem(item) && e('button', {
                            className: 'px-3 py-2 bg-blue-100 hover:bg-blue-200 text-blue-700 rounded text-xs',
                            title: i18n.t('item-action-preview-image'),
                            onClick: () => previewImage(item)
                        }, e(IconLabel, {
                            icon: ImageIcon,
                            label: i18n.t('item-action-preview-image')
                        })),
                        e('button', {
                            className: 'px-3 py-2 bg-green-100 hover:bg-green-200 text-green-700 rounded text-xs',
                            title: item.type === 'text' ? i18n.t('item-action-copy-text') : i18n.t('item-action-download-file'),
                            onClick: () => loadItem(item.type, item.id)
                        }, e(IconLabel, {
                            icon: item.type === 'text' ? Copy : Download,
                            label: item.type === 'text' ? i18n.t('item-action-copy-text') : i18n.t('item-action-download-file')
                        }))
                    )
                )
            )),
        imagePreview && e('div', { className: 'fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-70 p-4', role: 'dialog', 'aria-modal': 'true', 'aria-label': i18n.t('image-preview-title') },
            e('div', { className: 'w-full max-w-4xl rounded-lg bg-white p-3 shadow-xl' },
                e('div', { className: 'mb-3 flex items-center justify-between gap-3' },
                    e('h3', { className: 'truncate text-base font-semibold text-gray-800' }, imagePreview.fileName || i18n.t('image-preview-title')),
                    e('button', {
                        className: 'inline-flex h-9 w-9 items-center justify-center rounded bg-gray-100 text-gray-700 hover:bg-gray-200',
                        title: i18n.t('close'),
                        onClick: closeImagePreview
                    }, e(X, { size: 18, 'aria-hidden': true }), e('span', { className: 'sr-only' }, i18n.t('close')))
                ),
                e('img', {
                    className: 'max-h-[75vh] w-full rounded object-contain',
                    src: imagePreview.url,
                    alt: imagePreview.fileName || i18n.t('image-preview-title')
                })
            )
        )
    );
}

function getDownloadFilename(contentDisposition) {
    if (!contentDisposition) {
        return 'download';
    }
    const encodedMatch = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
    if (encodedMatch) {
        try {
            return decodeURIComponent(encodedMatch[1]);
        } catch (error) {
            return 'download';
        }
    }
    const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
    if (filenameMatch) {
        return filenameMatch[1].replace(/['"]/g, '');
    }
    return 'download';
}
