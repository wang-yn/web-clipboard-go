import React, { useEffect, useMemo, useState } from 'react';
import {
    Copy,
    Download,
    FileIcon,
    FileText,
    FolderOpen,
    KeyRound,
    Languages,
    LogOut,
    Pencil,
    RotateCcw,
    Save,
    Trash2,
    Upload,
    UserPlus,
    X
} from 'lucide-react';
import { Auth } from './auth.js';
import { i18n } from './i18n.js';
import './styles.css';

const e = React.createElement;

function IconLabel({ icon: Icon, label }) {
    return e('span', { className: 'inline-flex items-center justify-center gap-2' },
        e(Icon, { size: 16, 'aria-hidden': true }),
        e('span', null, label)
    );
}

function RecentTypeIcon({ type }) {
    const Icon = type === 'text' ? FileText : FileIcon;
    return e('span', {
        className: type === 'text'
            ? 'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-blue-700'
            : 'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-amber-100 text-amber-700',
        'aria-label': type === 'text' ? 'Text item' : 'File item'
    },
        e(Icon, { size: 18, 'aria-hidden': true }),
        e('span', { className: 'sr-only' }, type === 'text' ? 'Text item' : 'File item')
    );
}

function useMessage() {
    const [message, setMessage] = useState(null);

    function showMessage(text, type = 'success') {
        setMessage({ text, type });
        setTimeout(() => setMessage(null), 5000);
    }

    return [message, showMessage];
}

export function AppShell() {
    const [user, setUser] = useState(Auth.getCurrentUser());
    const [ready, setReady] = useState(false);
    const [language, setLanguage] = useState(i18n.getCurrentLanguage());
    const [message, showMessage] = useMessage();
    const [passwordOpen, setPasswordOpen] = useState(false);

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
            onChangePassword: () => setPasswordOpen(true)
        }),
        e('h1', { className: 'text-2xl sm:text-3xl font-bold text-center text-gray-800 mt-6 mb-3' }, i18n.t('title')),
        e('p', { className: 'text-center text-sm text-gray-600 mb-6' }, i18n.t('expiry-notice')),
        e(ClipboardPanel, { showMessage }),
        user?.role === 'admin' && e(UserManagement, { currentUser: user, showMessage }),
        message && e(StatusMessage, { message }),
        passwordOpen && e(ChangePasswordModal, {
            user,
            onClose: () => setPasswordOpen(false),
            showMessage
        })
    );
}

function AccountMenu({ user, language, onLanguageChange, onChangePassword }) {
    return e('header', { className: 'flex flex-col sm:flex-row justify-between gap-3 items-stretch sm:items-center' },
        e('div', { className: 'bg-white rounded-lg shadow-sm px-4 py-2 flex flex-wrap items-center gap-2' },
            e('span', { className: 'text-sm text-gray-600' }, `${i18n.t('user')}:`),
            e('span', { className: 'text-sm font-medium text-gray-800' }, user?.username),
            e('span', { className: 'text-xs px-2 py-1 rounded-full bg-blue-100 text-blue-800' }, user?.role)
        ),
        e('div', { className: 'flex flex-wrap gap-2 justify-end' },
            e('button', {
                className: `px-3 py-2 rounded-lg text-sm font-medium inline-flex items-center gap-2 ${language === 'en' ? 'bg-blue-100 text-blue-700' : 'bg-white text-gray-700'}`,
                onClick: () => onLanguageChange('en')
            }, e(IconLabel, { icon: Languages, label: 'English' })),
            e('button', {
                className: `px-3 py-2 rounded-lg text-sm font-medium inline-flex items-center gap-2 ${language === 'zh' ? 'bg-blue-100 text-blue-700' : 'bg-white text-gray-700'}`,
                onClick: () => onLanguageChange('zh')
            }, e(IconLabel, { icon: Languages, label: '中文' })),
            e('button', {
                className: 'bg-gray-700 hover:bg-gray-800 text-white px-4 py-2 rounded-lg text-sm font-medium inline-flex items-center gap-2',
                onClick: onChangePassword
            }, e(IconLabel, { icon: KeyRound, label: i18n.t('change-password') })),
            e('button', {
                className: 'bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-lg text-sm font-medium inline-flex items-center gap-2',
                onClick: () => Auth.logout()
            }, e(IconLabel, { icon: LogOut, label: i18n.t('logout') }))
        )
    );
}

