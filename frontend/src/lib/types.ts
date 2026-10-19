export type SectionType = 'contact' | 'skills' | 'experience' | 'education' | 'custom';

export interface Profile {
	name: string;
	headline: string;
}

export interface Detail {
	id: number;
	itemId: number;
	text: string;
	included: boolean;
}

export interface Item {
	id: number;
	sectionId: number;
	data: Record<string, any>;
	details: Detail[];
	included: boolean;
}

export interface Section {
	id: number;
	type: SectionType;
	title: string;
	/** Optional "#rrggbb"; empty = default color. */
	titleColor?: string;
	/** Optional "#rrggbb"; empty = default color. */
	lineColor?: string;
	visible: boolean;
	items: Item[];
}

export interface Preset {
	id: number;
	name: string;
	createdAt: string;
}

/** Typography of one text role on the resume. */
export interface TextStyle {
	size: number;
	bold: boolean;
	italic: boolean;
}

/** Global typography, applied to every section of every preset. */
export interface TextStyles {
	sectionTitle: TextStyle;
	entryTitle: TextStyle;
	entryDates: TextStyle;
	body: TextStyle;
}

export type StyleRole = keyof TextStyles;

export const DEFAULT_TEXT_STYLES: TextStyles = {
	sectionTitle: { size: 10.5, bold: true, italic: false },
	entryTitle: { size: 10, bold: true, italic: false },
	entryDates: { size: 9, bold: false, italic: false },
	body: { size: 9.5, bold: false, italic: false }
};

export interface ResumeState {
	profile: Profile;
	styles: TextStyles;
	presets: Preset[];
	sections: Section[];
}

/** Pagination computed by the backend, matching the PDF export. */
export interface PageLayout {
	/** Section IDs that start pages after the first. */
	breaks: number[];
	pageCount: number;
}

export const SECTION_TYPE_LABELS: Record<SectionType, string> = {
	contact: 'Contact',
	skills: 'Skills',
	experience: 'Experience',
	education: 'Education',
	custom: 'Custom'
};
