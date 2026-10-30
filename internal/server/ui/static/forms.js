// Forms show only the fields that apply. An element with data-show is shown
// while its conditions hold:
//
//   data-show="role=server,spoke"              the field role is server or spoke
//   data-show="ip_family!=ipv4"                the field is none of these
//   data-show="role=server;mode=host"          every condition holds
//   data-show="containers=on|mode=container"   either alternative holds
//
// A ticked checkbox's value is "on", an unticked one's ""; a field that is
// missing, hidden or disabled counts as "". The fields inside a hidden
// element are disabled, so the browser neither checks nor sends them; showing
// it again enables them. Without this script every field shows, and the
// server checks the form as always.
(function () {
  "use strict";

  const fields = "input, select, textarea";

  function valueOf(form, name) {
    const el = form.elements.namedItem(name);
    if (!el || el.disabled) {
      return "";
    }
    if (el.type === "checkbox") {
      return el.checked ? "on" : "";
    }
    return el.value;
  }

  function holds(form, cond) {
    const not = cond.includes("!=");
    const [name, values] = cond.split(not ? "!=" : "=", 2);
    const listed = (values || "").split(",").map((v) => v.trim()).includes(valueOf(form, name.trim()));
    return not ? !listed : listed;
  }

  function wanted(form, spec) {
    return spec.split("|").some((alt) => alt.split(";").every((cond) => holds(form, cond)));
  }

  // update settles a form: showing or hiding one element can decide
  // another's, and a field is off while any element around it is hidden.
  function update(form) {
    const els = form.querySelectorAll("[data-show]");
    for (let pass = 0; pass < 5; pass++) {
      let changed = false;
      for (const el of els) {
        const on = wanted(form, el.dataset.show);
        if (el.hidden === on) {
          el.hidden = !on;
          changed = true;
        }
      }
      for (const f of form.querySelectorAll(fields)) {
        const off = f.closest("[hidden]") !== null;
        if (off && !f.disabled) {
          f.disabled = true;
          f.dataset.shownOff = "";
          changed = true;
        } else if (!off && "shownOff" in f.dataset) {
          f.disabled = false;
          delete f.dataset.shownOff;
          changed = true;
        }
      }
      if (!changed) {
        return;
      }
    }
  }

  function updateAll(root) {
    for (const form of root.querySelectorAll("form")) {
      if (form.querySelector("[data-show]")) {
        update(form);
      }
    }
  }

  document.addEventListener("change", (e) => {
    if (e.target.form) {
      update(e.target.form);
    }
  });
  document.addEventListener("DOMContentLoaded", () => updateAll(document));
  document.addEventListener("htmx:afterSettle", (e) => updateAll(e.target));
})();
