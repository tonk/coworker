'use strict';

(function () {
  const toggle = document.querySelector('.navbar-toggle');
  const nav = document.querySelector('.navbar-nav');

  function closeNav() {
    if (!nav || !toggle) return;
    nav.classList.remove('open');
    toggle.setAttribute('aria-expanded', 'false');
  }

  function navLinks() {
    if (!nav) return [];
    return Array.from(nav.querySelectorAll(':scope > li > a[href]'));
  }

  // Mobile nav toggle
  if (toggle && nav) {
    toggle.addEventListener('click', function () {
      const open = nav.classList.toggle('open');
      toggle.setAttribute('aria-expanded', String(open));
    });

    document.addEventListener('click', function (e) {
      if (!toggle.contains(e.target) && !nav.contains(e.target)) {
        closeNav();
      }
    });

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && nav.classList.contains('open')) {
        e.preventDefault();
        closeNav();
        toggle.focus();
        return;
      }

      if (e.key !== 'Tab' || !nav.classList.contains('open')) return;

      const links = navLinks();
      if (links.length === 0) return;

      const last = links[links.length - 1];

      if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        toggle.focus();
      } else if (e.shiftKey && document.activeElement === toggle) {
        e.preventDefault();
        last.focus();
      }
    });
  }

  // Highlight active docs nav link
  document.querySelectorAll('.docs-nav a').forEach(function (a) {
    if (a.pathname === window.location.pathname) {
      a.classList.add('active');
    }
  });
})();
