var app = require('express')();
// const { setServerUrl, loadXLH } = require("./ApexLink");
const base64url = require('base64url')
const bodyParser = require("body-parser");
const session = require('express-session');
const util = require('util')
const exec = util.promisify(require('child_process').exec);

/*
提供 Tekpass 身份驗證和 SSO 功能
通過 HTTP 和 WebSocket 與客戶端進行通信
並使用加密和解密來保護數據安全。
*/

// 日誌中添加時間戳。
require('log-timestamp');

// 讀取和寫入文件。
const fs = require('fs');

const crypto = require("crypto");

// 讀取憑證及金鑰
const prikey = fs.readFileSync('privatekey.pem', 'utf8');
const cert = fs.readFileSync('ca.pem', 'utf8');

// 建立憑證及金鑰
const credentials = {
    key: prikey,
    cert: cert
};

const http = require('http');

// const server = http.createServer((req, res) => {
//   res.statusCode = 200;
//   res.setHeader('Content-Type', 'text/plain');
//   res.end('Hello, World!');
// });

// 創建一個 HTTPS 服務器，並將 Express 應用實例作為請求處理器，傳入憑證和金鑰。
var https = require('https').createServer(credentials, app);

// WebSocket 服務器綁定在 HTTPS 服務器上。
// var io = require('socket.io')(https, {
//     cors: {
//         origin: "http://localhost:5000"
//         // origin: "https://localhost:403"
//     }
// });
const WebSocket = require('ws');
const wss = new WebSocket.Server({ server: https });
// const wss = new WebSocket.Server({ https });

let xlh = "2TNsA6a2C6JQLnZl26dNF7RDv3lBUokiUZdRZED_szx2VsVxKUODgT1DOwgTrs1Zr1IVtunk6d8vNqaB5zW-BhNDYK9HZ1THjZSLuRZ0eO-qPSUuLClQS3p7JMLoGVN24QBSrDUmxBM";
let xlhPin = "004309";
let xlhUser = "User-1670460972576";
// let serverUrl = "https://2hwi7j8zb7.execute-api.ap-northeast-1.amazonaws.com/default/apex-v2";

// setServerUrl(serverUrl);

// 處理跨域請求。
const cors = require('cors');

// 獲取本地 IPv4 地址。
const localIpV4Address = require('local-ipv4-address');

app.use(cors());

// 使用 body-parser 中間件來解析請求體中的 JSON 數據，並設置大小限制為 10kb。
app.use(bodyParser.json({ limit: '10kb' }));
app.use((req, res, next) => {
    console.log(`[${new Date()}] Received request: ${req.method} ${req.url}`);
    next();
});

const commonHeaders = () => ({
    // "x-spanish": "nuke2"
});

const textBody = (body, status) => ({
    "isBase64Encoded": false,
    "statusCode": status,
    "headers": commonHeaders(),
    "body": body
});

// var link = null;
var randBytes = null;
var apexid = null;
var localip = null;
var sso_id = null;
var sso_token = null;
var UserData = null;
var status = 0; //紀錄是否有人占用qrcode
var expire = 0; //紀錄是否qrcode過期

// 加密演算法
const algorithm = "aes-256-gcm";

// TODO - 1
// 定義 POST 請求的處理函數 "/cgi-bin/TekpassCheck"，用於驗證 SSO token 和用戶 ID。
app.post("/cgi-bin/TekpassCheck", async (req, res) => {
    console.log("/cgi-bin/TekpassCheck");

    try {
        console.log(req.body)
        // link = link ?? (await loadXLH(xlh, xlhUser, xlhPin));
        sso_token = req.body.sso_token;

        console.log(sso_id.length());
        console.log(sso_token.length());

        const plain = Buffer.from(sso_id, 'base64');
        console.log(plain)

        let qrKey = randBytes.subarray(0, 32);
        let qrIV = randBytes.subarray(32, 48);

        let userId = aesGcmDecrypt(qrKey, plain, qrIV);
        userId = userId.subarray(0, userId.length - 16);
        console.log("userId:" + userId.toString('utf8'));

        if (!(sso_token && userId)) return res.status(400).send(textBody(`${req.body} error ${userId} ${sso_token}`));
        let result = false;
        try {
            result = sso_check(userId, sso_token)
            // result = await link.verifySSOToken(sso_token, userId);
        } catch (er) {
            console.error(er);
        }

        if (result) {
            if (apexid == null) {
                apexid = userId.toString('utf8');
                return res.status(200).send(textBody(""));
            } else {
                if (userId.toString('utf8') != apexid) {
                    return res.status(400).send(textBody(""));
                } else {
                    return res.status(200).send(textBody(""));
                }
            }
        } else {
            return res.status(400).send(textBody(""));
        }
    } catch (err) {
        console.error(err);
        return res.status(500).send(textBody(`${req.body} ${err.toString()}`));
    }
});

