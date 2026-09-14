/* payment.js — выбор способа оплаты (доступен только СБП) */

async function paymentPage() {
  const root = document.querySelector('[data-payment-root]');
  if (!root) return;

  const orderId = new URLSearchParams(location.search).get('order');
  if (!orderId) {
    root.innerHTML = `<div class="rounded-2xl border border-line bg-paper px-6 py-16 text-center"><p class="text-xl font-bold">Заказ не найден</p></div>`;
    return;
  }

  let order;
  try {
    const r = await fetch('/api/orders/' + orderId);
    if (!r.ok) throw new Error();
    order = await r.json();
  } catch {
    root.innerHTML = `<div class="rounded-2xl border border-line bg-paper px-6 py-16 text-center"><p class="text-xl font-bold">Заказ не найден</p></div>`;
    return;
  }

  root.innerHTML = `
    <div class="grid gap-8 lg:grid-cols-[1.6fr_1fr]">
      <div class="grid gap-4">
        <button data-method="sbp" class="flex items-center justify-between gap-4 rounded-2xl border-2 border-forest bg-paper p-6 text-left transition">
          <span>
            <span class="block text-lg font-bold">СБП — Система быстрых платежей</span>
            <span class="mt-1 block text-sm text-subtle">Оплата по QR-коду в приложении вашего банка. Без комиссии.</span>
          </span>
          <span class="rounded-full bg-forest px-3 py-1 text-xs font-bold text-paper">Доступно</span>
        </button>

        <button data-method="card" class="flex cursor-not-allowed items-center justify-between gap-4 rounded-2xl border border-line bg-paper p-6 text-left opacity-60">
          <span>
            <span class="block text-lg font-bold">Банковская карта</span>
            <span class="mt-1 block text-sm text-subtle">Visa, Mastercard, МИР — временно недоступно.</span>
          </span>
          <span class="rounded-full bg-mist px-3 py-1 text-xs font-bold text-subtle">Недоступно</span>
        </button>

        <div data-pay-box class="hidden rounded-2xl border border-line bg-paper p-6 text-center"></div>
      </div>

      <aside class="h-max rounded-2xl border border-line bg-paper p-6">
        <div class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Заказ №${order.id}</div>
        <div class="mt-4 grid gap-3 text-sm">
          ${(order.items || []).map(i => `<div class="flex justify-between gap-3">
            <span class="text-subtle">${escapeHtml(i.name)} × ${i.qty}</span>
            <span class="font-bold">${fmtPrice(i.lineTotal ?? (i.price * i.qty))}</span>
          </div>`).join('')}
        </div>
        <div class="mt-5 flex justify-between border-t border-line pt-4">
          <span class="font-bold">К оплате</span>
          <span class="text-2xl font-extrabold">${fmtPrice(order.total)}</span>
        </div>
      </aside>
    </div>`;

  const box = root.querySelector('[data-pay-box]');

  root.querySelector('[data-method="card"]').addEventListener('click', () => {
    toast('Оплата картой временно недоступна — выберите СБП');
  });

  root.querySelector('[data-method="sbp"]').addEventListener('click', async () => {
    box.classList.remove('hidden');
    box.innerHTML = '<p class="text-subtle">Формируем QR-код…</p>';

    const r = await fetch(`/api/orders/${orderId}/pay`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ method: 'sbp' }),
    });
    const data = await r.json();
    if (!data.ok) { box.innerHTML = `<p class="font-bold">${escapeHtml(data.error)}</p>`; return; }

    box.innerHTML = `
      <p class="text-lg font-bold">Отсканируйте QR-код в приложении банка</p>
      <img src="${escapeHtml(data.qr)}" alt="QR-код СБП" class="mx-auto mt-5 size-60 rounded-xl border border-line">
      <p class="mt-4 text-sm text-subtle">Сумма: ${fmtPrice(data.total)}</p>
      <button data-confirm class="mt-6 h-12 w-full rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">Я оплатил(а)</button>`;

    box.querySelector('[data-confirm]').addEventListener('click', async () => {
      await fetch(`/api/orders/${orderId}/confirm`, { method: 'POST' });
      box.innerHTML = `<p class="text-xl font-bold">Оплата принята 🎉</p>
        <p class="mt-2 text-sm text-subtle">Заказ №${order.id} оплачен. Мы свяжемся с вами для доставки.</p>
        <a href="/catalog" class="mt-6 inline-flex h-12 items-center justify-center rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">В каталог</a>`;
    });
  });
}
