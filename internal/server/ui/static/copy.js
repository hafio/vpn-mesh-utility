// Copy buttons. Every <pre>, every <code class="secret"> (a join token) and
// every element with class "copy" gets a button that puts its text on the
// clipboard. Where the browser has no clipboard for the page (plain http
// other than localhost), the button selects the text for a key press to
// copy instead. Without this script the text is there to select as always.
(function () {
  "use strict";

  const targets = "pre, code.secret, .copy";

  function done(button, label) {
    button.textContent = label;
    setTimeout(() => {
      button.textContent = "Copy";
    }, 2000);
  }

  function select(el) {
    const range = document.createRange();
    range.selectNodeContents(el);
    const sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
  }

  function copy(el, button) {
    // No trailing newline: a pasted command waits for its Enter.
    const text = el.textContent.replace(/\s+$/, "");
    if (!navigator.clipboard) {
      select(el);
      done(button, "Press Ctrl+C");
      return;
    }
    navigator.clipboard.writeText(text).then(
      () => done(button, "Copied"),
      () => {
        select(el);
        done(button, "Press Ctrl+C");
      },
    );
  }

  function add(root) {
    for (const el of root.querySelectorAll(targets)) {
      if (el.closest(".copyable")) {
        continue;
      }
      const wrap = document.createElement(el.tagName === "PRE" ? "div" : "span");
      wrap.className = "copyable";
      const button = document.createElement("button");
      button.type = "button";
      button.className = "copy-button";
      button.textContent = "Copy";
      button.addEventListener("click", () => copy(el, button));
      el.replaceWith(wrap);
      wrap.append(el, button);
    }
  }

  document.addEventListener("DOMContentLoaded", () => add(document.body));
  document.addEventListener("htmx:afterSettle", (e) => add(e.target));
})();
