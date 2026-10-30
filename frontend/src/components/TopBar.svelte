<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import { theme } from '../lib/theme.svelte';
	import Icon from './Icon.svelte';
	import logoUrl from '../assets/images/logo-64.png';

	let showNewForm = $state(false);
	let showSaveAsForm = $state(false);
	let showRenameForm = $state(false);
	let confirmDelete = $state(false);
	let newName = $state('');
	let saveAsName = $state('');
	let renameName = $state('');

	const activePreset = $derived(
		resume.presets.find((p) => p.id === resume.activePresetId)
	);

	function createPreset() {
		if (!newName.trim()) return;
		resume.createPreset(newName.trim());
		newName = '';
		showNewForm = false;
	}

	function saveAsPreset() {
		if (!saveAsName.trim()) return;
		resume.duplicatePreset(saveAsName.trim());
		saveAsName = '';
		showSaveAsForm = false;
	}

	function startRename() {
		renameName = activePreset?.name ?? '';
		showRenameForm = true;
	}

	function submitRename() {
		if (!renameName.trim() || !activePreset) return;
		resume.renamePreset(activePreset.id, renameName.trim());
		showRenameForm = false;
	}

	function onDeletePreset() {
		if (!confirmDelete) {
			confirmDelete = true;
			setTimeout(() => (confirmDelete = false), 3000);
			return;
		}
		if (activePreset) resume.deletePreset(activePreset.id);
		confirmDelete = false;
	}

	function onKey(event: KeyboardEvent, submit: () => void, cancel: () => void) {
		if (event.key === 'Enter') submit();
		if (event.key === 'Escape') cancel();
	}
</script>

<header class="topbar">
	<div class="brand">
		<img src={logoUrl} alt="Resumator" class="brand-logo" />
	</div>
	<div class="presets">
		<span class="preset-label">Resume</span>
		<select
			value={resume.activePresetId}
			onchange={(e) => resume.switchPreset(Number((e.currentTarget as HTMLSelectElement).value))}
		>
			{#each resume.presets as preset (preset.id)}
				<option value={preset.id}>{preset.name}</option>
			{/each}
		</select>
		<button
			class="btn small"
			title="Duplicate the current selections as a new preset"
			onclick={() => {
				showSaveAsForm = !showSaveAsForm;
				showNewForm = false;
				showRenameForm = false;
			}}>Save as…</button
		>
		<button
			class="btn small"
			title="Create a new empty preset"
			onclick={() => {
				showNewForm = !showNewForm;
				showSaveAsForm = false;
				showRenameForm = false;
			}}>New</button
		>
		<button class="btn small" title="Rename the current preset" onclick={startRename}>Rename</button>
		<button
			class="btn small danger"
			title={resume.presets.length <= 1 ? 'The last preset cannot be deleted' : 'Delete the current preset'}
			disabled={resume.presets.length <= 1}
			onclick={onDeletePreset}>{confirmDelete ? 'Delete?' : 'Delete'}</button
		>
	</div>

	<div class="actions">
		<button class="btn primary" onclick={() => resume.exportPDF()}>
			<Icon name="download" size={14} />
			Export PDF
		</button>
		<button class="icon-btn" title="Toggle theme" onclick={() => theme.toggle()}>
			<Icon name={theme.current === 'light' ? 'moon' : 'sun'} />
		</button>
	</div>

	{#if showNewForm || showSaveAsForm || showRenameForm}
		<div class="inline-forms">
			{#if showNewForm}
				<form onsubmit={(e) => { e.preventDefault(); createPreset(); }}>
					<input
						placeholder="New preset name"
						bind:value={newName}
						onkeydown={(e) => onKey(e, createPreset, () => (showNewForm = false))}
					/>
					<button class="btn small primary" type="submit">Create</button>
				</form>
			{/if}
			{#if showSaveAsForm}
				<form onsubmit={(e) => { e.preventDefault(); saveAsPreset(); }}>
					<input
						placeholder="Save current selections as…"
						bind:value={saveAsName}
						onkeydown={(e) => onKey(e, saveAsPreset, () => (showSaveAsForm = false))}
					/>
					<button class="btn small primary" type="submit">Save</button>
				</form>
			{/if}
			{#if showRenameForm && activePreset}
				<form onsubmit={(e) => { e.preventDefault(); submitRename(); }}>
					<input
						placeholder="Rename preset"
						bind:value={renameName}
						onkeydown={(e) => onKey(e, submitRename, () => (showRenameForm = false))}
					/>
					<button class="btn small primary" type="submit">Rename</button>
				</form>
			{/if}
		</div>
	{/if}
</header>

{#if resume.status || resume.error}
	<div class="messages">
		{#if resume.status}<span class="status">{resume.status}</span>{/if}
		{#if resume.error}<span class="error">{resume.error}</span>{/if}
	</div>
{/if}

<style>
	.topbar {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 10px 16px;
		background: var(--card);
		border-bottom: 1px solid var(--border);
		flex-wrap: wrap;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-right: 8px;
	}

	.brand-logo {
		width: 22px;
		height: 22px;
		border-radius: 6px;
	}

	.presets {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.presets select {
		width: 160px;
		padding: 4px 8px;
	}

	.preset-label {
		font-size: 12px;
		color: var(--muted-foreground);
	}

	.actions {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.inline-forms {
		flex-basis: 100%;
		display: flex;
		gap: 8px;
	}

	.inline-forms form {
		display: flex;
		gap: 6px;
		flex: 1;
		max-width: 420px;
	}

	.messages {
		padding: 4px 16px;
		background: var(--card);
		border-bottom: 1px solid var(--border);
		font-size: 12.5px;
		display: flex;
		gap: 16px;
	}

	.status {
		color: var(--muted-foreground);
	}

	.error {
		color: var(--danger);
	}
</style>
