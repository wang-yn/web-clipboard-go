import React, { useEffect, useState } from 'react';
import { KeyRound, LogIn } from 'lucide-react';
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

export function LoginApp() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [rememberMe, setRememberMe] = useState(false);
    const [authProviders, setAuthProviders] = useState([]);
    const [passwordLoginEnabled, setPasswordLoginEnabled] = useState(true);
    const [message, setMessage] = useState(null);
    const [loading, setLoading] = useState(false);
    const [oauthLoading, setOauthLoading] = useState(null);

    useEffect(() => {
        loadAuthProviders();

        const params = new URLSearchParams(window.location.search);
        if (params.get('oauth') === 'complete') {
            setLoading(true);
            Auth.completeOAuthLogin().then(() => {
                setMessage({ type: 'success', text: i18n.t('login-success') });
                window.history.replaceState({}, document.title, '/login.html');
                setTimeout(() => {
                    window.location.href = '/';
                }, 300);
            }).catch((error) => {
                setMessage({ type: 'error', text: i18n.t('oauth-login-failed', error.message) });
                window.history.replaceState({}, document.title, '/login.html');
                setLoading(false);
            });
            return;
        }
        if (params.get('oauth') === 'error') {
            setMessage({ type: 'error', text: i18n.t('oauth-login-failed', params.get('message') || 'OAuth login failed') });
            window.history.replaceState({}, document.title, '/login.html');
        }

        Auth.checkAuth().then((authenticated) => {
            if (authenticated) {
                window.location.href = '/';
            }
        });
    }, []);

    async function loadAuthProviders() {
        const data = await Auth.getAuthProviders();
        setAuthProviders(data.providers || []);
        setPasswordLoginEnabled(data.passwordLoginEnabled !== false);
    }

    async function submit(event) {
        event.preventDefault();
        if (!username.trim() || !password) {
            setMessage({ type: 'error', text: 'Please enter both username and password' });
            return;
        }

        setLoading(true);
        try {
            await Auth.login(username.trim(), password, rememberMe);
            setMessage({ type: 'success', text: i18n.t('login-success') });
            setTimeout(() => {
                window.location.href = '/';
            }, 500);
        } catch (error) {
            setMessage({ type: 'error', text: i18n.t('login-failed', error.message) });
            setLoading(false);
        }
    }

    function providerIcon(provider) {
        return KeyRound;
    }

    return e('main', { className: 'min-h-screen flex items-center justify-center px-4' },
        e('section', { className: 'w-full max-w-md bg-white rounded-lg shadow-md p-8' },
            e('h1', { className: 'text-3xl font-bold text-center text-gray-800 mb-2' }, 'Web Clipboard'),
            e('p', { className: 'text-center text-gray-600 mb-8' }, i18n.t('login-title')),
            passwordLoginEnabled && e('form', { className: 'space-y-4', onSubmit: submit },
                e('label', { className: 'block' },
                    e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, i18n.t('username')),
                    e('input', {
                        className: 'w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent',
                        value: username,
                        onChange: (event) => setUsername(event.target.value),
                        autoComplete: 'username',
                        required: true
                    })
                ),
                e('label', { className: 'block' },
                    e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, i18n.t('password')),
                    e('input', {
                        type: 'password',
                        className: 'w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent',
                        value: password,
                        onChange: (event) => setPassword(event.target.value),
                        autoComplete: 'current-password',
                        required: true
                    })
                ),
                e('label', { className: 'flex items-center text-sm text-gray-700' },
                    e('input', {
                        type: 'checkbox',
                        className: 'h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded',
                        checked: rememberMe,
                        onChange: (event) => setRememberMe(event.target.checked)
                    }),
                    e('span', { className: 'ml-2' }, i18n.t('remember-me'))
                ),
                e('button', {
                    type: 'submit',
                    disabled: loading,
                    className: 'w-full bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white py-3 px-4 rounded-lg font-medium transition-colors inline-flex items-center justify-center gap-2'
                }, e(IconLabel, { icon: LogIn, label: loading ? i18n.t('login-loading') : i18n.t('login') }))
            ),
            authProviders.length > 0 && e('div', { className: 'mt-6 space-y-3' },
                e('div', { className: 'flex items-center gap-3 text-xs text-gray-500' },
                    e('span', { className: 'h-px flex-1 bg-gray-200' }),
                    e('span', null, i18n.t('or-login-with')),
                    e('span', { className: 'h-px flex-1 bg-gray-200' })
                ),
                authProviders.map((provider) => e('button', {
                    key: provider.name,
                    type: 'button',
                    disabled: Boolean(oauthLoading),
                    className: 'w-full border border-gray-300 hover:bg-gray-50 disabled:opacity-50 text-gray-800 py-3 px-4 rounded-lg font-medium transition-colors inline-flex items-center justify-center gap-2',
                    onClick: () => {
                        setOauthLoading(provider.name);
                        Auth.startOAuthLogin(provider.name);
                    }
                }, e(IconLabel, {
                    icon: providerIcon(provider),
                    label: oauthLoading === provider.name ? i18n.t('login-loading') : i18n.t('login-with-provider', provider.displayName)
                })))
            ),
            message && e('div', {
                className: `mt-4 p-3 rounded-lg text-sm ${message.type === 'success'
                    ? 'bg-green-100 text-green-700 border border-green-400'
                    : 'bg-red-100 text-red-700 border border-red-400'}`
            }, message.text)
        )
    );
}
