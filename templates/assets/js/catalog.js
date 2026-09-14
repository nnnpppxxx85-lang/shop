/* catalog.js — страница каталога */

async function catalogPage() {
  const grid = document.querySelector('[data-catalog-grid]');
  if (!grid) return;

  const search = document.querySelector('[data-category-search]');
  const pills  = document.querySelector('[data-filter-pills]');

  const cats = await API.categories();
  let selected = new URLSearchParams(location.search).get('category') || 'all';

  pills.innerHTML = [{ slug: 'all', name: 'Все' }, ...cats].map(c => `
    <button class="whitespace-nowrap rounded-full border px-4 py-2 text-sm font-semibold transition ${
      c.slug === selected
        ? 'border-forest bg-forest text-paper'
        : 'border-line bg-paper text-subtle hover:border-forest/30 hover:text-ink'
    }" data-filter="${c.slug}">${escapeHtml(c.name)}</button>`).join('');

  async function render(q = '') {
    q = q.trim().toLowerCase();

    grid.innerHTML = skeletonGrid();
    const products = await API.catalog(selected === 'all' ? '' : selected);
    const list = products.filter(p => `${p.name} ${p.variant || ''}`.toLowerCase().includes(q));

    grid.innerHTML = list.length ? list.map(productCard).join('') : emptyState('Ничего не найдено');
  }

  function skeletonGrid() {
    return Array.from({ length: 6 }).map(() => `
      <div class="animate-pulse overflow-hidden rounded-2xl border border-line bg-paper">
        <div class="aspect-square bg-warm"></div>
        <div class="grid gap-3 p-5">
          <div class="h-4 w-3/4 rounded bg-warm"></div>
          <div class="h-4 w-1/3 rounded bg-warm"></div>
        </div>
      </div>`).join('');
  }

  function productCard(p) {
    const badge = discountBadge(p);
    const oldPrice = p.marketPrice && p.marketPrice > p.finalPrice
      ? p.marketPrice
      : (p.discountType === 'percent' && p.discountPercent > 0 ? p.price : null);
    const img = p.images[0] || '';

    return `
      <a href="/item?slug=${encodeURIComponent(p.slug)}"
         class="group relative flex flex-col overflow-hidden rounded-2xl border border-line bg-paper transition hover:-translate-y-1 hover:border-forest/30 hover:shadow-soft">

        <div class="relative aspect-square overflow-hidden bg-warm">
          ${img ? `<img src="${escapeHtml(img)}" alt="${escapeHtml(p.name)}" loading="lazy" class="size-full object-cover transition duration-500 group-hover:scale-105">` : ''}
          ${badge ? `<span class="absolute left-4 top-4 rounded-full bg-forest px-3 py-1 text-xs font-bold text-paper">${badge}</span>` : ''}
        </div>

        <div class="flex flex-1 flex-col gap-3 p-5">
          <div>
            <h3 class="text-lg font-bold leading-snug">${escapeHtml(p.name)}</h3>
            ${p.variant ? `<p class="mt-1 text-xs text-subtle">${escapeHtml(p.variant)}</p>` : ''}
          </div>

          <div class="mt-auto flex items-end justify-between gap-3 border-t border-line pt-4">
            <div>
              ${oldPrice ? `<div class="text-xs text-subtle line-through">${fmtPrice(oldPrice)}</div>` : ''}
              <div class="text-xl font-extrabold">${fmtPrice(p.finalPrice)}</div>
            </div>
            <span class="flex size-9 items-center justify-center rounded-full bg-mist text-forest transition group-hover:bg-forest group-hover:text-paper">${icon('arrow')}</span>
          </div>
        </div>
      </a>`;
  }

  function emptyState(title) {
    return `<div class="col-span-full rounded-2xl border border-line bg-paper px-6 py-16 text-center"><p class="text-xl font-bold">${escapeHtml(title)}</p></div>`;
  }

  pills.addEventListener('click', e => {
    const b = e.target.closest('[data-filter]');
    if (!b) return;
    selected = b.dataset.filter;
    pills.querySelectorAll('button').forEach(x => {
      x.classList.toggle('border-forest', x === b);
      x.classList.toggle('bg-forest', x === b);
      x.classList.toggle('text-paper', x === b);
      x.classList.toggle('border-line', x !== b);
      x.classList.toggle('bg-paper', x !== b);
      x.classList.toggle('text-subtle', x !== b);
    });
    const url = new URL(location.href);
    selected === 'all' ? url.searchParams.delete('category') : url.searchParams.set('category', selected);
    history.replaceState(null, '', url);
    render(search.value);
  });

  search.addEventListener('input', () => render(search.value));
  render();
}