// TODO - 2
// 定義 POST 請求的處理函數 "/cgi-bin/TekpassResult"，用於接收 SSO ID 和用戶數據。
app.post("/cgi-bin/TekpassResult", async (req, res) => {
    console.log("/cgi-bin/TekpassResult");
    console.log(req.body);

    status = 0

    sso_id = req.body.sso_id;
    UserData = req.body.sso_token;
    console.log(sso_id.length)
    console.log(UserData.length)

    return res.status(200).send(textBody(""));
});

// TODO - 3 版本號
app.get('/', async (req, res) => {
    return res.status(200).send("version 0.3");
});

// 定義 GET 請求的處理函數 "/tekpass_auth"，用於生成 Tekpass 身份驗證的 URL。
app.get('/tekpass_auth', async (req, res) => {
    console.log('tekpass');
    randBytes = crypto.randomBytes(64);
    // console.log(randBytes.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, ''))
    let url = null;

    if (!status) {
        expire = 0
        status = 1
        setTimeout(() => {
            console.log('expired!!!')
            expire = 1
            status = 0
        }, 300000);
        url = 'https://tekpass.com.tw/sso?receiver=' + localip + ':8000&token=' + randBytes.toString('base64url');
        console.log('url:' + url)
        return res.status(200).json({ url })
    } else {
        return res.status(200).json({ url })
    }
})

// 定義 GET 請求的處理函數 "/tekpass_token"，用於返回 SSO token 和過期標誌。
app.get('/tekpass_token', async (req, res) => {
    console.log('sso_token');

    sso_token = UserData;
    console.log(sso_token)
    UserData = null;

    return res.status(200).json({ sso_token: sso_token, expire })
})

// 定義 WebSocket 服務器的連接和消息處理邏輯。
wss.on('connection', function (socket) {
    console.log('a user connected');

    socket.on('message', function (data) {
        const buffer = Buffer.from(data, 'hex');
        const event = buffer.toString('utf8');
        console.log('接收到訊息：', event);

        const randBytes = crypto.randomBytes(64);
        const response = 'https://tekpass.com.tw/sso?receiver=' + localip + ':8080&token=' + randBytes.toString('base64url');
        socket.key = randBytes.toString('base64url');
        if (event == 'tekpass') {
            data = { event: 'acknowledge', data: response }
            socket.send(JSON.stringify(data));
        }
        socket.key = randBytes.toString('base64url');
    });

    socket.on('close', function () {
        console.log('disconnect');
    });
});




// https.listen(8080, async () => {
//     console.log('listening on *:8080');

//     localip = await localIpV4Address();
//     console.log(localip);
// });
wss.on('listening', function () {
    console.log('WebSocket server is listening');

    localIpV4Address().then((ip) => {
        localip = ip;
        console.log(localip);
    });
});

https.listen(8000);
// server.listen(8080, () => {
//     console.log('HTTPS server is running on port 8080');
// });  


// 定義 AES-GCM 加密和解密函數。
const aesGcmEncrypt = (key, data, iv) => {
    iv = iv || crypto.randomBytes(16);
    let cipher = crypto.createCipheriv(algorithm, key, iv, { authTagLength: 16 });
    return Buffer.concat([cipher.update(data), cipher.final()]);
};
const aesGcmDecrypt = (key, encrypted, iv) => {
    iv = iv || crypto.randomBytes(16);
    let decipher = crypto.createDecipheriv(algorithm, key, iv, { authTagLength: 16 });
    // decipher.setAAD(Buffer.alloc(iv.length)); // not very neccessary.
    // decipher.setAuthTag(Buffer.alloc(iv.length));
    return decipher.update(encrypted);
};

// 定義 sso_check 函數，用於驗證 SSO token。
async function sso_check(user_id, sso_token) {
    const { stdout, stderr } = await exec(`/bin/sso-check-linux-amd64 https://2hwi7j8zb7.execute-api.ap-northeast-1.amazonaws.com/default/apex-v2 Free5GC-1699862217954 m6v8-1hSYujwO2xRne1QYK1EHlwzRu4tfCc0rMepSVxE_ViyVnJPJeDyJ_mwn-DBw-PEKVaK10yEGjiGgCAi1itDhi442v4bQGrn3mRbxJJVbsb4MMTKBSzMhum8u4a-G6mxgZrbMWbiJmr_7AzNatm8_RlGQ5y9 free5gc ${user_id} ${sso_token}`);
    console.log('stdout: ' + stdout)
    console.log(stderr)
    return stdout.result
}

// 定義 refresh 函數，用於生成新的隨機字節並向 WebSocket 客戶端發送刷新消息。
function refresh() {
    randBytes = crypto.randomBytes(64);
    io.in("5gc").emit('refresh', 'https://tekpass.com.tw/sso?receiver=' + localip + ':8000&token=' + randBytes.toString('base64url'));
}

// const WebSocket = require('ws');

// const wss = new WebSocket.Server({ port: 3000 });

// wss.on('connection', (socket) => {
//     console.log('WebSocket connection established.');

//     socket.on('message', (message) => {
//         console.log('Received message from WebSocket client:', message);
//         // 傳送訊息給 WebSocket 客戶端
//         socket.send('Hello from Node.js!');
//     });
// });