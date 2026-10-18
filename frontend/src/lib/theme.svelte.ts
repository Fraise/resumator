// Reactive theme store. Light is default, persisted in localStorage.
const KEY = 'resumator-theme';

function initialTheme(): 'light' | 'dark' {
	if (typeof window === 'undefined') return 'light';
	const saved = localStorage.getItem(KEY);
	if (saved === 'light' || saved === 'dark') return saved;
	return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

class ThemeStore {
	current = $state<'light' | 'dark'>(initialTheme());

	constructor() {
		this.apply();
	}

	toggle() {
		this.current = this.current === 'light' ? 'dark' : 'light';
		this.apply();
	}

	private apply() {
		if (typeof document === 'undefined') return;
		document.documentElement.dataset.theme = this.current;
		localStorage.setItem(KEY, this.current);
	}
}

export const theme = new ThemeStore();
