// Loads a page using Puppeteer and saves it as a Web Bundle.

const fs = require('fs');
const puppeteer = require('puppeteer');
const wbn = require('wbn');

async function SaveAsWbn(url, outFile) {
  const builder = new wbn.BundleBuilder(url);
  const browser = await puppeteer.launch();
  const page = await browser.newPage();

  page.on('requestfinished', async request => {
    if (request.method() !== 'GET' || !request.url().match(/^https?:/)) {
      return;
    }
    try {
      const response = await request.response();
      let responseBody = '';
      if (request.redirectChain().length === 0) {
        // Redirect responses don't have body.
        responseBody = await response.buffer();
      }
      console.log(request.url());
      builder.addExchange(request.url(),
        response.status(),
        response.headers(),
        responseBody);
    } catch (e) {
      console.log(request.url(), e);
    }
  });

  await page.goto(url, { waitUntil: 'networkidle0' });
  await browser.close();

  outFile = outFile || 'out.wbn';
  fs.writeFileSync(outFile, builder.createBundle());
}

if (process.argv[2]) {
  SaveAsWbn(process.argv[2], process.argv[3]);
} else {
  console.log(`Usage: ${process.argv[0]} ${process.argv[1]} URL`)
}
