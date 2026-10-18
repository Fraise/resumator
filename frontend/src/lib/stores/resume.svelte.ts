import { api, hasBackend } from '../api';
import type { Item, ResumeState, Section, SectionType, Detail } from '../types';

const SAVE_DEBOUNCE_MS = 400;

const timers = new Map<string, ReturnType<typeof setTimeout>>();

function debounce(key: string, fn: () => void) {
	clearTimeout(timers.get(key));
	timers.set(key, setTimeout(() => {
		timers.delete(key);
		fn();
	}, SAVE_DEBOUNCE_MS));
}

export function flushSave(key: string) {
	const t = timers.get(key);
	if (t) {
		clearTimeout(t);
		timers.delete(key);
	}
}

// selectedSectionId === 0 means the Profile view is selected.
export const PROFILE_VIEW = 0;

class ResumeStore {
	state = $state<ResumeState | null>(null);
	activePresetId = $state(0);
	selectedSectionId = $state(PROFILE_VIEW);
	status = $state('');
	error = $state('');

	get loaded() {
		return this.state !== null;
	}

	get sections(): Section[] {
		return this.state?.sections ?? [];
	}

	get presets() {
		return this.state?.presets ?? [];
	}

	get selectedSection(): Section | null {
		return this.sections.find((s) => s.id === this.selectedSectionId) ?? null;
	}

	fail(e: unknown) {
		this.error = String(e);
	}

	setStatus(text: string) {
		this.status = text;
		if (text) {
			setTimeout(() => {
				if (this.status === text) this.status = '';
			}, 4000);
		}
	}

	async init() {
		if (!hasBackend) {
			this.error = 'Backend not available (running outside Wails).';
			return;
		}
		try {
			const bootstrap = await api.getResumeState(0);
			this.activePresetId = bootstrap.presets[0]?.id ?? 0;
			this.state = await api.getResumeState(this.activePresetId);
			if (this.sections.length > 0) {
				this.selectedSectionId = this.sections[0].id;
			}
		} catch (e) {
			this.fail(e);
		}
	}

	async switchPreset(presetId: number) {
		try {
			this.state = await api.getResumeState(presetId);
			this.activePresetId = presetId;
			const stillThere = this.sections.some((s) => s.id === this.selectedSectionId);
			if (!stillThere) {
				this.selectedSectionId = this.sections[0]?.id ?? PROFILE_VIEW;
			}
		} catch (e) {
			this.fail(e);
		}
	}

	async reload() {
		try {
			this.state = await api.getResumeState(this.activePresetId);
		} catch (e) {
			this.fail(e);
		}
	}

	// ---- profile -----------------------------------------------------

	updateProfileField(field: 'name' | 'headline', value: string) {
		if (!this.state) return;
		this.state.profile[field] = value;
		debounce('profile', () => {
			api.updateProfile(this.state!.profile).catch((e) => this.fail(e));
		});
	}

	// ---- sections ----------------------------------------------------

	async addSection(type: SectionType, title: string) {
		try {
			await api.addSection(type, title);
			await this.reload();
			this.setStatus(`Section "${title}" added`);
		} catch (e) {
			this.fail(e);
		}
	}

	updateSectionTitle(sectionId: number, title: string) {
		const section = this.sections.find((s) => s.id === sectionId);
		if (!section) return;
		section.title = title;
		debounce(`section-title:${sectionId}`, () => {
			api.updateSection(sectionId, title).catch((e) => this.fail(e));
		});
	}

	async deleteSection(sectionId: number) {
		try {
			await api.deleteSection(sectionId);
			if (this.selectedSectionId === sectionId) {
				this.selectedSectionId = this.sections[0]?.id ?? PROFILE_VIEW;
			}
			await this.reload();
			this.setStatus('Section deleted');
		} catch (e) {
			this.fail(e);
		}
	}

	async reorderSections(ids: number[]) {
		if (!this.state) return;
		const previous = this.sections;
		this.state.sections = ids
			.map((id) => previous.find((s) => s.id === id))
			.filter((s): s is Section => !!s);
		try {
			await api.reorderSections(ids);
		} catch (e) {
			this.fail(e);
			await this.reload();
		}
	}

	async setSectionVisible(sectionId: number, visible: boolean) {
		const section = this.sections.find((s) => s.id === sectionId);
		if (!section) return;
		section.visible = visible;
		try {
			await api.setSectionVisible(this.activePresetId, sectionId, visible);
		} catch (e) {
			this.fail(e);
		}
	}

	// ---- items -------------------------------------------------------

	async addItem(sectionId: number) {
		try {
			await api.addItem(sectionId);
			await this.reload();
		} catch (e) {
			this.fail(e);
		}
	}

	updateItemData(item: Item, field: string, value: string) {
		item.data[field] = value;
		debounce(`item:${item.id}`, () => {
			api.updateItemData(item.id, item.data).catch((e) => this.fail(e));
		});
	}

	async deleteItem(itemId: number) {
		try {
			await api.deleteItem(itemId);
			await this.reload();
		} catch (e) {
			this.fail(e);
		}
	}

