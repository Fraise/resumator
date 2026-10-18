<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';

	const profile = $derived(resume.state?.profile);
</script>

<h2 class="editor-title">Profile</h2>
<p class="hint">
	The profile appears at the top of every resume. The contact section carries the
	contact details.
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

<style>
	.editor-title {
		font-size: 15px;
	}

	.hint {
		margin: 0;
		color: var(--muted-foreground);
		font-size: 13px;
	}

	.field {
		margin-bottom: 12px;
	}
</style>
