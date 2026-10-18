<script lang="ts">
	import { resume, PROFILE_VIEW } from '../lib/stores/resume.svelte';
	import type { SectionType } from '../lib/types';
	import Icon from './Icon.svelte';

	let showAddForm = $state(false);
	let newType = $state<SectionType>('custom');
	let newTitle = $state('');

	function addSection() {
		const title = newTitle.trim() || defaultTitle(newType);
		resume.addSection(newType, title);
		newTitle = '';
		showAddForm = false;
	}

	function defaultTitle(type: SectionType): string {
		switch (type) {
			case 'contact':
				return 'Contact';
			case 'skills':
				return 'Skills';
			case 'experience':
				return 'Professional Experience';
			case 'education':
				return 'Education';
			default:
				return 'New Section';
		}
	}
</script>

<nav class="sidebar">
	<div class="nav-group-label">Content</div>
	<button
		class="nav-item"
		class:active={resume.selectedSectionId === PROFILE_VIEW}
		onclick={() => (resume.selectedSectionId = PROFILE_VIEW)}
	>
		<Icon name="user" size={15} />
		<span>Profile</span>
	</button>

	<div class="nav-group-label">Sections</div>
	{#each resume.sections as section, index (section.id)}
		<div
			class="nav-item"
			class:active={resume.selectedSectionId === section.id}
			class:dimmed={!section.visible}
			role="button"
			tabindex="0"
			onclick={() => (resume.selectedSectionId = section.id)}
			onkeydown={(e) => e.key === 'Enter' && (resume.selectedSectionId = section.id)}
		>
			<button
				class="icon-btn"
				title={section.visible ? 'Hide from this resume' : 'Show in this resume'}
				onclick={(e) => {
					e.stopPropagation();
					resume.setSectionVisible(section.id, !section.visible);
				}}
			>
				<Icon name={section.visible ? 'eye' : 'eye-off'} size={15} />
			</button>
			<span class="nav-title">{section.title}</span>
			<span class="nav-count">{section.items?.length ?? 0}</span>
			<span class="nav-reorder">
				<button
					class="icon-btn"
					title="Move up"
					disabled={index === 0}
					onclick={(e) => {
						e.stopPropagation();
						resume.moveSection(section.id, -1);
					}}><Icon name="up" size={13} /></button
				>
				<button
					class="icon-btn"
					title="Move down"
					disabled={index === resume.sections.length - 1}
					onclick={(e) => {
						e.stopPropagation();
						resume.moveSection(section.id, 1);
					}}><Icon name="down" size={13} /></button
				>
			</span>
		</div>
	{/each}

	<button class="add-section btn ghost" onclick={() => (showAddForm = !showAddForm)}>
		<Icon name="plus" size={14} />
		Add section
	</button>

	{#if showAddForm}
		<form class="add-form" onsubmit={(e) => { e.preventDefault(); addSection(); }}>
			<select bind:value={newType} aria-label="Section type">
				<option value="custom">Custom</option>
				<option value="contact">Contact</option>
				<option value="skills">Skills</option>
				<option value="experience">Experience</option>
				<option value="education">Education</option>
			</select>
			<input placeholder="Section title" bind:value={newTitle} />
			<button class="btn small primary" type="submit">Add</button>
		</form>
	{/if}
</nav>

<style>
	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 12px 8px;
		overflow-y: auto;
	}

	.nav-group-label {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--muted-foreground);
		padding: 10px 10px 4px;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 6px 10px;
		border-radius: 8px;
		text-align: left;
		color: var(--foreground);
		cursor: pointer;
	}

	.nav-item:hover {
		background: var(--muted);
	}

	.nav-item.active {
		background: var(--accent);
		color: var(--primary);
	}

	.nav-item.dimmed .nav-title {
		color: var(--muted-foreground);
	}

	.nav-title {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.nav-count {
		font-size: 11px;
		color: var(--muted-foreground);
		background: var(--muted);
		border-radius: 9999px;
		padding: 1px 7px;
	}

	.nav-reorder {
		display: none;
		gap: 1px;
	}

	.nav-item:hover .nav-reorder,
	.nav-item.active .nav-reorder {
		display: inline-flex;
	}

	.add-section {
		justify-content: flex-start;
		margin-top: 6px;
	}

	.add-form {
		display: flex;
		gap: 6px;
		padding: 4px 2px;
	}

	.add-form select {
		width: 110px;
	}
</style>
