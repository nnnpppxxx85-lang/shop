/* main.js — что запускать на какой странице */

/* main.js — точка входа */

document.addEventListener('DOMContentLoaded', () => {
  captureReferral();
  header();
  footer();

  const page = document.body.dataset.page;

  if (page === 'home') {
    // категории на главной — 6 штук
    renderHomeCategories();
  }
  if (page === 'catalog')  catalogPage();
  if (page === 'item')     itemPage();
  if (page === 'cart')     cartPage();
  if (page === 'checkout') checkoutPage();
  if (page === 'payment')  paymentPage();
  if (page === 'admin')    adminPage();
  if (page === 'contacts') contactsPage();

  Cart.updateBadge();

  // мобильное меню
  const menuBtn = document.querySelector('[data-menu]');
  const menu    = document.querySelector('[data-mobile-menu]');
  menuBtn?.addEventListener('click', () => {
    const open = menu.classList.toggle('hidden');
    menuBtn.setAttribute('aria-expanded', String(!open));
    menuBtn.innerHTML = icon(open ? 'menu' : 'close');
  });

  document.querySelector('[data-toast-close]')?.addEventListener('click', () => {
    document.querySelector('[data-toast]')?.classList.add('translate-y-24', 'opacity-0');
  });

  // reveal-анимации — ВОТ ЭТО БЫЛО ПОТЕРЯНО
  initReveals();
});

function initReveals() {
  const obs = new IntersectionObserver((entries) => {
    entries.forEach((e) => {
      if (e.isIntersecting) {
        e.target.classList.add('is-visible');
        obs.unobserve(e.target);
      }
    });
  }, { threshold: 0.08 });

  document.querySelectorAll('.reveal:not(.is-visible)').forEach((el) => obs.observe(el));
}

async function renderHomeCategories() {
  const node = document.querySelector('[data-categories]');
  if (!node) return;

  const cats = await API.categories();
  const list = cats.slice(0, 6);

  node.innerHTML = list.map((c, i) => `
    <a href="/catalog?category=${c.slug}"
       class="group reveal relative flex min-h-56 flex-col justify-between overflow-hidden rounded-2xl border border-line bg-paper p-6 shadow-sm transition duration-300 hover:-translate-y-1 hover:border-forest/25 hover:shadow-soft"
       style="transition-delay:${Math.min(i,5)*45}ms">
      <span class="text-xs font-bold text-forest/55">${c.code}</span>
      <div>
        <h3 class="text-2xl font-bold">${escapeHtml(c.name)}</h3>
        <p class="mt-2 text-sm text-subtle">${escapeHtml(c.note)}</p>
      </div>
      <span class="absolute right-5 top-5 flex size-9 items-center justify-center rounded-full bg-mist text-forest transition group-hover:bg-forest group-hover:text-paper">${icon('arrow')}</span>
    </a>`).join('');

  // наблюдаем за вновь добавленными .reveal
  initReveals();
}