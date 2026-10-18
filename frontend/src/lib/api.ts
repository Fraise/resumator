import * as backend from '../../wailsjs/go/main/App';
import type { Item, PageLayout, Preset, ResumeState, Section } from './types';

// True when running inside the Wails webview with the Go backend bound.
export const hasBackend =
	typeof window !== 'undefined' && !!(window as any).go?.main?.App;

export const api = {
	getResumeState: (presetId: number) =>
		backend.GetResumeState(presetId) as unknown as Promise<ResumeState>,
	getPageLayout: (presetId: number, format: string) =>
		backend.GetPageLayout(presetId, format) as unknown as Promise<PageLayout>,
	updateProfile: (profile: { name: string; headline: string }) =>
		backend.UpdateProfile(profile),
	addSection: (type: string, title: string) => backend.AddSection(type, title),
	updateSection: (id: number, title: string) => backend.UpdateSection(id, title),
	deleteSection: (id: number) => backend.DeleteSection(id),
	reorderSections: (ids: number[]) => backend.ReorderSections(ids),
	addItem: (sectionId: number) =>
		backend.AddItem(sectionId) as unknown as Promise<Item>,
	updateItemData: (id: number, data: Record<string, any>) =>
		backend.UpdateItemData(id, data),
	deleteItem: (id: number) => backend.DeleteItem(id),
	reorderItems: (sectionId: number, ids: number[]) =>
		backend.ReorderItems(sectionId, ids),
	addDetail: (itemId: number) => backend.AddDetail(itemId),
	updateDetail: (id: number, text: string) => backend.UpdateDetail(id, text),
	deleteDetail: (id: number) => backend.DeleteDetail(id),
	reorderDetails: (itemId: number, ids: number[]) =>
		backend.ReorderDetails(itemId, ids),
	toggleItem: (presetId: number, itemId: number, included: boolean) =>
		backend.ToggleItem(presetId, itemId, included),
	toggleDetail: (presetId: number, detailId: number, included: boolean) =>
		backend.ToggleDetail(presetId, detailId, included),
	setSectionVisible: (presetId: number, sectionId: number, visible: boolean) =>
		backend.SetSectionVisible(presetId, sectionId, visible),
	setSectionColors: (sectionId: number, titleColor: string, lineColor: string) =>
		backend.SetSectionColors(sectionId, titleColor, lineColor),
	createPreset: (name: string) =>
		backend.CreatePreset(name) as unknown as Promise<Preset>,
	duplicatePreset: (name: string, sourceId: number) =>
		backend.DuplicatePreset(name, sourceId) as unknown as Promise<Preset>,
	renamePreset: (id: number, name: string) => backend.RenamePreset(id, name),
	deletePreset: (id: number) => backend.DeletePreset(id),
	exportPDF: (presetId: number, format: string) => backend.ExportPDF(presetId, format)
};

export type { ResumeState, Section };
