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
	visible: boolean;
	items: Item[];
}

export interface Preset {
	id: number;
	name: string;
	createdAt: string;
}

export interface ResumeState {
	profile: Profile;
	presets: Preset[];
	sections: Section[];
}

export const SECTION_TYPE_LABELS: Record<SectionType, string> = {
	contact: 'Contact',
	skills: 'Skills',
	experience: 'Experience',
	education: 'Education',
	custom: 'Custom'
};
