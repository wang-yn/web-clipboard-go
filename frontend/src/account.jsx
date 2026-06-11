import React, { useState } from 'react';
import { KeyRound, Languages, LogOut, Save, Settings, X } from 'lucide-react';
import { Auth } from './auth.js';
import { i18n } from './i18n.js';
import { IconLabel, Modal, PasswordField } from './shared.jsx';

const e = React.createElement;

export function AccountMenu({ user, language, onLanguageChange, settingsHref, onChangePassword }) {
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
            settingsHref && e('a', {
                href: '/settings.html',
                className: 'bg-gray-700 hover:bg-gray-800 text-white px-4 py-2 rounded-lg text-sm font-medium inline-flex items-center gap-2'
            }, e(IconLabel, { icon: Settings, label: i18n.t('settings') })),
            onChangePassword && e('button', {
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

export function ChangePasswordModal({ user, onClose, showMessage }) {
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
