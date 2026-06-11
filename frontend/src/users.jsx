import React, { useEffect, useState } from 'react';
import { Pencil, RotateCcw, Trash2, UserPlus } from 'lucide-react';
import { Auth } from './auth.js';
import { i18n } from './i18n.js';
import { IconLabel, Modal, ModalActions, PasswordField, TextField } from './shared.jsx';

const e = React.createElement;

export function UserManagement({ currentUser, showMessage }) {
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

    return e('section', { className: 'mt-6 bg-white rounded-lg shadow-md p-4 sm:p-6' },
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
