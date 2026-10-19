<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import { DEFAULT_TEXT_STYLES } from '../../lib/types';
	import type { StyleRole, TextStyles } from '../../lib/types';

	const profile = $derived(resume.state?.profile);
	const styles = $derived(resume.state?.styles ?? DEFAULT_TEXT_STYLES);

	const roles: { role: StyleRole; label: string; hint: string }[] = [
		{ role: 'sectionTitle', label: 'Section titles', hint: 'Headings like Skills or Education' },
		{ role: 'entryTitle', label: 'Role & company', hint: 'Entry titles and custom labels' },
		{ role: 'entryDates', label: 'Dates', hint: 'Date ranges next to entries' },
		{ role: 'body', label: 'Body text', hint: 'Summaries, bullets, skills and text' }
	];

	function setSize(role: StyleRole, event: Event) {
		const value = Number((event.currentTarget as HTMLInputElement).value);
		if (Number.isFinite(value) && value >= 6 && value <= 24) {
			resume.updateTextStyle(role, { size: value });
		}
	}

	function toggle(role: StyleRole, key: 'bold' | 'italic') {
		const current = styles[role];
		resume.updateTextStyle(role, { [key]: !current[key] } as Partial<TextStyles[StyleRole]>);
	}
</script>

<h2 class="editor-title">Profile</h2>
<p class="hint">
	The profile appears at the top of every resume. The contact section carries the contact
	details.
</p>

{#if profile}
	<label class="field">
		<span>Full name</span>
		<input
			placeholder="Jane Doe"
			value={profile.name}
			oninput={(e) => resume.updateProfileField('name', (e.currentTarget as HTMLInputElement).value)}
		/>
	</label>
	<label class="field">
		<span>Headline</span>
		<input
			placeholder="Senior Software Engineer"
			value={profile.headline}
			oninput={(e) =>
				resume.updateProfileField('headline', (e.currentTarget as HTMLInputElement).value)}
		/>
	</label>
{/if}

<div class="typography">
	<h3 class="typo-title">Typography</h3>
	<p class="hint">Font size and formatting for every section, applied to all resumes.</p>
	{#each roles as { role, label, hint } (role)}
		<div class="type-row" title={hint}>
			<span class="type-label">{label}</span>
			<label class="size-field">
				<input
					class="size-input"
					type="number"
					min="6"
					max="24"
					step="0.5"
					value={styles[role].size}
					oninput={(e) => setSize(role, e)}
				/>
				<span class="size-unit">pt</span>
			</label>
			<button
				class="fmt-btn"
				class:active={styles[role].bold}
				title="Bold"
				onclick={() => toggle(role, 'bold')}
			>
				<span class="fmt-bold">B</span>
			</button>
			<button
				class="fmt-btn"
				class:active={styles[role].italic}
				title="Italic"
				onclick={() => toggle(role, 'italic')}
			>
				<span class="fmt-italic">I</span>
			</button>
		</div>
	{/each}
</div>

<style>
	.editor-title {
		font-size: 15px;
	}

	.hint {
		margin: 0 0 12px;
		color: var(--muted-foreground);
		font-size: 13px;
	}

	.field {
		display: block;
		margin-bottom: 12px;
	}

	.field > span {
		display: block;
		font-size: 12px;
		color: var(--muted-foreground);
		margin-bottom: 3px;
	}

	.typography {
		margin-top: 12px;
		border-top: 1px solid var(--border);
		padding-top: 14px;
	}

	.typo-title {
		font-size: 14px;
		margin-bottom: 4px;
	}

	.typography .hint {
		margin-bottom: 10px;
	}

	.type-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 5px 0;
	}

	.type-label {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: 13px;
	}

	.size-field {
		display: inline-flex;
		align-items: center;
		gap: 4px;
	}

	.size-input {
		width: 64px;
		text-align: right;
	}

	.size-unit {
		font-size: 12px;
		color: var(--muted-foreground);
	}

	.fmt-btn {
		width: 28px;
		height: 27px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
		color: var(--muted-foreground);
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.fmt-btn:hover {
		background: var(--muted);
	}

	.fmt-btn.active {
		background: var(--accent);
		color: var(--primary);
		border-color: var(--primary);
	}

	.fmt-bold {
		font-weight: 700;
	}

	.fmt-italic {
		font-style: italic;
	}
</style>
