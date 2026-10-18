<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Item, Section } from '../../lib/types';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';
	import SortableList from '../SortableList.svelte';

	let { section }: { section: Section } = $props();

	const addLabel = $derived(section.items.length === 0 ? 'Add your first skill' : 'Add skill');
</script>

<SortableList items={section.items} reorder={(ids) => resume.reorderItems(section.id, ids)}>
	{#snippet row(item: Item)}
		<ItemCard {item}>
			<div class="fields">
				<label class="field">
					<span>Skill</span>
					<input
						placeholder="Go"
						value={item.data.name ?? ''}
						oninput={(e) => resume.updateItemData(item, 'name', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Level (optional)</span>
					<input
						placeholder="Expert"
						value={item.data.level ?? ''}
						oninput={(e) => resume.updateItemData(item, 'level', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
			</div>
		</ItemCard>
	{/snippet}
</SortableList>

<button class="btn ghost add-btn" onclick={() => resume.addItem(section.id)}>
	<Icon name="plus" size={14} />
	{addLabel}
</button>

<style>
	.fields {
		display: grid;
		grid-template-columns: 2fr 1fr;
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
