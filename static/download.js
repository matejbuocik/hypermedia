window.ws = null;

function onClickDownload() {
    if (window.ws) {
        return false;
    }

    const host = document.querySelector("[data-download]").attributes.getNamedItem("data-download").value;
    window.ws = new WebSocket(`ws://${host}/contacts/download`);
    window.ws.onopen = () => {
        const download = document.querySelector("[data-download]");
        if (download) download.innerHTML = progressBox;
    }
    window.ws.onmessage = (evt) => {
        const progress = document.getElementById("progressBar");
        if (progress) progress.style.width = `${evt.data}%`;
    }
    window.ws.onclose = (evt) => {
        window.ws = null;
        const download = document.querySelector("[data-download]")

        if (evt.code !== 1000) {
            if (download) download.innerHTML = `<a href="/contacts">Error, please refresh.</a>`;
            return;
        }

        window.location = `/contacts/file?uuid=${evt.reason}`;
        if (download) download.innerHTML = downloadButton;
    }

    return false;
}

const downloadButton = `<button onclick="onClickDownload()">Download Contacts</button>`;
const progressBox = `
    <div role="progressbar" style="
    width: 15rem;
    height: 1.5rem;
    background-color: var(--bg-hl);
    border-radius: 5px;
    box-shadow: inset 0 1px 2px var(--bg);"
    >
        <div id="progressBar" style="
            width: 0%;
            height: 100%;
            background-color: var(--link);
            border-radius: 5px;
            box-shadow: inset 0 -1px 0 rgba(0,0,0,.15);
            transition: width .6s ease;">
        </div>
    </div>`;
