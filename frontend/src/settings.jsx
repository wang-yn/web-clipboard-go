import React, { useEffect, useState } from 'react';
import { ArrowLeft, KeyRound } from 'lucide-react';
import { createRoot } from 'react-dom/client';
import { AccountMenu, ChangePasswordModal } from './account.jsx';
import { Auth } from './auth.js';
import { i18n } from './i18n.js';
import { IconLabel, StatusMessage, useMessage } from './shared.jsx';
import './styles.css';
import { UserManagement } from './users.jsx';

const e = React.createElement;

function SettingsShell() {
    const [user, setUser] = useState(Auth.getCurrentUser());
    const [ready, setReady] = useState(false);
    const [language, setLanguage] = useState(i18n.getCurrentLanguage());
    const [message, showMessage] = useMessage();
    const [passwordOpen, setPasswordOpen] = useState(false);

    useEffect(() => {
        Auth.requireAuth().then((authenticated) => {
            if (!authenticated) {
                window.location.href = '/login.html';
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
        e('main', { className: 'mt-6 space-y-6' },
            e('div', { className: 'flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3' },
                e('div', null,
                    e('h1', { className: 'text-2xl sm:text-3xl font-bold text-gray-800' }, i18n.t('settings-title')),
                    e('p', { className: 'text-sm text-gray-600 mt-1' }, i18n.t('settings-subtitle'))
                ),
                e('a', {
                    href: '/',
                    className: 'inline-flex items-center justify-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50'
                }, e(IconLabel, { icon: ArrowLeft, label: i18n.t('back-to-clipboard') }))
            ),
            e(AccountSettings, {
                user,
                onChangePassword: () => setPasswordOpen(true)
            }),
            user?.role === 'admin' && e(AdminSettings, { currentUser: user, showMessage }),
            e('section', { className: 'bg-white rounded-lg shadow-md p-4 sm:p-6' },
                e('h2', { className: 'text-lg sm:text-xl font-semibold text-gray-700' }, i18n.t('system-settings')),
                e('p', { className: 'mt-2 text-sm text-gray-600' }, i18n.t('system-settings-placeholder'))
            )
        ),
        message && e(StatusMessage, { message }),
        passwordOpen && e(ChangePasswordModal, {
            user,
            onClose: () => setPasswordOpen(false),
            showMessage
        })
    );
}

function AccountSettings({ user, onChangePassword }) {
    return e('section', { className: 'bg-white rounded-lg shadow-md p-4 sm:p-6' },
        e('div', { className: 'flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4' },
            e('div', null,
                e('h2', { className: 'text-lg sm:text-xl font-semibold text-gray-700' }, i18n.t('account-settings')),
                e('dl', { className: 'mt-3 grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm' },
                    e('div', null,
                        e('dt', { className: 'text-gray-500' }, i18n.t('username')),
                        e('dd', { className: 'font-medium text-gray-800' }, user?.username)
                    ),
                    e('div', null,
                        e('dt', { className: 'text-gray-500' }, i18n.t('role')),
                        e('dd', { className: 'font-medium text-gray-800' }, user?.role)
                    ),
                    e('div', null,
                        e('dt', { className: 'text-gray-500' }, i18n.t('email')),
                        e('dd', { className: 'font-medium text-gray-800' }, user?.email || '-')
                    )
                )
            ),
            e('button', {
                className: 'inline-flex items-center justify-center gap-2 rounded-lg bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600',
                onClick: onChangePassword
            }, e(IconLabel, { icon: KeyRound, label: i18n.t('change-password') }))
        )
    );
}

function AdminSettings({ currentUser, showMessage }) {
    return e(UserManagement, { currentUser, showMessage });
}

createRoot(document.getElementById('root')).render(e(SettingsShell));
