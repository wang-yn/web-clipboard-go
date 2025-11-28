// Authentication utility class
class Auth {
    // Get the stored token
    static getToken() {
        return localStorage.getItem('authToken');
    }

    // Set the token
    static setToken(token) {
        localStorage.setItem('authToken', token);
    }

    // Get token expiry time
    static getTokenExpiry() {
        return localStorage.getItem('tokenExpiry');
    }

    // Set token expiry time
    static setTokenExpiry(expiresAt) {
        localStorage.setItem('tokenExpiry', expiresAt);
    }

    // Clear authentication data
    static clearAuth() {
        localStorage.removeItem('authToken');
        localStorage.removeItem('tokenExpiry');
        localStorage.removeItem('currentUser');
    }

    // Get current user info
    static getCurrentUser() {
        const userStr = localStorage.getItem('currentUser');
        if (userStr) {
            try {
                return JSON.parse(userStr);
            } catch (e) {
                return null;
            }
        }
        return null;
    }

    // Set current user info
    static setCurrentUser(user) {
        localStorage.setItem('currentUser', JSON.stringify(user));
    }

    // Login
    static async login(username, password, rememberMe = false) {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password,
                    rememberMe: rememberMe
                })
            });

            if (response.ok) {
                const data = await response.json();
                this.setToken(data.token);
                this.setTokenExpiry(data.expiresAt);
                this.setCurrentUser(data.user);
                return true;
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    }

    // Logout
    static async logout() {
        const token = this.getToken();
        if (!token) {
            this.clearAuth();
            window.location.href = '/login.html';
            return;
        }

        try {
            await fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            this.clearAuth();
            window.location.href = '/login.html';
        }
    }

    // Check if user is authenticated
    static async checkAuth() {
        const token = this.getToken();
        if (!token) {
            return false;
        }

        // Check token expiry (client-side check)
        const expiry = this.getTokenExpiry();
        if (expiry && new Date(expiry) < new Date()) {
            this.clearAuth();
            return false;
        }

        // Verify with server
        try {
            const response = await fetch('/api/auth/me', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (response.ok) {
                const user = await response.json();
                this.setCurrentUser(user);
                return true;
            } else {
                this.clearAuth();
                return false;
            }
        } catch (error) {
            console.error('Auth check error:', error);
            return false;
        }
    }

    // Get authorization header
    static getAuthHeader() {
        const token = this.getToken();
        return token ? `Bearer ${token}` : '';
    }

    // Make authenticated fetch request
    static async fetch(url, options = {}) {
        const token = this.getToken();
        if (!token) {
            throw new Error('Not authenticated');
        }

        const headers = {
            ...options.headers,
            'Authorization': `Bearer ${token}`
        };

        const response = await fetch(url, {
            ...options,
            headers
        });

        // Handle unauthorized response
        if (response.status === 401) {
            this.clearAuth();
            window.location.href = '/login.html';
            throw new Error('Authentication required');
        }

        return response;
    }

    // Check if user is admin
    static isAdmin() {
        const user = this.getCurrentUser();
        return user && user.role === 'admin';
    }

    // Require authentication (redirect to login if not authenticated)
    static async requireAuth() {
        const isAuthenticated = await this.checkAuth();
        if (!isAuthenticated) {
            window.location.href = '/login.html';
            return false;
        }
        return true;
    }
}
