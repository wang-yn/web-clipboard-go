import React from 'react';
import { Save, X } from 'lucide-react';
import { i18n } from './i18n.js';

const e = React.createElement;

export function IconLabel({ icon: Icon, label }) {
    return e('span', { className: 'inline-flex items-center justify-center gap-2' },
        e(Icon, { size: 16, 'aria-hidden': true }),
        e('span', null, label)
    );
}

export function Modal({ title, children, onClose }) {
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

export function ModalActions({ onClose }) {
    return e('div', { className: 'flex justify-end gap-2' },
        e('button', { type: 'button', className: 'px-4 py-2 rounded bg-gray-100 text-gray-700 inline-flex items-center gap-2', onClick: onClose }, e(IconLabel, { icon: X, label: i18n.t('cancel') })),
        e('button', { type: 'submit', className: 'px-4 py-2 rounded bg-blue-500 text-white inline-flex items-center gap-2' }, e(IconLabel, { icon: Save, label: i18n.t('save') }))
    );
}

export function TextField({ label, value, onChange, required = false }) {
    return e('label', { className: 'block' },
        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, label),
        e('input', { className: 'w-full p-2 border rounded', value, required, onChange: (event) => onChange(event.target.value) })
    );
}

export function PasswordField({ label, value, onChange, required = true }) {
    return e('label', { className: 'block' },
        e('span', { className: 'block text-sm font-medium text-gray-700 mb-1' }, label),
        e('input', { type: 'password', className: 'w-full p-2 border rounded', value, required, onChange: (event) => onChange(event.target.value) })
    );
}

export function StatusMessage({ message }) {
    return e('div', {
        role: 'status',
        'aria-live': 'polite',
        className: `fixed top-4 right-4 z-50 max-w-sm rounded-lg px-4 py-3 text-sm shadow-lg ${message.type === 'error'
            ? 'bg-red-50 text-red-700 border border-red-200'
            : 'bg-green-50 text-green-700 border border-green-200'}`
    }, message.text);
}

export function useMessage() {
    const [message, setMessage] = React.useState(null);

    function showMessage(text, type = 'success') {
        setMessage({ text, type });
        setTimeout(() => setMessage(null), 5000);
    }

    return [message, showMessage];
}
