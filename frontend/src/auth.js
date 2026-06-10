export class Auth {
    static getToken() {
        return localStorage.getItem('authToken');
    }

    static setToken(token) {
        localStorage.setItem('authToken', token);
    }

    static getTokenExpiry() {
        return localStorage.getItem('tokenExpiry');
    }

    static setTokenExpiry(expiresAt) {
        localStorage.setItem('tokenExpiry', expiresAt);
    }

    static clearAuth() {
        localStorage.removeItem('authToken');
        localStorage.removeItem('tokenExpiry');
        localStorage.removeItem('currentUser');
    }

    static getCurrentUser() {
        const userStr = localStorage.getItem('currentUser');
        if (!userStr) {
            return null;
        }
        try {
            return JSON.parse(userStr);
        } catch (error) {
            return null;
        }
    }

    static setCurrentUser(user) {
        localStorage.setItem('currentUser', JSON.stringify(user));
    }

    static async login(username, password, rememberMe = false) {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password, rememberMe })
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Login failed' }));
            throw new Error(error.error || 'Login failed');
        }

        const data = await response.json();
        this.setToken(data.token);
        this.setTokenExpiry(data.expiresAt);
        this.setCurrentUser(data.user);
        return data.user;
    }

    static async logout() {
        const token = this.getToken();
        try {
            if (token) {
                await fetch('/api/auth/logout', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
            }
        } finally {
            this.clearAuth();
            window.location.href = '/login.html';
        }
    }

    static async checkAuth() {
        const token = this.getToken();
        if (!token) {
            return false;
        }

        const expiry = this.getTokenExpiry();
        if (expiry && new Date(expiry) < new Date()) {
            this.clearAuth();
            return false;
        }

        try {
            const response = await this.fetch('/api/auth/me');
            if (!response.ok) {
                this.clearAuth();
                return false;
            }
            const user = await response.json();
            this.setCurrentUser(user);
            return true;
        } catch (error) {
            this.clearAuth();
            return false;
        }
    }

    static getAuthHeader() {
        const token = this.getToken();
        return token ? `Bearer ${token}` : '';
    }

    static async fetch(url, options = {}) {
        const token = this.getToken();
        if (!token) {
            throw new Error('Not authenticated');
        }

        const headers = {
            ...(options.headers || {}),
            'Authorization': `Bearer ${token}`
        };

        const response = await fetch(url, {
            ...options,
            headers
        });

        if (response.status === 401) {
            this.clearAuth();
            window.location.href = '/login.html';
            throw new Error('Authentication required');
        }

        return response;
    }

    static async json(url, options = {}) {
        const response = await this.fetch(url, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...(options.headers || {})
            }
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Request failed' }));
            throw new Error(error.error || 'Request failed');
        }

        return response.json();
    }

    static isAdmin() {
        const user = this.getCurrentUser();
        return Boolean(user && user.role === 'admin');
    }

    static async requireAuth() {
        const isAuthenticated = await this.checkAuth();
        if (!isAuthenticated) {
            window.location.href = '/login.html';
            return false;
        }
        return true;
    }

    static async changePassword(userId, newPassword) {
        return this.json(`/api/users/${userId}/password`, {
            method: 'PUT',
            body: JSON.stringify({ newPassword })
        });
    }
}
