const root = 'http://localhost:9000/api?';

function commonAPIRequest() {
    const r = new XMLHttpRequest();
    const parameters = document.getElementById('parameters').value;
    r.open('GET', root+parameters, true);
    r.onload = function () {
        if (this.status !== 200) {
            document.getElementById('error-text').textContent = this.responseText;
            document.getElementById('error-view').style.display = 'block';
        } else {
            document.getElementById('canvas').textContent = this.responseText;
            let err = document.getElementById('error-view');
            err.style.display = 'none';
        }
    };
    r.send();
}

document.getElementById('sync-btn').onclick = commonAPIRequest;
commonAPIRequest();