	async reorderItems(sectionId: number, ids: number[]) {
		const section = this.sections.find((s) => s.id === sectionId);
		if (!section) return;
		const previous = section.items ?? [];
		section.items = ids
			.map((id) => previous.find((i) => i.id === id))
			.filter((i): i is Item => !!i);
		try {
			await api.reorderItems(sectionId, ids);
		} catch (e) {
			this.fail(e);
			await this.reload();
		}
	}

	// A skill group is a marker item with a title; the skills listed after
	// it belong to the group.
	async addSkillGroup(sectionId: number) {
		try {
			const item = await api.addItem(sectionId);
			await api.updateItemData(item.id, { divider: true, title: '' });
			await this.reload();
		} catch (e) {
			this.fail(e);
		}
	}

	// Adds a skill and places it at the end of the given skill group's run
	// (dividerId null = the ungrouped skills before any divider).
	async addItemToGroup(sectionId: number, dividerId: number | null) {
		const section = this.sections.find((s) => s.id === sectionId);
		if (!section) return;
		try {
			const created = await api.addItem(sectionId);
			const ids = (section.items ?? []).map((i) => i.id);
			ids.push(created.id);
			const byId = new Map((section.items ?? []).map((i) => [i.id, i]));
			let insertAt = ids.length - 1;
			if (dividerId === null) {
				let count = 0;
				for (const it of section.items ?? []) {
					if (it.data.divider) break;
					count++;
				}
				insertAt = count;
			} else {
				const divIdx = ids.indexOf(dividerId);
				if (divIdx >= 0) {
					// Walk to the end of the group's run (before the next divider).
					let pos = divIdx + 1;
					while (pos < ids.length - 1 && !byId.get(ids[pos])?.data.divider) pos++;
					insertAt = pos;
				}
			}
			ids.splice(ids.indexOf(created.id), 1);
			ids.splice(insertAt, 0, created.id);
			await api.reorderItems(sectionId, ids);
			await this.reload();
		} catch (e) {
			this.fail(e);
		}
	}

	async toggleItem(itemId: number, included: boolean) {
		const item = this.findItem(itemId);
		if (item) item.included = included;
		try {
			await api.toggleItem(this.activePresetId, itemId, included);
		} catch (e) {
			this.fail(e);
		}
	}

	findItem(itemId: number): Item | null {
		for (const section of this.sections) {
			const item = (section.items ?? []).find((i) => i.id === itemId);
			if (item) return item;
		}
		return null;
	}

	// ---- details -----------------------------------------------------

	async addDetail(itemId: number) {
		try {
			await api.addDetail(itemId);
			await this.reload();
		} catch (e) {
			this.fail(e);
		}
	}

	updateDetail(detail: Detail, text: string) {
		detail.text = text;
		debounce(`detail:${detail.id}`, () => {
			api.updateDetail(detail.id, text).catch((e) => this.fail(e));
		});
	}

	async deleteDetail(detailId: number) {
		try {
			await api.deleteDetail(detailId);
			await this.reload();
		} catch (e) {
			this.fail(e);
		}
	}

	async toggleDetail(detailId: number, included: boolean) {
		const detail = this.findDetail(detailId);
		if (detail) detail.included = included;
		try {
			await api.toggleDetail(this.activePresetId, detailId, included);
		} catch (e) {
			this.fail(e);
		}
	}

	async reorderDetails(itemId: number, ids: number[]) {
		const item = this.findItem(itemId);
		if (!item) return;
		const previous = item.details ?? [];
		item.details = ids
			.map((id) => previous.find((d) => d.id === id))
			.filter((d): d is Detail => !!d);
		try {
			await api.reorderDetails(itemId, ids);
		} catch (e) {
			this.fail(e);
			await this.reload();
		}
	}

	findDetail(detailId: number): Detail | null {
		for (const section of this.sections) {
			for (const item of section.items ?? []) {
				const detail = (item.details ?? []).find((d) => d.id === detailId);
				if (detail) return detail;
			}
		}
		return null;
	}

	// ---- presets -----------------------------------------------------

	async createPreset(name: string) {
		try {
			const preset = await api.createPreset(name);
			await this.switchPreset(preset.id);
			this.setStatus(`Preset "${name}" created`);
		} catch (e) {
			this.fail(e);
		}
	}

	async duplicatePreset(name: string) {
		try {
			const preset = await api.duplicatePreset(name, this.activePresetId);
			await this.switchPreset(preset.id);
			this.setStatus(`Preset "${name}" created`);
		} catch (e) {
			this.fail(e);
		}
	}

	async renamePreset(id: number, name: string) {
		try {
			await api.renamePreset(id, name);
			await this.reload();
			this.setStatus('Preset renamed');
		} catch (e) {
			this.fail(e);
		}
	}

	async deletePreset(id: number) {
		try {
			await api.deletePreset(id);
			const next = this.presets[0]?.id ?? 0;
			await this.switchPreset(next);
			this.setStatus('Preset deleted');
		} catch (e) {
			this.fail(e);
		}
	}

	// ---- export ------------------------------------------------------

	async exportPDF() {
		try {
			const path = await api.exportPDF(this.activePresetId);
			if (path) this.setStatus(`Exported to ${path}`);
		} catch (e) {
			this.fail(e);
		}
	}
}

export const resume = new ResumeStore();
