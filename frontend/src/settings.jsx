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
    const [systemSettings, setSystemSettings] = useState(null);

    useEffect(() => {
        Auth.requireAuth().then((authenticated) => {
            if (!authenticated) {
                window.location.href = '/login.html';
                return;
            }
            const currentUser = Auth.getCurrentUser();
            setUser(currentUser);
            setReady(true);
            if (currentUser?.role === 'admin') {
                loadSystemSettings();
            }
        });
    }, []);

    async function loadSystemSettings() {
        try {
            const settings = await Auth.getSettings();
            setSystemSettings(settings);
        } catch (error) {
            showMessage(i18n.t('load-settings-failed', error.message), 'error');
        }
    }

    async function saveSystemSettings(settings) {
        try {
            const updated = await Auth.updateSettings(settings);
            setSystemSettings(updated);
            showMessage(i18n.t('settings-saved'));
        } catch (error) {
            showMessage(i18n.t('save-settings-failed', error.message), 'error');
        }
    }

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
            user?.role === 'admin' && e(SystemSettings, {
                settings: systemSettings,
                onSubmit: saveSystemSettings
            })
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

function SystemSettings({ settings, onSubmit }) {
    const [form, setForm] = useState(settings);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        setForm(settings);
    }, [settings]);

    if (!form) {
        return e('section', { className: 'bg-white rounded-lg shadow-md p-4 sm:p-6' },
            e('h2', { className: 'text-lg sm:text-xl font-semibold text-gray-700' }, i18n.t('system-settings')),
            e('p', { className: 'mt-2 text-sm text-gray-600' }, 'Loading...')
        );
    }

    function update(path, value) {
        setForm((current) => {
            const next = structuredClone(current);
            let target = next;
            for (let i = 0; i < path.length - 1; i++) {
                target = target[path[i]];
            }
            target[path[path.length - 1]] = value;
            return next;
        });
    }

    function updateDomains(value) {
        update(['auth', 'allowedEmailDomains'], value.split(',').map((item) => item.trim()).filter(Boolean));
    }

    async function submit(event) {
        event.preventDefault();
        setSaving(true);
        try {
            await onSubmit(form);
        } finally {
            setSaving(false);
        }
    }

    const expirationNever = form.clipboard.expirationUnit === 'never';

    return e('section', { className: 'bg-white rounded-lg shadow-md p-4 sm:p-6' },
        e('h2', { className: 'text-lg sm:text-xl font-semibold text-gray-700' }, i18n.t('system-settings')),
        e('form', { className: 'mt-4 space-y-5', onSubmit: submit },
            e('div', { className: 'grid grid-cols-1 md:grid-cols-2 gap-4' },
                e(ToggleField, {
                    label: i18n.t('password-login-enabled'),
                    checked: form.auth.passwordLoginEnabled,
                    onChange: (checked) => update(['auth', 'passwordLoginEnabled'], checked)
                }),
                e(ToggleField, {
                    label: i18n.t('oauth-auto-provision'),
                    checked: form.auth.oauthAutoProvision,
                    onChange: (checked) => update(['auth', 'oauthAutoProvision'], checked)
                })
            ),
            e(TextField, {
                label: i18n.t('allowed-email-domains'),
                value: (form.auth.allowedEmailDomains || []).join(', '),
                onChange: updateDomains
            }),
            e(OAuthProviderSettings, {
                name: 'google',
                title: i18n.t('google-login'),
                config: form.auth.google,
                onChange: (config) => update(['auth', 'google'], config)
            }),
            e(OAuthProviderSettings, {
                name: 'github',
                title: i18n.t('github-login'),
                config: form.auth.github,
                onChange: (config) => update(['auth', 'github'], config)
            }),
            e('div', { className: 'border-t pt-4' },
                e('h3', { className: 'text-base font-semibold text-gray-700 mb-3' }, i18n.t('clipboard-expiration')),
                e('div', { className: 'grid grid-cols-1 sm:grid-cols-2 gap-4' },
                    e('label', { className: 'block' },
                        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, i18n.t('expiration-value')),
                        e('input', {
                            type: 'number',
                            min: 1,
                            disabled: expirationNever,
                            className: 'w-full p-3 border border-gray-300 rounded-lg disabled:bg-gray-100',
                            value: form.clipboard.expirationValue || '',
                            onChange: (event) => update(['clipboard', 'expirationValue'], Number(event.target.value))
                        })
                    ),
                    e('label', { className: 'block' },
                        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, i18n.t('expiration-unit')),
                        e('select', {
                            className: 'w-full p-3 border border-gray-300 rounded-lg',
                            value: form.clipboard.expirationUnit,
                            onChange: (event) => update(['clipboard', 'expirationUnit'], event.target.value)
                        },
                            e('option', { value: 'minute' }, i18n.t('minutes')),
                            e('option', { value: 'hour' }, i18n.t('hours')),
                            e('option', { value: 'day' }, i18n.t('days')),
                            e('option', { value: 'never' }, i18n.t('never'))
                        )
                    )
                )
            ),
            e('div', { className: 'flex justify-end' },
                e('button', {
                    type: 'submit',
                    disabled: saving,
                    className: 'inline-flex items-center justify-center rounded-lg bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600 disabled:opacity-50'
                }, saving ? i18n.t('saving') : i18n.t('save-system-settings'))
            )
        )
    );
}

function ToggleField({ label, checked, onChange }) {
    return e('label', { className: 'flex items-center justify-between gap-3 rounded-lg border border-gray-200 p-3 text-sm text-gray-700' },
        e('span', { className: 'font-medium' }, label),
        e('input', {
            type: 'checkbox',
            checked,
            onChange: (event) => onChange(event.target.checked),
            className: 'h-4 w-4 text-blue-600'
        })
    );
}

function TextField({ label, value, onChange, type = 'text', placeholder = '' }) {
    return e('label', { className: 'block' },
        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, label),
        e('input', {
            type,
            className: 'w-full p-3 border border-gray-300 rounded-lg',
            value,
            placeholder,
            onChange: (event) => onChange(event.target.value)
        })
    );
}

function OAuthProviderSettings({ title, config, onChange }) {
    function update(key, value) {
        onChange({ ...config, [key]: value });
    }

    return e('fieldset', { className: 'border border-gray-200 rounded-lg p-4' },
        e('legend', { className: 'px-1 text-sm font-semibold text-gray-700' }, title),
        e('div', { className: 'space-y-3' },
            e(ToggleField, {
                label: i18n.t('provider-enabled'),
                checked: Boolean(config.enabled),
                onChange: (checked) => update('enabled', checked)
            }),
            e(TextField, {
                label: 'Client ID',
                value: config.clientId || '',
                onChange: (value) => update('clientId', value)
            }),
            e(TextField, {
                label: config.clientSecretSet ? i18n.t('client-secret-configured') : 'Client Secret',
                type: 'password',
                value: config.clientSecret || '',
                placeholder: config.clientSecretSet ? i18n.t('leave-blank-to-keep') : '',
                onChange: (value) => update('clientSecret', value)
            }),
            config.clientSecretSet && e(ToggleField, {
                label: i18n.t('clear-client-secret'),
                checked: Boolean(config.clearClientSecret),
                onChange: (checked) => update('clearClientSecret', checked)
            })
        )
    );
}

createRoot(document.getElementById('root')).render(e(SettingsShell));
