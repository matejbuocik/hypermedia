function contactMenu(tree = document) {
    tree.querySelectorAll("[data-contact-menu]").forEach(menuRoot => {
        const
            button = menuRoot.querySelector("[aria-haspopup]"),
            menu = menuRoot.querySelector("[role=menu]"),
            items = [...menu.querySelectorAll("[role=menuitem]")];

        const isOpen = () => menu.style.display !== "none";
        items.forEach(item => item.setAttribute("tabindex", "-1"));

        function toggleMenu(open = !isOpen()) {
            if (open) {
                menu.style.display = "flex";
                button.setAttribute("aria-expanded", "true");
                button.style.color = "var(--fg)";
                button.style.textDecoration = "none";
                items[0].focus();
            } else {
                menu.style.display = "none";
                button.style.color = "var(--link)";
                button.style.textDecoration = "underline";
                button.setAttribute("aria-expanded", "false");
            }
        }

        toggleMenu(isOpen());
        button.addEventListener("click", () => toggleMenu());
        menuRoot.addEventListener("blur", e => toggleMenu(false));

        window.addEventListener("click", function clickAway(event) {
            if (!menuRoot.isConnected)
                window.removeEventListener("click", clickAway);
            if (!menuRoot.contains(event.target)) toggleMenu(false);
        });
    });
}

addEventListener("htmx:load", e => contactMenu(e.target));