function ClipboardPanel({ showMessage }) {
    const [textContent, setTextContent] = useState('');
    const [selectedFile, setSelectedFile] = useState(null);
    const [dragActive, setDragActive] = useState(false);
    const [recentItems, setRecentItems] = useState(() => JSON.parse(localStorage.getItem('recentItems') || '[]'));

    useEffect(() => {
        const timer = setInterval(cleanupExpiredItems, 60000);
        cleanupExpiredItems();
        return () => clearInterval(timer);
    }, []);

    function persistRecent(items) {
        setRecentItems(items);
        localStorage.setItem('recentItems', JSON.stringify(items));
    }

    function cleanupExpiredItems() {
        const now = new Date();
        const validItems = JSON.parse(localStorage.getItem('recentItems') || '[]')
            .filter((item) => new Date(item.expiresAt) > now);
        persistRecent(validItems);
    }

    function addToRecent(type, id, description, expiresAt) {
        const item = {
            type,
            id,
            description,
            createdAt: new Date().toISOString(),
            expiresAt
        };
        setRecentItems((currentItems) => {
            const nextItems = [item, ...currentItems.filter((current) => current.id !== id)].slice(0, 10);
            localStorage.setItem('recentItems', JSON.stringify(nextItems));
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
            addToRecent('file', data.id, data.fileName, data.expiresAt);
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
        e(RecentItems, { items: recentItems, persistRecent, showMessage })
    );
}

function RecentItems({ items, persistRecent, showMessage }) {
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

    function loadItem(type, id) {
        if (type === 'text') {
            return copyTextItem(id);
        }
        return downloadFile(id);
    }

    useEffect(() => {
        if (validItems.length !== items.length) {
            persistRecent(validItems);
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
                            e(RecentTypeIcon, { type: item.type }),
                            e('span', { className: 'font-medium text-sm truncate' }, item.description)
                        ),
                        e('div', { className: 'text-xs text-gray-500 mt-1' }, i18n.t('created', new Date(item.createdAt).toLocaleString()))
                    ),
                    e('button', {
                        className: 'px-3 py-2 bg-green-100 hover:bg-green-200 text-green-700 rounded text-xs',
                        title: item.type === 'text' ? i18n.t('item-action-copy-text') : i18n.t('item-action-download-file'),
                        onClick: () => loadItem(item.type, item.id)
                    }, e(IconLabel, {
                        icon: item.type === 'text' ? Copy : Download,
                        label: item.type === 'text' ? i18n.t('item-action-copy-text') : i18n.t('item-action-download-file')
                    }))
                )
            ))
    );
}

function ChangePasswordModal({ user, onClose, showMessage }) {
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [saving, setSaving] = useState(false);

    async function submit(event) {
        event.preventDefault();
        if (password.length < 6) {
            showMessage(i18n.t('password-too-short'), 'error');
            return;
        }
        if (password !== confirmPassword) {
            showMessage(i18n.t('password-mismatch'), 'error');
            return;
        }
        setSaving(true);
        try {
            await Auth.changePassword(user.id, password);
            showMessage(i18n.t('password-changed'));
            Auth.clearAuth();
            setTimeout(() => {
                window.location.href = '/login.html';
            }, 700);
        } catch (error) {
            showMessage(error.message, 'error');
            setSaving(false);
        }
    }

    return e(Modal, { title: i18n.t('change-password'), onClose },
        e('form', { className: 'space-y-4', onSubmit: submit },
            e(PasswordField, { label: i18n.t('new-password'), value: password, onChange: setPassword }),
            e(PasswordField, { label: i18n.t('confirm-password'), value: confirmPassword, onChange: setConfirmPassword }),
            e('div', { className: 'flex justify-end gap-2' },
                e('button', { type: 'button', className: 'px-4 py-2 rounded bg-gray-100 text-gray-700 inline-flex items-center gap-2', onClick: onClose }, e(IconLabel, { icon: X, label: i18n.t('cancel') })),
                e('button', { type: 'submit', disabled: saving, className: 'px-4 py-2 rounded bg-blue-500 text-white disabled:opacity-50 inline-flex items-center gap-2' }, e(IconLabel, { icon: Save, label: i18n.t('save') }))
            )
        )
    );
}

