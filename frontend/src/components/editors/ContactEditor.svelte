<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Item, Section } from '../../lib/types';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';
	import SortableList from '../SortableList.svelte';

	let { section }: { section: Section } = $props();

	const addLabel = $derived(section.items.length === 0 ? 'Add your first contact detail' : 'Add contact detail');
</script>

<p class="hint">Your name and headline appear at the top of every resume.</p>

<label class="field">
	<span>Full name</span>
	<input
		placeholder="Jane Doe"
		value={section.name}
		oninput={(e) =>
			resume.updateContactHeader(section.id, 'name', (e.currentTarget as HTMLInputElement).value)}
	/>
</label>
<label class="field">
	<span>Headline</span>
	<input
		placeholder="Senior Software Engineer"
		value={section.headline}
		oninput={(e) =>
			resume.updateContactHeader(section.id, 'headline', (e.currentTarget as HTMLInputElement).value)}
	/>
</label>

<SortableList items={section.items} reorder={(ids) => resume.reorderItems(section.id, ids)}>
	{#snippet row(item: Item)}
		<ItemCard {item}>
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
	{/snippet}
</SortableList>

<button class="btn ghost add-btn" onclick={() => resume.addItem(section.id)}>
	<Icon name="plus" size={14} />
	{addLabel}
</button>

<style>
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
