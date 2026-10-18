// Page format preference for the preview and PDF export.
export type ResumeFormat = 'a4' | 'letter';

const KEY = 'resumator-format';

export const PAGE_SIZES: Record<
	ResumeFormat,
	{ label: string; widthMm: number; heightMm: number }
> = {
	a4: { label: 'A4', widthMm: 210, heightMm: 297 },
	letter: { label: 'US Letter', widthMm: 215.9, heightMm: 279.4 }
};

function initialFormat(): ResumeFormat {
	if (typeof localStorage === 'undefined') return 'a4';
	const saved = localStorage.getItem(KEY);
	return saved === 'letter' ? 'letter' : 'a4';
}

class FormatStore {
	current = $state<ResumeFormat>(initialFormat());

	set(next: ResumeFormat) {
		this.current = next;
		localStorage.setItem(KEY, next);
	}
}

export const format = new FormatStore();
