<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import Icon from './Icon.svelte';
	import ProfileEditor from './editors/ProfileEditor.svelte';
	import ContactEditor from './editors/ContactEditor.svelte';
	import SkillsEditor from './editors/SkillsEditor.svelte';
	import ExperienceEditor from './editors/ExperienceEditor.svelte';
	import EducationEditor from './editors/EducationEditor.svelte';
	import CustomEditor from './editors/CustomEditor.svelte';

	const section = $derived(resume.selectedSection);

	let confirmDelete = $state(false);
	let presetTitle = $state('');

	// Track the resolved title to seed the override field when the
	// selected section changes.
	$effect(() => {
		const s = section;
		presetTitle = '';
		if (s) {
			presetTitle = s.title;
		}
	});
</script>

<section class="editor">
	{#if section === null}
		<ProfileEditor />
	{:else}
		<div class="section-header">
			<div class="header-fields">
				<label class="field">
					<span>Section name</span>
					<input
						value={section.title}
						oninput={(e) =>
							resume.updateSectionTitle(section.id, (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Title in this resume (optional)</span>
					<input
						placeholder="Defaults to the section name"
						bind:value={presetTitle}
						oninput={() => resume.setPresetSectionTitle(section.id, presetTitle)}
					/>
				</label>
			</div>
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
		gap: 12px;
		align-items: flex-start;
	}

	.header-fields {
		flex: 1;
		display: flex;
		gap: 12px;
	}

	.header-fields .field {
		flex: 1;
	}

	.section-body {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
</style>
