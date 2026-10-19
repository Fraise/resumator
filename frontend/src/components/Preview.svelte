<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import { format, PAGE_SIZES } from '../lib/format.svelte';
	import type { Item, Section } from '../lib/types';
	import Icon from './Icon.svelte';

	const resumeState = $derived(resume.state);
	const styles = $derived(resume.state?.styles);
	const page = $derived(PAGE_SIZES[format.current]);
	const visibleSections = $derived(
		(resume.state?.sections ?? []).filter((s) => s.visible)
	);

	// Pages are chunked from the backend-computed breaks (the same engine
	// as the PDF export), so preview pages match the exported document.
	// The breaks guard tolerates a null payload.
	const pages = $derived.by(() => {
		const breaks = resume.pageLayout.breaks ?? [];
		const result: Section[][] = [];
		let current: Section[] = [];
		for (const section of visibleSections) {
			if (section.type === 'contact') continue;
			if (current.length > 0 && breaks.includes(section.id)) {
				result.push(current);
				current = [];
			}
			current.push(section);
		}
		if (current.length > 0 || result.length === 0) {
			result.push(current);
		}
		return result;
	});

	let currentPage = $state(0);
	// Stays in range when content shrinks or presets switch.
	$effect(() => {
		if (currentPage > pages.length - 1) {
			currentPage = Math.max(0, pages.length - 1);
		}
	});

	function included<T extends { included: boolean }>(entries: T[] | null | undefined): T[] {
		return entries?.filter((e) => e.included) ?? [];
	}

	function contactLine(section: Section): string {
		return included(section.items)
			.map((i) => String(i.data.value ?? '').trim())
			.filter(Boolean)
			.join('   •   ');
	}

	function skillText(item: Item): string {
		const name = String(item.data.name ?? '').trim();
		const level = String(item.data.level ?? '').trim();
		if (!name) return '';
		return level ? `${name} (${level})` : name;
	}

	function skillGroups(
		section: Section
	): { title: string; visible: boolean; skills: Item[] }[] {
		const groups: { title: string; visible: boolean; skills: Item[] }[] = [
			{ title: '', visible: true, skills: [] }
		];
		for (const item of section.items ?? []) {
			if (item.data.divider) {
				groups.push({
					title: String(item.data.title ?? '').trim(),
					visible: item.included,
					skills: []
				});
			} else {
				groups[groups.length - 1].skills.push(item);
			}
		}
		return groups;
	}

	function dateRange(item: Item): string {
		const start = String(item.data.start ?? '').trim();
		const end = String(item.data.end ?? '').trim();
		if (!start && !end) return '';
		if (!end) return `${start} – Present`;
		if (!start) return end;
		return `${start} – ${end}`;
	}

	function entryTitle(item: Item, kind: 'experience' | 'education' | 'custom'): string {
		if (kind === 'experience') {
			const role = String(item.data.role ?? '').trim();
			const company = String(item.data.company ?? '').trim();
			if (role && company) return `${role} — ${company}`;
			return role || company;
		}
		if (kind === 'education') {
			const degree = String(item.data.degree ?? '').trim();
			const field = String(item.data.field ?? '').trim();
			const school = String(item.data.school ?? '').trim();
			const left = [degree, field].filter(Boolean).join(' — ');
			if (left && school) return `${left}, ${school}`;
			return left || school;
		}
		return String(item.data.label ?? '').trim();
	}
</script>

