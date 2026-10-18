<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Item, Section } from '../../lib/types';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';
	import SortableList from '../SortableList.svelte';
	import DetailsList from '../DetailsList.svelte';

	let { section }: { section: Section } = $props();

	const addLabel = $derived(
		section.items.length === 0 ? 'Add your first entry' : 'Add entry'
	);
</script>

<SortableList items={section.items} reorder={(ids) => resume.reorderItems(section.id, ids)}>
	{#snippet row(item: Item)}
		<ItemCard {item}>
			<label class="field">
				<span>Label (optional)</span>
				<input
					placeholder="Heading for this entry"
					value={item.data.label ?? ''}
					oninput={(e) => resume.updateItemData(item, 'label', (e.currentTarget as HTMLInputElement).value)}
				/>
			</label>
			<label class="field">
				<span>Text</span>
				<textarea
					placeholder="Free-form content"
					value={item.data.text ?? ''}
					oninput={(e) => resume.updateItemData(item, 'text', (e.currentTarget as HTMLTextAreaElement).value)}
				></textarea>
			</label>
			<DetailsList {item} />
		</ItemCard>
	{/snippet}
</SortableList>

<button class="btn ghost add-btn" onclick={() => resume.addItem(section.id)}>
	<Icon name="plus" size={14} />
	{addLabel}
</button>

<style>
	.add-btn {
		justify-content: center;
		border: 1px dashed var(--border);
		border-radius: 12px;
		padding: 10px;
		color: var(--muted-foreground);
	}
</style>
