'use strict';

(function () {
  // Mobile nav toggle
  const toggle = document.querySelector('.navbar-toggle');
  const nav    = document.querySelector('.navbar-nav');
  if (toggle && nav) {
    toggle.addEventListener('click', function () {
      const open = nav.classList.toggle('open');
      toggle.setAttribute('aria-expanded', String(open));
    });
    document.addEventListener('click', function (e) {
      if (!toggle.contains(e.target) && !nav.contains(e.target)) {
        nav.classList.remove('open');
        toggle.setAttribute('aria-expanded', 'false');
      }
    });
  }

  // Highlight active docs nav link
  const links = document.querySelectorAll('.docs-nav a');
  links.forEach(function (a) {
    if (a.pathname === window.location.pathname) {
      a.classList.add('active');
    }
  });
})();