function UserManagement({ currentUser, showMessage }) {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(false);
    const [editingUser, setEditingUser] = useState(null);
    const [resetUser, setResetUser] = useState(null);
    const [creating, setCreating] = useState(false);

    useEffect(() => {
        loadUsers();
    }, []);

    async function loadUsers() {
        setLoading(true);
        try {
            const data = await Auth.json('/api/users');
            setUsers(data.users || []);
        } catch (error) {
            showMessage(i18n.t('load-users-failed', error.message), 'error');
        } finally {
            setLoading(false);
        }
    }

    async function createUser(form) {
        try {
            await Auth.json('/api/users', {
                method: 'POST',
                body: JSON.stringify(form)
            });
            showMessage(i18n.t('user-created'));
            setCreating(false);
            await loadUsers();
        } catch (error) {
            showMessage(i18n.t('save-user-failed', error.message), 'error');
        }
    }

    async function updateUser(id, form) {
        try {
            await Auth.json(`/api/users/${id}`, {
                method: 'PUT',
                body: JSON.stringify(form)
            });
            showMessage(i18n.t('user-updated'));
            setEditingUser(null);
            await loadUsers();
        } catch (error) {
            showMessage(i18n.t('save-user-failed', error.message), 'error');
        }
    }

    async function deleteUser(user) {
        if (user.id === currentUser.id) {
            showMessage(i18n.t('cannot-delete-self'), 'error');
            return;
        }
        if (!window.confirm(i18n.t('confirm-delete-user', user.username))) {
            return;
        }
        try {
            await Auth.json(`/api/users/${user.id}`, { method: 'DELETE' });
            showMessage(i18n.t('user-deleted'));
            await loadUsers();
        } catch (error) {
            showMessage(i18n.t('delete-user-failed', error.message), 'error');
        }
    }

    async function resetUserPassword(user, newPassword) {
        try {
            await Auth.changePassword(user.id, newPassword);
            showMessage(i18n.t('password-changed'));
            setResetUser(null);
        } catch (error) {
            showMessage(error.message, 'error');
        }
    }

    return e('section', { className: 'mt-6 sm:mt-8 bg-white rounded-lg shadow-md p-4 sm:p-6' },
        e('div', { className: 'flex justify-between items-center mb-4 gap-3' },
            e('h2', { className: 'text-lg sm:text-xl font-semibold text-gray-700' }, i18n.t('user-management')),
            e('button', { className: 'px-4 py-2 bg-blue-500 text-white rounded-lg text-sm inline-flex items-center gap-2', onClick: () => setCreating(true) }, e(IconLabel, { icon: UserPlus, label: i18n.t('create-user') }))
        ),
        loading
            ? e('p', { className: 'text-sm text-gray-500' }, 'Loading...')
            : e('div', { className: 'overflow-x-auto' },
                e('table', { className: 'w-full text-sm' },
                    e('thead', null, e('tr', { className: 'text-left border-b' },
                        e('th', { className: 'py-2 pr-3' }, i18n.t('username')),
                        e('th', { className: 'py-2 pr-3' }, i18n.t('email')),
                        e('th', { className: 'py-2 pr-3' }, i18n.t('role')),
                        e('th', { className: 'py-2 pr-3' }, i18n.t('status')),
                        e('th', { className: 'py-2 pr-3' }, '')
                    )),
                    e('tbody', null, users.map((user) => e('tr', { key: user.id, className: 'border-b last:border-0' },
                        e('td', { className: 'py-2 pr-3 font-medium' }, user.username),
                        e('td', { className: 'py-2 pr-3' }, user.email),
                        e('td', { className: 'py-2 pr-3' }, user.role),
                        e('td', { className: 'py-2 pr-3' }, user.isActive ? i18n.t('active') : i18n.t('inactive')),
                        e('td', { className: 'py-2 pr-3' },
                            e('div', { className: 'flex flex-wrap gap-2 justify-end' },
                                e('button', { className: 'px-2 py-1 rounded bg-gray-100 inline-flex items-center gap-1', onClick: () => setEditingUser(user) }, e(IconLabel, { icon: Pencil, label: i18n.t('edit-user') })),
                                e('button', { className: 'px-2 py-1 rounded bg-yellow-100 text-yellow-800 inline-flex items-center gap-1', onClick: () => setResetUser(user) }, e(IconLabel, { icon: RotateCcw, label: i18n.t('reset-password') })),
                                e('button', { className: 'px-2 py-1 rounded bg-red-100 text-red-700 inline-flex items-center gap-1', onClick: () => deleteUser(user) }, e(IconLabel, { icon: Trash2, label: i18n.t('delete-user') }))
                            )
                        )
                    )))
                )
            ),
        creating && e(UserFormModal, {
            title: i18n.t('create-user'),
            onClose: () => setCreating(false),
            onSubmit: createUser,
            includePassword: true,
            includeStatus: false
        }),
        editingUser && e(UserFormModal, { title: i18n.t('edit-user'), user: editingUser, onClose: () => setEditingUser(null), onSubmit: (form) => updateUser(editingUser.id, form) }),
        resetUser && e(ResetPasswordModal, {
            user: resetUser,
            onClose: () => setResetUser(null),
            onSubmit: resetUserPassword,
            showMessage: showMessage
        })
    );
}

