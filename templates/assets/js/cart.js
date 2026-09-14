/* cart.js — корзина в localStorage + страница корзины */

const CART_KEY = 'dm_cart';

const Cart = {
  read() {
    try { return JSON.parse(localStorage.getItem(CART_KEY)) || []; }
    catch { return []; }
  },
  write(items) {
    localStorage.setItem(CART_KEY, JSON.stringify(items));
    Cart.updateBadge();
  },
  add(product, qty = 1) {
    const items = Cart.read();
    const found = items.find(i => i.id === product.id);
    if (found) found.qty += qty;
    else items.push({
      id: product.id,
      slug: product.slug,
      name: product.name,
      variant: product.variant || '',
      price: product.finalPrice ?? product.price,
      image: (product.images && product.images[0]) || '',
      discountType: product.discountType || 'none',
      bundleBuyQty: product.bundleBuyQty || 0,
      bundleTotalQty: product.bundleTotalQty || 0,
      qty,
    });
    Cart.write(items);
  },
  setQty(id, qty) {
    const items = Cart.read();
    const it = items.find(i => i.id === id);
    if (!it) return;
    it.qty = Math.max(1, qty);
    Cart.write(items);
  },
  remove(id) {
    Cart.write(Cart.read().filter(i => i.id !== id));
  },
  clear() { Cart.write([]); },
  count() { return Cart.read().reduce((s, i) => s + i.qty, 0); },
  total() { return Cart.read().reduce((s, i) => s + lineTotal(i), 0); },
  updateBadge() {
    const badge = document.querySelector('[data-cart-count]');
    if (!badge) return;
    const n = Cart.count();
    badge.textContent = n;
    badge.classList.toggle('hidden', n === 0);
  },
};

function toast(text) {
  const t = document.querySelector('[data-toast]');
  if (!t) return;
  const span = t.querySelector('span');
  if (span) span.textContent = text;
  t.classList.remove('translate-y-24', 'opacity-0');
  clearTimeout(window.__toastTimer);
  window.__toastTimer = setTimeout(() => t.classList.add('translate-y-24', 'opacity-0'), 2600);
}

/* ---------- страница корзины ---------- */

function cartPage() {
  const root = document.querySelector('[data-cart-root]');
  if (!root) return;

  function render() {
    const items = Cart.read();

    if (!items.length) {
      root.innerHTML = `
        <div class="rounded-2xl border border-line bg-paper px-6 py-16 text-center">
          <p class="text-xl font-bold">Корзина пуста</p>
          <p class="mt-2 text-sm text-subtle">Загляните в каталог — там много интересного.</p>
          <a href="/catalog" class="mt-6 inline-flex h-12 items-center justify-center rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">В каталог</a>
        </div>`;
      return;
    }

    root.innerHTML = `
      <div class="grid gap-8 lg:grid-cols-[1.6fr_1fr]">
        <div class="grid gap-4">
          ${items.map(i => `
            <div class="flex gap-4 rounded-2xl border border-line bg-paper p-4">
              <div class="size-24 shrink-0 overflow-hidden rounded-xl bg-warm">
                ${i.image ? `<img src="${escapeHtml(i.image)}" alt="${escapeHtml(i.name)}" class="size-full object-cover">` : ''}
              </div>
              <div class="flex flex-1 flex-col gap-2">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <a href="/item?slug=${encodeURIComponent(i.slug)}" class="text-base font-bold leading-snug hover:text-forest">${escapeHtml(i.name)}</a>
                    ${i.variant ? `<p class="mt-1 text-xs text-subtle">${escapeHtml(i.variant)}</p>` : ''}
                  </div>
                  <button data-remove="${i.id}" class="text-sm text-subtle transition hover:text-ink">×</button>
                </div>
                ${i.discountType === 'bundle' && i.bundleTotalQty ? `<p class="text-xs font-semibold text-forest">Акция: ${i.bundleBuyQty}+${i.bundleTotalQty - i.bundleBuyQty}=${i.bundleTotalQty}</p>` : ''}
                <div class="mt-auto flex items-center justify-between gap-3">
                  <div class="flex items-center gap-2">
                    <button data-dec="${i.id}" class="flex size-8 items-center justify-center rounded-full border border-line bg-paper font-bold">−</button>
                    <span class="w-8 text-center text-sm font-bold">${i.qty}</span>
                    <button data-inc="${i.id}" class="flex size-8 items-center justify-center rounded-full border border-line bg-paper font-bold">+</button>
                  </div>
                  <div class="text-lg font-extrabold">${fmtPrice(lineTotal(i))}</div>
                </div>
              </div>
            </div>`).join('')}
        </div>

        <aside class="h-max rounded-2xl border border-line bg-paper p-6">
          <div class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Итого</div>
          <div class="mt-3 text-3xl font-extrabold">${fmtPrice(Cart.total())}</div>
          <p class="mt-2 text-sm text-subtle">${Cart.count()} товар(ов)</p>
          <a href="/checkout" class="mt-6 flex h-12 items-center justify-center rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">Оформить заказ</a>
          <button data-clear class="mt-3 h-12 w-full rounded-full border border-line bg-paper px-6 text-sm font-bold transition hover:border-forest/30">Очистить корзину</button>
        </aside>
      </div>`;
  }

  root.addEventListener('click', (e) => {
    const t = e.target.closest('[data-remove],[data-inc],[data-dec],[data-clear]');
    if (!t) return;
    if (t.dataset.clear !== undefined) Cart.clear();
    else if (t.dataset.remove) Cart.remove(+t.dataset.remove);
    else if (t.dataset.inc) Cart.setQty(+t.dataset.inc, (Cart.read().find(i => i.id === +t.dataset.inc)?.qty || 1) + 1);
    else if (t.dataset.dec) Cart.setQty(+t.dataset.dec, (Cart.read().find(i => i.id === +t.dataset.dec)?.qty || 1) - 1);
    render();
  });

  render();
}
