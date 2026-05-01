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

  // Screenshots: click thumbnail for full-screen overlay; click overlay or image to dismiss (Escape too)
  const shotsSection = document.getElementById('screenshots');
  if (shotsSection) {
    const overlay = document.createElement('div');
    overlay.className = 'screenshot-popout-overlay';
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-modal', 'true');
    overlay.setAttribute('aria-label', 'Enlarged screenshot');
    const popImg = document.createElement('img');
    overlay.appendChild(popImg);
    document.body.appendChild(overlay);

    var prevOverflow = '';

    function closePopout() {
      overlay.classList.remove('is-open');
      document.body.style.overflow = prevOverflow;
      prevOverflow = '';
    }

    function openPopout(thumb) {
      popImg.src = thumb.currentSrc || thumb.src;
      popImg.alt = thumb.alt || '';
      prevOverflow = document.body.style.overflow;
      document.body.style.overflow = 'hidden';
      overlay.classList.add('is-open');
    }

    shotsSection.querySelectorAll('.screenshot-item img').forEach(function (thumb) {
      thumb.addEventListener('click', function (e) {
        e.preventDefault();
        openPopout(thumb);
      });
    });

    overlay.addEventListener('click', function () {
      closePopout();
    });

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && overlay.classList.contains('is-open')) {
        closePopout();
      }
    });
  }
})();
