<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Section } from '../../lib/types';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';

	let { section }: { section: Section } = $props();

	const addLabel = $derived(section.items.length === 0 ? 'Add your first contact detail' : 'Add contact detail');
</script>

{#each section.items as item (item.id)}
	<ItemCard {item} {section}>
		<div class="fields">
			<label class="field">
				<span>Label</span>
				<input
					placeholder="Email"
					value={item.data.label ?? ''}
					oninput={(e) => resume.updateItemData(item, 'label', (e.currentTarget as HTMLInputElement).value)}
				/>
			</label>
			<label class="field">
				<span>Value</span>
				<input
					placeholder="jane@example.com"
					value={item.data.value ?? ''}
					oninput={(e) => resume.updateItemData(item, 'value', (e.currentTarget as HTMLInputElement).value)}
				/>
			</label>
		</div>
	</ItemCard>
{/each}

<button class="btn ghost add-btn" onclick={() => resume.addItem(section.id)}>
	<Icon name="plus" size={14} />
	{addLabel}
</button>

<style>
	.fields {
		display: grid;
		grid-template-columns: 1fr 2fr;
		gap: 10px;
	}

	.add-btn {
		justify-content: center;
		border: 1px dashed var(--border);
		border-radius: 12px;
		padding: 10px;
		color: var(--muted-foreground);
	}
</style>
