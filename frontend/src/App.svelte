<script lang="ts">
	import { onMount } from 'svelte';
	import { resume } from './lib/stores/resume.svelte';
	import TopBar from './components/TopBar.svelte';
	import Sidebar from './components/Sidebar.svelte';
	import Editor from './components/Editor.svelte';
	import Preview from './components/Preview.svelte';

	onMount(() => {
		resume.init();
	});
</script>

<TopBar />

<div class="split">
	<div class="pane pane-left">
		<Sidebar />
		<Editor />
	</div>
	<div class="pane pane-right">
		<Preview />
	</div>
</div>

{#if !resume.loaded && !resume.error}
	<div class="loading">Loading…</div>
{/if}

<style>
	.split {
		flex: 1;
		display: grid;
		grid-template-columns: minmax(420px, 42%) 1fr;
		min-height: 0;
	}

	.pane {
		min-width: 0;
		min-height: 0;
		overflow: hidden;
		display: flex;
	}

	.pane-left {
		border-right: 1px solid var(--border);
		flex-direction: row;
	}

	.pane-left :global(nav.sidebar) {
		width: 232px;
		flex-shrink: 0;
		border-right: 1px solid var(--border);
		background: var(--card);
	}

	.pane-left > :global(.editor) {
		flex: 1;
		background: var(--background);
	}

	.pane-right {
		background: var(--muted);
	}

	.loading {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--background);
		color: var(--muted-foreground);
	}
</style>
