/* item.js — страница товара */

async function itemPage() {
  const root = document.querySelector('[data-item-root]');
  const crumb = document.querySelector('[data-breadcrumbs]');
  if (!root) return;

  const slug = new URLSearchParams(location.search).get('slug');
  if (!slug) { root.innerHTML = '<p class="col-span-full text-subtle">Товар не выбран.</p>'; return; }

  let p;
  try { p = await API.product(slug); }
  catch { root.innerHTML = '<p class="col-span-full text-subtle">Товар не найден.</p>'; return; }

  const settings = await API.settings().catch(() => ({}));
  const consultantUrl = settings.consultantTelegram || '#';

  document.title = `${p.name} — DragonMobile`;

  const images = p.images || [];
  const badge = discountBadge(p);
  const hint = discountHint(p);
  const oldPrice = p.discountType === 'percent' && p.discountPercent > 0
    ? p.price
    : (p.marketPrice && p.marketPrice > p.finalPrice ? p.marketPrice : null);

  const storageOptions = p.storageOptions || [];
  let storageIdx = 0;

  crumb.innerHTML = `
    <a href="/catalog" class="hover:text-ink">Каталог</a>
    <span class="mx-2 text-line">/</span>
    <a href="/catalog?category=${encodeURIComponent(p.categorySlug)}" class="hover:text-ink">${escapeHtml(p.categoryName)}</a>
    <span class="mx-2 text-line">/</span>
    <span class="text-ink">${escapeHtml(p.name)}</span>`;

  root.innerHTML = `
    <!-- Слайдер -->
    <div class="flex flex-col gap-4">
      <div class="relative aspect-square overflow-hidden rounded-2xl border border-line bg-warm" data-slider>
        <div class="flex h-full transition-transform duration-300 ease-out" data-track>
          ${images.length
            ? images.map((src, i) => `<img src="${escapeHtml(src)}" alt="${escapeHtml(p.name)}" data-slide="${i}" class="h-full w-full shrink-0 object-contain p-6">`).join('')
            : '<div class="flex h-full w-full items-center justify-center text-subtle">Нет фото</div>'}
        </div>

        ${badge ? `<span class="absolute left-5 top-5 rounded-full bg-forest px-3 py-1 text-xs font-bold text-paper">${badge}</span>` : ''}

        ${images.length > 1 ? `
          <button data-prev class="absolute left-4 top-1/2 -translate-y-1/2 flex size-10 items-center justify-center rounded-full bg-paper/90 text-ink shadow-soft transition hover:bg-paper">${icon('left')}</button>
          <button data-next class="absolute right-4 top-1/2 -translate-y-1/2 flex size-10 items-center justify-center rounded-full bg-paper/90 text-ink shadow-soft transition hover:bg-paper">${icon('right')}</button>
          <span data-counter class="absolute bottom-4 right-4 rounded-full bg-ink/70 px-3 py-1 text-xs font-semibold text-paper">1 / ${images.length}</span>
        ` : ''}
      </div>

      ${images.length > 1 ? `
        <div class="flex gap-2 overflow-x-auto pb-1" data-thumbs>
          ${images.map((src, i) => `
            <button data-thumb="${i}" class="shrink-0 overflow-hidden rounded-xl border ${i===0?'border-forest':'border-line'} bg-paper transition hover:border-forest/40">
              <img src="${escapeHtml(src)}" alt="" loading="lazy" class="size-20 object-contain p-1">
            </button>`).join('')}
        </div>` : ''}
    </div>

    <!-- Инфо -->
    <div class="flex flex-col gap-6">
      <div>
        <p class="text-xs font-bold uppercase tracking-[.15em] text-forest">${escapeHtml(p.categoryName)}</p>
        <h1 class="mt-3 text-3xl font-extrabold leading-tight sm:text-4xl">${escapeHtml(p.name)}</h1>
        ${p.variant ? `<p class="mt-3 text-sm text-subtle">${escapeHtml(p.variant)}</p>` : ''}
      </div>

      <div class="rounded-2xl border border-line bg-paper p-6">
        ${storageOptions.length ? `
        <div class="mb-5 grid gap-2">
          <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Объём памяти</span>
          <div class="flex flex-wrap gap-2" data-storage-group>
            ${storageOptions.map((o, i) => `
              <button type="button" data-storage-idx="${i}" class="rounded-full border px-4 py-2 text-sm font-semibold transition ${i === 0 ? 'border-forest bg-forest text-paper' : 'border-line bg-paper text-subtle hover:border-forest/30 hover:text-ink'}">${escapeHtml(o.label)}${o.priceDelta ? ` +${fmtPrice(o.priceDelta)}` : ''}</button>`).join('')}
          </div>
        </div>` : ''}

        <div class="flex items-end justify-between gap-4">
          <div>
            ${oldPrice ? `<div data-price-old class="text-sm text-subtle line-through">${fmtPrice(oldPrice)}</div>` : ''}
            <div data-price-current class="mt-1 text-3xl font-extrabold">${fmtPrice(p.finalPrice)}</div>
          </div>
          ${badge ? `<div class="rounded-full bg-forest/10 px-3 py-1 text-sm font-bold text-forest">${badge}</div>` : ''}
        </div>

        ${hint ? `<p class="mt-4 text-sm font-semibold text-forest">${escapeHtml(hint)}</p>` : ''}

        <div class="mt-6 grid gap-3">
          <button class="h-12 rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark" data-buy>В корзину</button>
          <a href="${escapeHtml(consultantUrl)}" target="_blank" rel="noopener" class="flex h-12 items-center justify-center rounded-full border border-line bg-paper px-6 text-sm font-bold transition hover:border-forest/30">Спросить консультанта</a>
        </div>
      </div>

      <ul class="grid gap-2 text-sm text-subtle">
        <li>• Оригинальная техника, проверяем каждое устройство</li>
        <li>• Доставка по Москве — 1 день</li>
        <li>• Гарантия 365 дней</li>
      </ul>
    </div>`;

  const priceCurrent = root.querySelector('[data-price-current]');
  const priceOld = root.querySelector('[data-price-old]');

  function updatePrice() {
    const delta = storageOptions[storageIdx]?.priceDelta || 0;
    priceCurrent.textContent = fmtPrice(p.finalPrice + delta);
    if (priceOld && oldPrice != null) priceOld.textContent = fmtPrice(oldPrice + delta);
  }

  root.querySelector('[data-storage-group]')?.addEventListener('click', (e) => {
    const btn = e.target.closest('[data-storage-idx]');
    if (!btn) return;
    storageIdx = +btn.dataset.storageIdx;
    root.querySelectorAll('[data-storage-idx]').forEach((b, i) => {
      const active = i === storageIdx;
      b.classList.toggle('border-forest', active);
      b.classList.toggle('bg-forest', active);
      b.classList.toggle('text-paper', active);
      b.classList.toggle('border-line', !active);
      b.classList.toggle('bg-paper', !active);
      b.classList.toggle('text-subtle', !active);
    });
    updatePrice();
  });

  root.querySelector('[data-buy]')?.addEventListener('click', () => {
    Cart.add(p, 1, storageOptions[storageIdx] || null);
    toast('Товар добавлен в корзину');
  });

  initSlider(root, images.length);
}

