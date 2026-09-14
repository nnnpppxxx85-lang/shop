function header() {
  const node = document.querySelector('[data-site-header]');
  if (!node) return;
  const page = document.body.dataset.page;

  node.innerHTML = `
    <div class="bg-forest px-5 py-2 text-center text-[11px] font-semibold tracking-[.12em] text-paper uppercase">Скидки участникам СВО</div>
    <header class="sticky top-0 z-40 border-b border-line/70 bg-canvas/90 backdrop-blur-xl">
      <div class="mx-auto flex h-18 max-w-7xl items-center justify-between px-5 lg:px-8">
        <a href="/" class="flex items-center gap-3">
          <span class="text-[17px] font-extrabold">DragonMobile</span>
        </a>
        <nav class="hidden items-center gap-8 text-[13px] font-semibold lg:flex">
          <a class="${page==='home'?'text-forest':'text-subtle hover:text-ink'} transition" href="/">Главная</a>
          <a class="${page==='catalog'||page==='item'?'text-forest':'text-subtle hover:text-ink'} transition" href="/catalog">Каталог</a>
          <a class="${page==='delivery'?'text-forest':'text-subtle hover:text-ink'} transition" href="/delivery">Доставка и оплата</a>
          <a class="${page==='contacts'?'text-forest':'text-subtle hover:text-ink'} transition" href="/contacts">Контакты</a>
        </nav>
        <div class="flex items-center gap-1">
          <a href="/catalog#search" class="flex size-10 items-center justify-center rounded-full text-ink transition hover:bg-warm">${icon('search')}</a>
          <a href="/cart" class="relative flex size-10 items-center justify-center rounded-full text-ink transition hover:bg-warm" aria-label="Корзина">${icon('bag')}<span data-cart-count class="absolute -right-0.5 -top-0.5 hidden min-w-5 rounded-full bg-forest px-1 text-center text-[11px] font-bold leading-5 text-paper">0</span></a>
          <button class="flex size-10 items-center justify-center rounded-full text-ink transition hover:bg-warm lg:hidden" data-menu aria-expanded="false">${icon('menu')}</button>
        </div>
      </div>
      <div class="hidden border-t border-line bg-paper px-5 py-5 lg:hidden" data-mobile-menu>
        <nav class="grid gap-1 text-lg font-semibold">
          <a class="rounded-lg px-3 py-3 hover:bg-mist" href="/">Главная</a>
          <a class="rounded-lg px-3 py-3 hover:bg-mist" href="/catalog">Каталог</a>
          <a class="rounded-lg px-3 py-3 hover:bg-mist" href="/delivery">Доставка и оплата</a>
          <a class="rounded-lg px-3 py-3 hover:bg-mist" href="/contacts">Контакты</a>
          <a class="rounded-lg px-3 py-3 hover:bg-mist" href="/cart">Корзина</a>
        </nav>
      </div>
    </header>`;
}

function footer() {
  const node = document.querySelector('[data-site-footer]');
  if (!node) return;

  node.innerHTML = `
    <footer class="border-t border-line bg-ink text-paper">
      <div class="mx-auto grid max-w-7xl gap-10 px-5 py-12 md:grid-cols-[1.4fr_1fr_1fr] lg:px-8 lg:py-16">
        <div>
          <div class="text-xl font-extrabold">DragonMobile</div>
          <p class="mt-4 max-w-sm text-sm leading-6 text-paper/60">Оригинальная техника с вниманием к каждой детали.</p>
        </div>
        <div>
          <div class="text-xs font-bold uppercase tracking-[.12em] text-paper/45">Навигация</div>
          <div class="mt-4 grid gap-3 text-sm">
            <a href="/catalog">Каталог</a><a href="/delivery">Доставка и оплата</a><a href="/contacts">Контакты</a>
          </div>
        </div>
        <div>
          <div class="text-xs font-bold uppercase tracking-[.12em] text-paper/45">Покупателям</div>
          <div class="mt-4 grid gap-3 text-sm">
            <a href="/delivery#payment">Способы оплаты</a><a href="/delivery#warranty">Гарантия</a><a href="/contacts">Связаться с нами</a>
          </div>
        </div>
      </div>
      <div class="mx-auto flex max-w-7xl flex-col gap-2 border-t border-paper/10 px-5 py-5 text-xs text-paper/45 sm:flex-row sm:justify-between lg:px-8">
        <span>© 2026 DragonMobile</span><span>Москва</span>
      </div>
    </footer>`;
}