{#snippet headerBlock()}
	<header class="resume-header">
		<h1 class="name">{resumeState?.profile.name || 'Your Name'}</h1>
		{#if resumeState?.profile.headline}
			<div class="headline">{resumeState.profile.headline}</div>
		{/if}
		{#each visibleSections as section (section.id)}
			{#if section.type === 'contact' && contactLine(section)}
				<div class="contact">{contactLine(section)}</div>
			{/if}
		{/each}
	</header>
{/snippet}

{#snippet sectionBlock(section: Section)}
	<section class="resume-section">
		<h2
			class="section-title"
			style:color={section.titleColor || null}
			style:border-bottom-color={section.lineColor || null}
		>
			{section.title}
		</h2>
		{#if section.type === 'skills'}
			<div class="skills-groups">
				{#each skillGroups(section) as group}
					{@const skills = group.skills
						.filter((i) => i.included)
						.map(skillText)
						.filter(Boolean)}
					{#if group.visible && (group.title || skills.length)}
						<span class="skills-group-title">{group.title}</span>
						<span class="skills">{skills.join(', ')}</span>
					{/if}
				{/each}
			</div>
		{:else}
			{#each included(section.items) as item (item.id)}
				{#if entryTitle(item, section.type as 'experience' | 'education' | 'custom') || dateRange(item)}
					<div class="entry-head">
						<span class="entry-title">{entryTitle(item, section.type as 'experience' | 'education' | 'custom')}</span>
						<span class="entry-dates">{dateRange(item)}</span>
					</div>
				{/if}
				{#if section.type === 'experience' && item.data.summary}
					<p class="summary">{item.data.summary}</p>
				{/if}
				{#if section.type === 'custom' && item.data.text}
					<p class="summary">{item.data.text}</p>
				{/if}
				{#if included(item.details ?? []).length}
					<ul class="bullets">
						{#each included(item.details ?? []) as detail (detail.id)}
							{#if detail.text.trim()}
								<li>{detail.text.trim()}</li>
							{/if}
						{/each}
					</ul>
				{/if}
			{/each}
		{/if}
	</section>
{/snippet}

<div class="preview-root">
	<div class="preview-backdrop">
		<!-- min-height rather than a fixed height: page breaks come from
			the backend (PDF metrics), but the browser renders text
			slightly taller, so the sheet grows instead of clipping. -->
		<div
			class="page"
			style:width="{page.widthMm}mm"
			style:min-height="{page.heightMm}mm"
			style:--font-section-title="{styles?.sectionTitle.size ?? 10.5}pt"
			style:--weight-section-title={styles?.sectionTitle.bold ? '700' : '400'}
			style:--slant-section-title={styles?.sectionTitle.italic ? 'italic' : 'normal'}
			style:--font-entry-title="{styles?.entryTitle.size ?? 10}pt"
			style:--weight-entry-title={styles?.entryTitle.bold ? '700' : '400'}
			style:--slant-entry-title={styles?.entryTitle.italic ? 'italic' : 'normal'}
			style:--font-entry-dates="{styles?.entryDates.size ?? 9}pt"
			style:--weight-entry-dates={styles?.entryDates.bold ? '700' : '400'}
			style:--slant-entry-dates={styles?.entryDates.italic ? 'italic' : 'normal'}
			style:--font-body="{styles?.body.size ?? 9.5}pt"
			style:--weight-body={styles?.body.bold ? '700' : '400'}
			style:--slant-body={styles?.body.italic ? 'italic' : 'normal'}
		>
			{#if currentPage === 0}
				{@render headerBlock()}
			{/if}
			{#each pages[currentPage] ?? [] as section (section.id)}
				{@render sectionBlock(section)}
			{/each}
		</div>
	</div>

	<div class="preview-footer">
		{#if pages.length > 1}
			<button
				class="icon-btn"
				disabled={currentPage === 0}
				title="Previous page"
				onclick={() => (currentPage -= 1)}
			>
				<Icon name="left" />
			</button>
			<span class="page-label">Page {currentPage + 1} of {pages.length}</span>
			<button
				class="icon-btn"
				disabled={currentPage >= pages.length - 1}
				title="Next page"
				onclick={() => (currentPage += 1)}
			>
				<Icon name="right" />
			</button>
		{:else}
			<span class="page-label" title="All visible sections fit on a single sheet">
				Fits on one {page.label} sheet
			</span>
		{/if}
	</div>
</div>

<style>
	.preview-root {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-height: 0;
	}

	.preview-backdrop {
		flex: 1;
		overflow: auto;
		padding: 24px;
	}

	.page {
		max-width: 100%;
		margin: 0 auto;
		background: #ffffff;
		color: #000000;
		/* Same font family and page margins as the PDF export
		   (Helvetica metrics, 12mm top/bottom, 14mm left/right). */
		font-family: Helvetica, 'Liberation Sans', Arial, sans-serif;
		padding: 12mm 14mm;
		/* Typography role variables; overridden inline from the saved
		   text styles. */
		--font-section-title: 10.5pt;
		--weight-section-title: 700;
		--slant-section-title: normal;
		--font-entry-title: 10pt;
		--weight-entry-title: 700;
		--slant-entry-title: normal;
		--font-entry-dates: 9pt;
		--weight-entry-dates: 400;
		--slant-entry-dates: normal;
		--font-body: 9.5pt;
		--weight-body: 400;
		--slant-body: normal;
		box-shadow: 0 2px 6px rgba(0, 0, 0, 0.18), 0 12px 36px rgba(0, 0, 0, 0.14);
		border-radius: 2px;
	}

	.preview-footer {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		padding: 6px;
		border-top: 1px solid var(--border);
		background: var(--card);
		flex-shrink: 0;
	}

	.page-label {
		font-size: 12.5px;
		color: var(--muted-foreground);
		min-width: 90px;
		text-align: center;
	}

	.resume-header {
		text-align: center;
		margin-bottom: 0;
	}

	.name {
		font-size: 18pt;
		font-weight: 700;
		line-height: 1;
	}

	.headline {
		font-size: 10.5pt;
		font-style: italic;
		color: #646970;
		margin-top: 1.5mm;
		line-height: 1.15;
	}

	.contact {
		font-size: 9.5pt;
		color: #646970;
		margin-top: 2mm;
		line-height: 1.15;
	}

	/* Section spacing mirrors the PDF: the title's Top padding of 3mm
	   separates sections, the line sits 1.8mm below the title glyphs and
	   content starts 1.2mm below the line. */
	.resume-section {
		margin-top: 3mm;
	}

	.section-title {
		font-size: var(--font-section-title);
		font-weight: var(--weight-section-title);
		font-style: var(--slant-section-title);
		text-transform: uppercase;
		line-height: 1;
		padding-bottom: 1.8mm;
		border-bottom: 0.25mm solid #646970;
		margin: 0;
	}

	.resume-section .section-title + * {
		margin-top: 2.2mm;
	}

	.resume-section .section-title + .entry-head {
		margin-top: 3.2mm;
	}

	/* Two-column layout: titles right-aligned in a shared column sized to
	   the widest title, skills left-aligned at a common edge. */
	.skills-groups {
		display: grid;
		grid-template-columns: max-content 1fr;
		column-gap: 1mm;
		row-gap: 1mm;
		font-size: var(--font-body);
		font-weight: var(--weight-body);
		font-style: var(--slant-body);
	}

	.skills-groups .skills-group-title {
		font-weight: 700;
		text-align: right;
		line-height: 1.15;
		font-style: normal;
	}

	.skills-groups .skills {
		line-height: 1.15;
		margin: 0;
	}

	.entry-head {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 6mm;
		margin-top: 2mm;
	}

	.entry-title {
		font-weight: var(--weight-entry-title);
		font-style: var(--slant-entry-title);
		font-size: var(--font-entry-title);
		line-height: 1;
	}

	.entry-dates {
		font-size: var(--font-entry-dates);
		font-weight: var(--weight-entry-dates);
		font-style: var(--slant-entry-dates);
		color: #646970;
		white-space: nowrap;
		line-height: 1;
	}

	.summary {
		margin: 1mm 0 0;
		font-size: var(--font-body);
		font-weight: var(--weight-body);
		font-style: var(--slant-body);
		color: #646970;
		line-height: 1.15;
	}

	.bullets {
		list-style: none;
		margin: 0.8mm 0 0;
		padding-left: 3mm;
		font-size: var(--font-body);
		font-weight: var(--weight-body);
		font-style: var(--slant-body);
	}

	.bullets li {
		line-height: 1.15;
	}

	.bullets li::before {
		content: '• ';
	}

	.bullets li + li {
		margin-top: 0.8mm;
	}
</style>
