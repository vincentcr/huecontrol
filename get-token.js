/*
  Get a remote access token to the Hue api.

  1. get device id. assumes in same lan
  2. go to login page, login to get cookie
  3. post grant url to activate token
  4. cookie should be in go back to app page
*/


import cheerio from 'cheerio';
import _ from 'lodash';
import request from 'request';
import minimist from 'minimist';
import readlineSync from 'readline-sync';

const BASE_URL = 'https://www.meethue.com';
const TOKEN_HREF_RE = /^phhueapp:\/\/sdk\/login\/(.+)$/;


function main() {
  return getCredentialFromCommandLine()
    .then(getToken)
    .then(printResults)
    .catch((err) => {
      console.log('ERR!', err.stack);
    })
  ;
}


function getCredentialFromCommandLine() {
  const argv = minimist(process.argv.slice(2));

  return Promise.resolve({
    email: promptIfNull('email', argv.email),
    password: promptIfNull('password', argv.password, {echo: false}),
  });
}

function promptIfNull(label, val, {echo} = {echo:true}) {
  if (val != null) {
    return val;
  } else {
    return readlineSync.question(`${label}: `, { hideEchoBack: !echo });
  }
}


export default function getToken(creds) {
  return getBridgeID()
    .then(startTokenSession)
    .then(() => grantAccess(creds))
    .then(fetchToken)
  ;
}

function getBridgeID() {
  const bridgeDiscoveryURL = `${BASE_URL}/api/nupnp`;
  return fetch(bridgeDiscoveryURL)
    .then(res => {
      const bridges = JSON.parse(res.body);
      const bridge = bridges[0];
      if (bridge == null || bridge.id == null) {
        throw new Error('bridge not found. Please ensure you are in the same network as your Hue Bridge.');
      } else {
        return bridge.id;
      }
    });
}

function startTokenSession(bridgeID) {
  const tokenSessionURL = `${BASE_URL}/en-us/api/gettoken?devicename=iPhone+5&appid=hueapp&deviceid=${bridgeID}`;
  return fetch(tokenSessionURL);
}

function grantAccess(creds) {
  const grantAccessURL = `${BASE_URL}/en-us/api/getaccesstokengivepermission`;
  const req = {
    method: 'POST',
    headers: { 'content-type': 'application/x-www-form-urlencoded' },
    form: creds,
  };
  return fetch(grantAccessURL, req).then((res) => {
    const $ = cheerio.load(res.body);
    const href = $('[data-role="yes"]').attr('href');
    if (!href) {
      throw new Error('Unable to grant access to bridge. Please verify your username and password');
    }
    return href;
  });
}

function fetchToken(tokenResultURL) {
  return fetch(BASE_URL + tokenResultURL).then((res) => {
    const $ = cheerio.load(res.body);
    const href = $('a.button-primary').attr('href');
    const token = TOKEN_HREF_RE.exec(href)[1];
    return token;
  });
}

function printResults(token) {
  console.log(token);
}

function fetch(url, settings) {
  const settingsWithJar =  _.merge({ jar: true }, settings);
  return new Promise((resolve, reject) => {
    request(url, settingsWithJar, (err, res) => {
      if (err != null) {
        reject(err);
      } else if (res.statusCode >= 400) {
        const httpErr = new Error(`${url}: unexpected status ${res.statusCode}. body: ${res.body}`);
        reject(httpErr);
      } else {
        resolve(res);
      }
    });
  });

}

if (require.main === module) {
  main();
}