/* ---- слайдер ---- */

function initSlider(root, total) {
  const track   = root.querySelector('[data-track]');
  const prev    = root.querySelector('[data-prev]');
  const next    = root.querySelector('[data-next]');
  const counter = root.querySelector('[data-counter]');
  const thumbs  = root.querySelector('[data-thumbs]');
  if (!track || total < 2) return;

  let index = 0;

  function go(i) {
    index = (i + total) % total;
    track.style.transform = `translateX(-${index * 100}%)`;
    if (counter) counter.textContent = `${index + 1} / ${total}`;
    thumbs?.querySelectorAll('[data-thumb]').forEach((b, k) => {
      b.classList.toggle('border-forest', k === index);
      b.classList.toggle('border-line', k !== index);
    });
  }

  prev?.addEventListener('click', () => go(index - 1));
  next?.addEventListener('click', () => go(index + 1));

  thumbs?.addEventListener('click', e => {
    const b = e.target.closest('[data-thumb]');
    if (b) go(+b.dataset.thumb);
  });

  // свайп на мобиле
  let x0 = null;
  track.addEventListener('touchstart', e => { x0 = e.touches[0].clientX; });
  track.addEventListener('touchend', e => {
    if (x0 === null) return;
    const dx = e.changedTouches[0].clientX - x0;
    if (Math.abs(dx) > 40) go(dx > 0 ? index - 1 : index + 1);
    x0 = null;
  });

  // стрелки на клавиатуре
  document.addEventListener('keydown', e => {
    if (e.key === 'ArrowLeft')  go(index - 1);
    if (e.key === 'ArrowRight') go(index + 1);
  });
}