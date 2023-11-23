const express = require('express');
const jwt = require('jsonwebtoken');
const app = express();
const port = 3000;

// Function to convert JSON to HTML
function jsonToHtml(json) {
    return Object.entries(json).map(([key, value]) => {
        return `<span class="key">${key}</span>: <span class="value">${value}</span>`;
    }).join(',<br>');
}

app.get('*', (req, res) => {
    const hostHeader = req.headers.host;
    const uriPath = req.originalUrl;
    const secretKey = 'your-256-bit-secret';
    const token = jwt.sign({ user: 'test' }, secretKey);

    // Decode the JWT
    const decoded = jwt.verify(token, secretKey);

    // HTML for displaying the variables
    const html = `
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; background-color: white; color: black; }
            .banner { background-color: green; color: white; padding: 10px; }
            .main { text-align: center; }
            .main h1 { font-size: 48px; font-weight: bold; }
            .jwt { background-color: lightgrey; color: black; padding: 10px; margin: 20px; }
            .key { color: blue; font-weight: bold; }
            .value { color: darkgreen; }
        </style>
    </head>
    <body>
        <div class="banner">
            Host header: ${hostHeader}<br>
            URI Path: ${uriPath}<br>
            Authorization: Bearer ${token}
        </div>
        <div class="jwt">
            <strong>Decoded JWT:</strong><br>
            {<br>
            ${jsonToHtml(decoded)}<br>
            }
        </div>
        <div class="main">
            <h1 id="jobTitle">Loading...</h1>
        </div>
        <script>
            fetch('/get-job')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('jobTitle').innerText = data.job;
                })
                .catch(error => {
                    console.error('Error fetching job title:', error);
                    document.getElementById('jobTitle').innerText = 'Error fetching job title';
                });
        </script>
    </body>
    </html>`;

    res.send(html);
});

app.listen(port, () => {
    console.log(`Server running at http://localhost:${port}`);
});

