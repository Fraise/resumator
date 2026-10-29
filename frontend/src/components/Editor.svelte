<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import TypographyEditor from './editors/TypographyEditor.svelte';
	import ContactEditor from './editors/ContactEditor.svelte';
	import SkillsEditor from './editors/SkillsEditor.svelte';
	import ExperienceEditor from './editors/ExperienceEditor.svelte';
	import EducationEditor from './editors/EducationEditor.svelte';
	import CustomEditor from './editors/CustomEditor.svelte';

	const section = $derived(resume.selectedSection);

	let confirmDelete = $state(false);
</script>

<section class="editor">
	{#if section === null}
		<TypographyEditor />
	{:else}
		<div class="section-header">
			<div class="header-row">
				<label class="header-field" title="Name of the section, shown as its title on the resume">
					<span class="legend">Section name</span>
					<input
						class="name-input"
						value={section.title}
						oninput={(e) =>
							resume.updateSectionTitle(section.id, (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<button
					class="btn small danger"
					onclick={() => {
						if (!confirmDelete) {
							confirmDelete = true;
							setTimeout(() => (confirmDelete = false), 3000);
							return;
						}
						resume.deleteSection(section.id);
						confirmDelete = false;
					}}>{confirmDelete ? 'Delete section?' : 'Delete section'}</button
				>
			</div>
			<div class="header-row color-row">
				<label class="header-field color-field" title="Color of the section title on the resume">
					<span class="legend">Name color</span>
					<input
						class="color-input"
						type="color"
						value={section.titleColor || '#3b2f2a'}
						oninput={(e) =>
							resume.setSectionColors(
								section.id,
								(e.currentTarget as HTMLInputElement).value,
								section.lineColor ?? ''
							)}
					/>
				</label>
				<label class="header-field color-field" title="Color of the line under the section title">
					<span class="legend">Line color</span>
					<input
						class="color-input"
						type="color"
						value={section.lineColor || '#64646e'}
						oninput={(e) =>
							resume.setSectionColors(
								section.id,
								section.titleColor ?? '',
								(e.currentTarget as HTMLInputElement).value
							)}
					/>
				</label>
			</div>
		</div>

		<div class="section-body">
			{#if section.type === 'contact'}
				<ContactEditor {section} />
			{:else if section.type === 'skills'}
				<SkillsEditor {section} />
			{:else if section.type === 'experience'}
				<ExperienceEditor {section} />
			{:else if section.type === 'education'}
				<EducationEditor {section} />
			{:else}
				<CustomEditor {section} />
			{/if}
		</div>
	{/if}
</section>

<style>
	.editor {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 16px;
		overflow-y: auto;
	}

	.section-header {
		display: flex;
		flex-direction: column;
		gap: 10px;
		border-bottom: 1px solid var(--border);
		padding-bottom: 14px;
	}

	.header-row {
		display: flex;
		gap: 12px;
		align-items: center;
		min-width: 0;
	}

	.header-row > .header-field {
		flex: 1;
		min-width: 0;
	}

	.header-field {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}

	.header-field .legend {
		font-size: 12px;
		color: var(--muted-foreground);
		white-space: nowrap;
	}

	.name-input {
		flex: 1;
		min-width: 0;
	}

	.color-row {
		gap: 16px;
	}

	.color-field {
		flex: 0 0 auto;
	}

	.color-input {
		width: 34px;
		height: 27px;
		padding: 2px;
		border-radius: 6px;
		cursor: pointer;
	}

	.section-body {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
</style>
