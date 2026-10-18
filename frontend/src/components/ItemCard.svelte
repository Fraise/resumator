<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import type { Item, Section } from '../lib/types';
	import Icon from './Icon.svelte';

	let {
		item,
		section,
		children
	}: { item: Item; section: Section; children?: import('svelte').Snippet } = $props();
</script>

<article class="item-card" class:excluded={!item.included}>
	<header class="card-header">
		<label class="include-label" title="Include this entry in the current resume">
			<input
				type="checkbox"
				checked={item.included}
				onchange={(e) =>
					resume.toggleItem(item.id, (e.currentTarget as HTMLInputElement).checked)}
			/>
			In this resume
		</label>
		<span class="card-actions">
			<button
				class="icon-btn"
				disabled={section.items.indexOf(item) === 0}
				title="Move up"
				onclick={() => resume.moveItem(section.id, item.id, -1)}><Icon name="up" size={14} /></button
			>
			<button
				class="icon-btn"
				disabled={section.items.indexOf(item) === (section.items?.length ?? 1) - 1}
				title="Move down"
				onclick={() => resume.moveItem(section.id, item.id, 1)}><Icon name="down" size={14} /></button
			>
			<button class="icon-btn" title="Remove entry" onclick={() => resume.deleteItem(item.id)}>
				<Icon name="trash" size={14} />
			</button>
		</span>
	</header>
	<div class="card-body">
		{@render children?.()}
	</div>
</article>

<style>
	.item-card {
		background: var(--card);
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 12px;
		box-shadow: var(--shadow);
	}

	.item-card.excluded {
		opacity: 0.6;
	}

	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 10px;
	}

	.include-label {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: var(--muted-foreground);
		cursor: pointer;
	}

	.include-label input {
		width: auto;
	}

	.card-actions {
		display: inline-flex;
		gap: 2px;
	}

	.card-body {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
</style>