function UserFormModal({ title, user, onClose, onSubmit, includePassword = false, includeStatus = true }) {
    const [form, setForm] = useState({
        username: user?.username || '',
        password: '',
        email: user?.email || '',
        role: user?.role || 'user',
        isActive: user?.isActive ?? true
    });

    function setField(name, value) {
        setForm((current) => ({ ...current, [name]: value }));
    }

    function submit(event) {
        event.preventDefault();
        const payload = includePassword
            ? { username: form.username, password: form.password, email: form.email, role: form.role }
            : { email: form.email, role: form.role, isActive: form.isActive };
        onSubmit(payload);
    }

    return e(Modal, { title, onClose },
        e('form', { className: 'space-y-4', onSubmit: submit },
            includePassword && e(TextField, { label: i18n.t('username'), value: form.username, onChange: (value) => setField('username', value), required: true }),
            includePassword && e(PasswordField, { label: i18n.t('password'), value: form.password, onChange: (value) => setField('password', value), required: true }),
            e(TextField, { label: i18n.t('email'), value: form.email, onChange: (value) => setField('email', value), required: true }),
            e('label', { className: 'block' },
                e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, i18n.t('role')),
                e('select', { className: 'w-full p-2 border rounded', value: form.role, onChange: (event) => setField('role', event.target.value) },
                    e('option', { value: 'user' }, i18n.t('user')),
                    e('option', { value: 'admin' }, i18n.t('admin'))
                )
            ),
            includeStatus && e('label', { className: 'flex items-center gap-2 text-sm text-gray-700' },
                e('input', { type: 'checkbox', checked: form.isActive, onChange: (event) => setField('isActive', event.target.checked) }),
                i18n.t('active')
            ),
            e(ModalActions, { onClose })
        )
    );
}

function ResetPasswordModal({ user, onClose, onSubmit, showMessage }) {
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');

    function submit(event) {
        event.preventDefault();
        if (password.length < 6) {
            showMessage(i18n.t('password-too-short'), 'error');
            return;
        }
        if (password !== confirmPassword) {
            showMessage(i18n.t('password-mismatch'), 'error');
            return;
        }
        onSubmit(user, password);
    }

    return e(Modal, { title: `${i18n.t('reset-password')}: ${user.username}`, onClose },
        e('form', { className: 'space-y-4', onSubmit: submit },
            e(PasswordField, { label: i18n.t('new-password'), value: password, onChange: setPassword }),
            e(PasswordField, { label: i18n.t('confirm-password'), value: confirmPassword, onChange: setConfirmPassword }),
            e(ModalActions, { onClose })
        )
    );
}

function Modal({ title, children, onClose }) {
    return e('div', { className: 'fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center p-4 z-50' },
        e('div', { className: 'bg-white rounded-lg shadow-lg w-full max-w-lg p-5' },
            e('div', { className: 'flex justify-between items-center mb-4' },
                e('h3', { className: 'text-lg font-semibold text-gray-800' }, title),
                e('button', { className: 'text-gray-500 hover:text-gray-800', 'aria-label': i18n.t('cancel'), onClick: onClose },
                    e(X, { size: 20, 'aria-hidden': true })
                )
            ),
            children
        )
    );
}

function ModalActions({ onClose }) {
    return e('div', { className: 'flex justify-end gap-2' },
        e('button', { type: 'button', className: 'px-4 py-2 rounded bg-gray-100 text-gray-700 inline-flex items-center gap-2', onClick: onClose }, e(IconLabel, { icon: X, label: i18n.t('cancel') })),
        e('button', { type: 'submit', className: 'px-4 py-2 rounded bg-blue-500 text-white inline-flex items-center gap-2' }, e(IconLabel, { icon: Save, label: i18n.t('save') }))
    );
}

function TextField({ label, value, onChange, required = false }) {
    return e('label', { className: 'block' },
        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, label),
        e('input', { className: 'w-full p-2 border rounded', value, required, onChange: (event) => onChange(event.target.value) })
    );
}

function PasswordField({ label, value, onChange, required = true }) {
    return e('label', { className: 'block' },
        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, label),
        e('input', { type: 'password', className: 'w-full p-2 border rounded', value, required, onChange: (event) => onChange(event.target.value) })
    );
}

function StatusMessage({ message }) {
    return e('div', {
        role: 'status',
        'aria-live': 'polite',
        className: `fixed top-4 right-4 z-50 max-w-sm rounded-lg px-4 py-3 text-sm shadow-lg ${message.type === 'error'
            ? 'bg-red-50 text-red-700 border border-red-200'
            : 'bg-green-50 text-green-700 border border-green-200'}`
    }, message.text);
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
